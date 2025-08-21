package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 调光旋钮工作,每次亮度减少20%
// 只调当前打开的灯，关闭的灯忽略
func dimmmingIncrease(c *ava.Context) {
	entities, ok := data.GetEntityCategoryMap()[data.CategoryDimming]
	if !ok {
		return
	}

	for _, e := range entities {
		ens, ok := data.GetEntityAreaMap()[e.AreaID]
		if !ok {
			continue
		}

		var script = &Script{
			Alias:       data.SpiltAreaName(data.GetAreaName(e.AreaID)) + "调高灯光亮度",
			Description: "对" + data.SpiltAreaName(data.GetAreaName(e.AreaID)) + "区域开着的灯进行亮度加调节",
			Sequence:    make([]interface{}, 0, 2),
		}

		// 收集所有符合条件的灯具
		var validLights []*data.Entity
		for _, en := range ens {
			if en.Category != data.CategoryLight && en.Category != data.CategoryLightGroup {
				continue
			}

			if en.Category == data.CategoryLight {
				if !strings.Contains(en.DeviceName, "彩") && !strings.Contains(en.DeviceName, "夜") {
					continue
				}
			}
			validLights = append(validLights, en)
		}

		// 如果没有符合条件的灯具，跳过
		if len(validLights) == 0 {
			continue
		}

		// 构建条件：检查是否有任何灯是开着的
		var anyLightOnCondition []interface{}
		for _, en := range validLights {
			anyLightOnCondition = append(anyLightOnCondition, Conditions{
				EntityID:  en.EntityID,
				State:     "on",
				Condition: "state",
			})
		}

		// 构建If-Then-Else逻辑
		var mainAction IfThenELSEAction

		// If 条件：检查是否有灯开着
		mainAction.If = append(mainAction.If, ifCondition{
			Condition:  "or",
			Conditions: anyLightOnCondition,
		})

		// Then 分支：有灯开着，只调节开着的灯的亮度
		for _, en := range validLights {
			var act IfThenELSEAction
			var conditions []interface{}
			conditions = append(conditions, Conditions{
				EntityID:  en.EntityID,
				State:     "on",
				Condition: "state",
			})
			act.If = append(act.If, ifCondition{
				Condition:  "and",
				Conditions: conditions,
			})

			act.Then = append(act.Then, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessStepPct: 20,
				},
				Target: &targetLightData{DeviceId: en.DeviceID},
			})

			mainAction.Then = append(mainAction.Then, act)
		}

		// Else 分支：所有灯都关闭，打开所有灯
		for _, en := range validLights {
			mainAction.Else = append(mainAction.Else, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessStepPct: 10,
				},
				Target: &targetLightData{DeviceId: en.DeviceID},
			})
		}

		script.Sequence = append(script.Sequence, mainAction)

		if len(script.Sequence) > 0 {
			CreateScript(c, script)
		}
	}
}
