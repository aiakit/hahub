package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 灯控
// 开/关：表示控制全部
// 其他：控制名字相同的灯组
func LightControl(c *ava.Context) {
	entities := data.GetEntityAreaMap()
	if len(entities) == 0 {
		return
	}

	for bName, v := range switchSelectSameName {
		bns := strings.Split(bName, "_")
		if len(bns) < 2 {
			continue
		}

		areaId := strings.Join(bns[:len(bns)-1], "_")
		areaName := data.SpiltAreaName(data.GetAreaName(areaId))
		buttonName := bns[len(bns)-1]

		//遍历当前区域所有设备
		//找到所有设备
		es := entities[areaId]
		if len(es) == 0 {
			continue
		}

		auto := &Automation{
			Alias:       areaName + "按键" + buttonName + "控制灯",
			Description: "对" + areaName + "区域的" + buttonName + "做对应的灯控",
			Mode:        "single",
		}

		//按键触发和条件
		for _, e := range v {
			auto.Triggers = append(auto.Triggers, Triggers{
				EntityID: e.EntityID,
				Trigger:  "state",
			})

			if e.Category == data.CategorySwitchClickOnce {
				auto.Conditions = append(auto.Conditions, Conditions{
					Condition: "state",
					EntityID:  e.EntityID,
					Attribute: e.Attribute,
					State:     e.SeqButton,
				})
			}
		}

		// 收集所有匹配的灯组实体ID

		var conditions []interface{}
		var actionsOn []interface{}
		var actionsOff []interface{}
		var xinguangLights []*data.Entity

		// 先收集所有匹配的灯组
		for _, e := range es {
			if e.Category == data.CategoryXinGuang {
				if !strings.Contains(e.DeviceName, "主机") && strings.HasPrefix(e.EntityID, "light.") && (strings.Contains(e.DeviceName, buttonName) || strings.Contains(buttonName, "开/关") || strings.Contains(buttonName, "开关")) {
					xinguangLights = append(xinguangLights, e)
				}
			}
			if e.Category == data.CategoryLightGroup {
				if strings.Contains(e.DeviceName, buttonName) || strings.Contains(buttonName, "开/关") || strings.Contains(buttonName, "开关") {
					conditions = append(conditions, Conditions{
						EntityID:  e.EntityID,
						State:     "on",
						Condition: "state",
					})

					actionsOn = append(actionsOn, ActionLight{
						Action: "light.turn_on",
						Data: &actionLightData{
							ColorTempKelvin: 4500,
							BrightnessPct:   100,
						},
						Target: &targetLightData{DeviceId: e.DeviceID},
					})
					actionsOff = append(actionsOff, ActionLight{
						Action: "light.turn_off",
						Target: &targetLightData{DeviceId: e.DeviceID},
					})
				}
			}

			if e.Category == data.CategoryLight {
				if (strings.Contains(e.DeviceName, buttonName) || strings.Contains(buttonName, "开/关") || strings.Contains(buttonName, "开关")) && (strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜灯")) {
					conditions = append(conditions, Conditions{
						EntityID:  e.EntityID,
						State:     "on",
						Condition: "state",
					})
					actionsOn = append(actionsOn, ActionLight{
						Action: "light.turn_on",
						Data: &actionLightData{
							ColorTempKelvin: 4500,
							BrightnessPct:   100,
						},
						Target: &targetLightData{DeviceId: e.DeviceID},
					})
					actionsOff = append(actionsOff, ActionLight{
						Action: "light.turn_off",
						Target: &targetLightData{DeviceId: e.DeviceID},
					})
				}
			}
		}

		// 重构馨光灯的控制逻辑，使用conditions, actionsOn, actionsOff方式
		if len(xinguangLights) > 0 {
			// 为馨光灯添加条件检查
			for _, l := range xinguangLights {
				if data.GetXinGuang(l.DeviceID) != "" {
					conditions = append(conditions, Conditions{
						EntityID:  l.EntityID,
						State:     "on",
						Condition: "state",
					})
				}
			}

			// 构建开启馨光灯的动作
			for _, l := range xinguangLights {
				if data.GetXinGuang(l.DeviceID) == "" {
					continue
				}

				// 改为静态模式,不能并行执行，必须优先执行
				actionsOn = append(actionsOn, &ActionLight{
					DeviceID: l.DeviceID,
					Domain:   "select",
					EntityID: data.GetXinGuang(l.DeviceID),
					Type:     "select_option",
					Option:   "静态模式",
				})

				// 修改颜色
				actionsOn = append(actionsOn, &ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct: 100,
						RgbColor:      GetRgbColor(4000),
					},
					Target: &targetLightData{DeviceId: l.DeviceID},
				})

				// 关闭馨光灯的动作
				actionsOff = append(actionsOff, &ActionLight{
					Action: "light.turn_off",
					Target: &targetLightData{DeviceId: l.DeviceID},
				})
			}
		}

		if len(conditions) == 0 || len(actionsOn) == 0 {
			for _, l := range es {
				//如果没有就开启单灯
				if l.Category == data.CategoryLight {
					if strings.Contains(l.DeviceName, buttonName) || strings.Contains(buttonName, "开/关") || strings.Contains(buttonName, "开关") {
						conditions = append(conditions, Conditions{
							EntityID:  l.EntityID,
							State:     "on",
							Condition: "state",
						})
						actionsOn = append(actionsOn, ActionLight{
							Action: "light.turn_on",
							Data: &actionLightData{
								ColorTempKelvin: 4500,
								BrightnessPct:   100,
							},
							Target: &targetLightData{DeviceId: l.DeviceID},
						})
						actionsOff = append(actionsOff, ActionLight{
							Action: "light.turn_off",
							Target: &targetLightData{DeviceId: l.DeviceID},
						})
					}
				}
			}

			if len(conditions) == 0 || len(actionsOn) == 0 {
				continue
			}
		}

		var act IfThenELSEAction
		act.If = append(act.If, ifCondition{
			Condition:  "and",
			Conditions: conditions,
		})
		act.Then = actionsOff
		act.Else = actionsOn

		auto.Actions = append(auto.Actions, act)

		CreateAutomation(c, auto)
	}
}
