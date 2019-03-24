package tealogs

import (
	"context"
	"github.com/TeaWeb/code/teamongo"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/timers"
	"github.com/iwind/TeaGo/utils/time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"runtime"
	"sync"
	"time"
)

var (
	accessLogger = NewAccessLogger()
)

// 访问日志记录器
type AccessLogger struct {
	queue chan *AccessLogItem

	logs      []*AccessLogItem
	timestamp int64
	qps       int

	collectionCacheMap    map[string]*teamongo.Collection
	collectionCacheLocker sync.Mutex
}

type AccessLogItem struct {
	log *AccessLog
}

func NewAccessLogger() *AccessLogger {
	logger := &AccessLogger{
		queue:              make(chan *AccessLogItem, 10240),
		collectionCacheMap: map[string]*teamongo.Collection{},
	}

	go logger.wait()
	return logger
}

func SharedLogger() *AccessLogger {
	return accessLogger
}

func (this *AccessLogger) Push(log *AccessLog) {
	this.queue <- &AccessLogItem{
		log: log,
	}
}

func (this *AccessLogger) client() *mongo.Client {
	return teamongo.SharedClient()
}

func (this *AccessLogger) collection() *teamongo.Collection {
	this.collectionCacheLocker.Lock()
	defer this.collectionCacheLocker.Unlock()

	collName := "logs." + timeutil.Format("Ymd")
	coll, found := this.collectionCacheMap[collName]
	if found {
		return coll
	}

	// 构建索引
	coll = teamongo.FindCollection(collName)
	coll.CreateIndex(map[string]bool{
		"serverId": true,
	})
	coll.CreateIndex(map[string]bool{
		"status":   true,
		"serverId": true,
	})
	coll.CreateIndex(map[string]bool{
		"remoteAddr": true,
		"serverId":   true,
	})
	coll.CreateIndex(map[string]bool{
		"hasErrors": true,
		"serverId":  true,
	})

	this.collectionCacheMap[collName] = coll

	return coll
}

func (this *AccessLogger) wait() {
	timestamp := time.Now().Unix()

	var docs = []interface{}{}
	var docsLocker = sync.Mutex{}

	// 写入到数据库
	countCPU := runtime.NumCPU()
	timers.Loop(500*time.Millisecond, func(looper *timers.Looper) {
		// 写入到本地数据库
		if this.client() != nil {
			docsLocker.Lock()
			if len(docs) == 0 {
				docsLocker.Unlock()
				return
			}
			newDocs := docs
			docs = []interface{}{}
			docsLocker.Unlock()

			total := len(newDocs)

			// 批量写入数据库
			// 需合理控制此数值的大小，避免CPU占用太高
			bulkSize := countCPU * 64
			offset := 0
			for {
				end := offset + bulkSize
				if end > total {
					end = total
				}

				// logs.Println("dump", end-offset, "access logs ...")
				docSlice := newDocs[offset:end]

				// 分析
				for _, doc := range docSlice {
					accessLog := doc.(*AccessLog)
					accessLog.Parse()
					accessLog.Id = primitive.NewObjectID()

					// 执行处理器
					CallAccessLogHooks(accessLog)
				}

				// 写入数据库
				_, err := this.collection().InsertMany(context.Background(), docSlice)
				if err != nil {
					logs.Error(err)
					return
				}
				//logs.Println("done")
				docSlice = []interface{}{}

				offset = end
				if end >= total {
					break
				}
			}

			// 清空
			newDocs = []interface{}{}
		}
	})

	// 接收日志
	for {
		item := <-this.queue
		log := item.log

		// 计算QPS
		this.timestamp = log.Timestamp
		if log.Timestamp-timestamp <= 1 {
			this.qps ++
		} else {
			this.qps = 1
			timestamp = log.Timestamp
		}

		docsLocker.Lock()
		docs = append(docs, log)
		docsLocker.Unlock()
	}
}

// 关闭
func (this *AccessLogger) Close() {
	if this.client() != nil {
		this.client().Disconnect(context.Background())
	}
}

func (this *AccessLogger) QPS() int {
	if time.Now().Unix()-this.timestamp < 2 {
		return this.qps
	}
	return 0
}
