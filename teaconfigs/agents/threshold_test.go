package agents

import (
	"github.com/iwind/TeaGo/maps"
	"testing"
	"time"
)

func TestThreshold_Test(t *testing.T) {
	threshold := NewThreshold()
	threshold.Param = "${0}"
	threshold.Operator = ThresholdOperatorGt
	threshold.Value = "12"
	threshold.Validate()
	t.Log(threshold.Test("123", nil))

	threshold.Param = "${1}"
	threshold.Operator = ThresholdOperatorGt
	threshold.Validate()
	t.Log(threshold.Test([]interface{}{1, 200, 3}, nil))

	threshold.Param = "${host}"
	threshold.Operator = ThresholdOperatorPrefix
	threshold.Value = "127."
	threshold.Validate()
	t.Log(threshold.Test(map[string]interface{}{
		"host": "127.0.0.1",
	}, nil))

	threshold.Param = "${data.version}"
	threshold.Operator = ThresholdOperatorEq
	threshold.Value = "1.0.25"
	threshold.Validate()
	t.Log(threshold.Test(map[string]interface{}{
		"data": maps.Map{
			"version": "1.0.25",
		},
	}, nil))

	threshold.Param = "${data.hello.world.0}"
	threshold.Operator = ThresholdOperatorEq
	threshold.Value = "1"
	t.Log(threshold.Test(map[string]interface{}{
		"data": maps.Map{
			"version": "1.0.25",
			"hello": maps.Map{
				"world": []string{"1", "2", "3", "4", "5"},
			},
		},
	}, nil))
}

func TestThreshold_Test2(t *testing.T) {
	threshold := NewThreshold()
	threshold.Param = "${changes}"
	threshold.Operator = ThresholdOperatorEq
	threshold.Value = "true"
	err := threshold.Validate()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(threshold.Test(maps.Map{
		"changes": true,
	}, nil))
}

func TestThreshold_Eval(t *testing.T) {
	threshold := NewThreshold()
	threshold.Param = "${data.hello.world.0} * 100 / ${data.hello.world.1}"
	t.Log(threshold.Eval(map[string]interface{}{
		"data": maps.Map{
			"version": "1.0.25",
			"hello": maps.Map{
				"world": []string{"1", "2", "3", "4", "5"},
			},
		},
	}, nil))
}

func TestThreshold_Eval_Date(t *testing.T) {
	threshold := NewThreshold()
	threshold.Param = "new Date().getTime() / 1000 - ${timestamp}"
	t.Log(threshold.Eval(map[string]interface{}{
		"timestamp": time.Now().Unix() - 10,
	}, nil))
}

func TestThreshold_Old(t *testing.T) {
	threshold := NewThreshold()
	threshold.Param = "${rows} - ${OLD.rows234}"
	t.Log(threshold.Eval(map[string]interface{}{
		"rows": 1,
	}, map[string]interface{}{
		"rows234": 123,
	}, ))
}

func TestThreshold_RunActions(t *testing.T) {
	threshold := NewThreshold()
	threshold.Actions = []map[string]interface{}{
		{
			"code": "script",
			"options": map[string]interface{}{
				"scriptType": "path",
				"path":       "1",
			},
		},
	}
	t.Log(threshold.RunActions(nil))
}
