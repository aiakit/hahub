package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

// 灯控
// 开/关：表示控制全部
// 其他：控制名字相同的灯组
func lightControl(c *ava.Context) {
	entities := core.GetEntityAreaMap()
	if len(entities) == 0 {
		return
	}

	for bName, v := range switchSelectSameName {
		bns := strings.Split(bName, "_")
		if len(bns) != 2 {
			continue
		}

		areaId := bns[0]
		areaName := core.SpiltAreaName(core.GetAreaName(areaId))
		buttonName := bns[1]

		//遍历当前区域所有设备
		//找到所有设备
		es := entities[areaId]
		if len(es) == 0 {
			continue
		}

		auto := &Automation{
			Alias:       areaName + buttonName + "灯开关",
			Description: "对" + areaName + "区域的" + buttonName + "做对应的灯控",
			Mode:        "single",
		}

		//按键触发和条件
		for _, e := range v {
			auto.Triggers = append(auto.Triggers, Triggers{
				EntityID: e.EntityID,
				Trigger:  "state",
			})

			if e.Category == core.CategorySwitchClickOnce {
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
		// 先收集所有匹配的灯组
		for _, l := range es {
			if l.Category == core.CategoryLightGroup {
				if strings.Contains(l.DeviceName, buttonName) || strings.Contains(buttonName, "开/关") {
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

			if l.Category == core.CategoryLight {
				if (strings.Contains(l.DeviceName, buttonName) || strings.Contains(buttonName, "开/关")) && (strings.Contains(l.DeviceName, "彩") || strings.Contains(l.DeviceName, "夜灯")) {
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
