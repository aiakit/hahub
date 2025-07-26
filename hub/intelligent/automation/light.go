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

	// 获取开关和按键实体映射
	entitiesSenor, ok1 := core.GetEntityCategoryMap()[core.CategorySwitchSenorSingle]
	entitiesSwitch, ok2 := core.GetEntityCategoryMap()[core.CategorySwitch]

	if !ok1 || !ok2 {
		return
	}

	// 修改：优先遍历按键实体，而不是所有设备
	for _, e := range entitiesSenor {
		// 根据按键的DeviceID找到对应的开关实体数组
		var switchEntities []*core.Entity
		for _, sw := range entitiesSwitch {
			if sw.DeviceID == e.DeviceID {
				switchEntities = append(switchEntities, sw)
			}
		}

		// 如果找不到对应的开关，跳过
		if len(switchEntities) == 0 {
			continue
		}

		// 遍历所有匹配的开关实体
		for _, switchEntity := range switchEntities {
			// 获取开关所在区域的实体列表
			v, exists := entities[switchEntity.AreaID]
			if !exists {
				continue
			}

			var areaName = core.SpiltAreaName(switchEntity.AreaName)
			var entityId = e.EntityID
			var buttonName string

			name := strings.Split(switchEntity.OriginalName, " ")

			for _, v1 := range name {
				if v1 != " " && v1 != "" {
					buttonName = v1
					break
				}
			}

			var seqButton interface{}
			var Attribute = "按键类型"
			switch {
			case strings.Contains(switchEntity.OriginalName, "按键1"):
				seqButton = 1
			case strings.Contains(switchEntity.OriginalName, "按键2"):
				seqButton = 2
			case strings.Contains(switchEntity.OriginalName, "按键3"):
				seqButton = 3
			case strings.Contains(switchEntity.OriginalName, "按键4"):
				seqButton = 4
			default:
				seqButton = nil
			}

			if seqButton == nil {
				Attribute = ""
			}

			auto := &Automation{
				Alias:       areaName + buttonName + "灯开关",
				Description: "对" + areaName + "区域的" + "单击按键" + buttonName + "做对应的灯控",
				Mode:        "single",
			}

			auto.Triggers = append(auto.Triggers, Triggers{
				EntityID: entityId,
				Trigger:  "state",
			})

			auto.Conditions = append(auto.Conditions, Conditions{
				Condition: "state",
				EntityID:  entityId,
				Attribute: Attribute,
				State:     seqButton,
			})

			// 收集所有匹配的灯组实体ID

			var conditions []interface{}
			var actionsOn []interface{}
			var actionsOff []interface{}

			// 先收集所有匹配的灯组
			for _, l := range v {
				if l.Category == core.CategoryLightGroup {
					if strings.Contains(l.DeviceName, buttonName) || strings.Contains(switchEntity.OriginalName, "开/关") {
						conditions = append(conditions, Conditions{
							EntityID:  l.EntityID,
							State:     "on",
							Condition: "state",
						})

						actionsOn = append(actionsOn, ActionLight{
							Action: "light.turn_on",
							Target: &targetLightData{DeviceId: l.DeviceID},
						})
						actionsOff = append(actionsOff, ActionLight{
							Action: "light.turn_off",
							Target: &targetLightData{DeviceId: l.DeviceID},
						})
					}
				}

				if l.Category == core.CategoryLight {
					if (strings.Contains(l.DeviceName, buttonName) || strings.Contains(switchEntity.OriginalName, "开/关")) && (strings.Contains(l.DeviceName, "彩") || strings.Contains(l.DeviceName, "夜灯")) {
						conditions = append(conditions, Conditions{
							EntityID:  l.EntityID,
							State:     "on",
							Condition: "state",
						})
						actionsOn = append(actionsOn, ActionLight{
							Action: "light.turn_on",
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
}
