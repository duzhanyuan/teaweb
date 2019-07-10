package board

import (
	"github.com/TeaWeb/code/teaconfigs"
	"github.com/TeaWeb/code/teaweb/actions/default/proxy/proxyutils"
	"github.com/iwind/TeaGo/actions"
)

type UseChartAction actions.Action

// 使用某个Chart
func (this *UseChartAction) Run(params struct {
	ServerId string
	WidgetId string
	ChartId  string
	Type     string
}) {
	server := teaconfigs.NewServerConfigFromId(params.ServerId)
	if server == nil {
		this.Fail("找不到Server")
	}

	switch params.Type {
	case "realtime":
		if server.RealtimeBoard == nil {
			server.RealtimeBoard = teaconfigs.NewBoard()
		}
		server.RealtimeBoard.AddChart(params.WidgetId, params.ChartId)
	case "stat":
		if server.StatBoard == nil {
			server.StatBoard = teaconfigs.NewBoard()
		}
		server.StatBoard.AddChart(params.WidgetId, params.ChartId)
	}

	err := server.Save()
	if err != nil {
		this.Fail("保存失败：" + err.Error())
	}

	// 重启统计
	proxyutils.ReloadServerStats(server.Id)

	this.Success()
}
