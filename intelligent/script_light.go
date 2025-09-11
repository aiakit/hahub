package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func LightScriptSetting(c *ava.Context) {
	lightScene(c, "会客", 80, 6000)
	lightScene(c, "观影", 20, 3800)
	lightScene(c, "娱乐", 100, 4800)
	lightScene(c, "休息", 30, 3000)
	lightScene(c, "阅读", 70, 4000)
	lightScene(c, "就餐", 70, 3500)
	lightScene(c, "日光", 100, 5500)
	lightScene(c, "月光", 20, 2700)
	lightScene(c, "黄昏", 100, 2700)
	lightScene(c, "温馨", 50, 3000)
	lightScene(c, "冬天", 80, 5000)
	lightScene(c, "夏天", 80, 3500)
	lightScene(c, "节能", 80, 3800)
}

func lightScene(c *ava.Context, simpleName string, brightness float64, kelvin int) {

	var entities = data.GetEntityAreaMap()

	if len(entities) == 0 {
		return
	}

	for areaId, v := range entities {

		areaName := data.SpiltAreaName(data.GetAreaName(areaId))
		if strings.Contains(simpleName, "就餐") && (!strings.Contains(areaName, "厨房") && !strings.Contains(areaName, "餐厅")) {
			continue
		}

		if (strings.Contains(simpleName, "会客") || strings.Contains(simpleName, "娱乐")) && strings.Contains(areaName, "卧室") {
			continue
		}

		if (strings.Contains(simpleName, "观影") || strings.Contains(simpleName, "阅读")) && (!strings.Contains(areaName, "客厅") && !strings.Contains(areaName, "卧室")) {
			continue
		}

		//// 检查当前区域是否有灯组，如果没有则不创建场景
		//hasLightGroup := false
		//for _, e := range v {
		//	if e.Category == data.CategoryLightGroup {
		//		hasLightGroup = true
		//		break
		//	}
		//}
		//
		//// 如果没有灯组，则不创建任何场景
		//if !hasLightGroup {
		//	continue
		//}

		// 为每个区域创建独立的en和meta map
		var script = &Script{}
		var automation = &Automation{
			Triggers:   make([]*Triggers, 0),
			Conditions: make([]*Conditions, 0),
			Actions:    make([]interface{}, 0),
		}

		func() {
			for bName, v1 := range switchSelectSameName {
				bns := strings.Split(bName, "_")
				if len(bns) < 2 {
					continue
				}
				buttonName := bns[len(bns)-1]
				if strings.Contains(buttonName, simpleName) {
					//按键触发和条件
					var con = &Conditions{
						Condition: "or",
					}

					for _, e := range v1 {

						if e.AreaID != areaId {
							continue
						}

						automation.Triggers = append(automation.Triggers, &Triggers{
							EntityID: e.EntityID,
							Trigger:  "state",
						})

						if e.Category == data.CategorySwitchClickOnce && e.SeqButton > 0 {
							con.ConditionChild = append(con.ConditionChild, &Conditions{
								Condition: "state",
								EntityID:  e.EntityID,
								Attribute: e.Attribute,
								State:     e.SeqButton,
							})
						}
					}

					if len(con.ConditionChild) > 0 {
						automation.Conditions = append(automation.Conditions, con)
					}
				}
			}
		}()

		var actions = make([]interface{}, 0, 2)

		entityFilter := findLightsWithOutLightCategory("", v)
		if len(entityFilter) == 0 {
			continue
		}

		action := turnOnLights(entityFilter, brightness, kelvin, false)
		if len(action) == 0 {
			continue
		}

		for _, e := range action {
			if strings.Contains(simpleName, "节能") {
				if e.Target == nil {
					continue
				}
				e1, ok := data.GetEntityByEntityId()[e.Target.EntityId]
				if !ok {
					continue
				}
				if e1.Category != data.CategoryLightGroup && e1.Category != data.CategoryLight {
					continue
				}

				if !strings.Contains(e1.DeviceName, "氛围") && !strings.Contains(e1.DeviceName, "夜") {
					continue
				}
			}
			actions = append(actions, e)
		}

		if len(actions) > 0 {
			script.Sequence = append(script.Sequence, actions...)
		}

		if len(script.Sequence) > 0 {
			script.Alias = areaName + simpleName + "场景"
			script.Description = "点击开关按键执行" + areaName + simpleName + "场景"
			scriptId := AddScript2Queue(c, script)

			if scriptId != "" && len(automation.Triggers) > 0 {
				automation.Alias = areaName + simpleName + "自动化"
				automation.Description = "点击开关按键执行" + areaName + simpleName + "自动化"
				automation.Mode = "single"

				automation.Actions = append(automation.Actions, ActionService{
					Action: "script.turn_on",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: scriptId},
				})
				// 确保automation对象有效再添加到队列
				if automation.Alias != "" && len(automation.Actions) > 0 {
					AddAutomation2Queue(c, automation)
				}
			}
		}
	}
}
