package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 灯控
// 开关：表示控制全部
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
			auto.Triggers = append(auto.Triggers, &Triggers{
				EntityID: e.EntityID,
				Trigger:  "state",
			})

			if e.Category == data.CategorySwitchClickOnce {
				auto.Conditions = append(auto.Conditions, &Conditions{
					Condition: "state",
					EntityID:  e.EntityID,
					Attribute: e.Attribute,
					State:     e.SeqButton,
				})
			}
		}

		var prefix = buttonName
		if buttonName == "开关" {
			prefix = ""
		}

		entitiesFilter := findLightsWithOutLightCategory(prefix, es)
		if len(entitiesFilter) == 0 {
			continue
		}
		actionsTurnOn := turnOnLights(entitiesFilter, 100, 4800, true)

		if len(actionsTurnOn) == 0 {
			continue
		}

		conditionTurnOn, actionsOn := spiltCondition(nil, actionsTurnOn)

		actionsOff := turnOffLights(entitiesFilter)
		if len(actionsOn) == 0 || len(actionsOff) == 0 {
			continue
		}

		// 收集所有匹配的灯组实体ID
		var conditions []interface{}
		var offs []interface{}
		for _, e := range actionsOff {
			offs = append(offs, e)
		}

		for _, e := range conditionTurnOn {
			conditions = append(conditions, e)
		}

		if len(conditions) == 0 || len(conditions) == 0 || len(offs) == 0 {
			continue
		}

		var act IfThenELSEAction
		act.If = append(act.If, ifCondition{
			Condition:  "and",
			Conditions: conditions,
		})
		act.Then = actionsOn
		act.Else = offs

		auto.Actions = append(auto.Actions, act)

		AddAutomation2Queue(c, auto)
	}
}
