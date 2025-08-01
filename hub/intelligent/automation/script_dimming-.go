package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

// 调光旋钮工作,每次亮度减少20%
// 只调当前打开的灯，关闭的灯忽略
func dimmmingReduce(c *ava.Context) {
	entities, ok := core.GetEntityCategoryMap()[core.CategoryDimming]
	if !ok {
		return
	}

	for _, e := range entities {
		ens, ok := core.GetEntityAreaMap()[e.AreaID]
		if !ok {
			continue
		}

		var script = &Script{
			Alias:       core.SpiltAreaName(core.GetAreaName(e.AreaID)) + "调低灯光亮度",
			Description: "对" + core.SpiltAreaName(core.GetAreaName(e.AreaID)) + "区域开着的灯进行亮度加调节",
			Sequence:    make([]interface{}, 0, 2),
		}

		for _, en := range ens {
			if en.Category != core.CategoryLight && en.Category != core.CategoryLightGroup {
				continue
			}

			if en.Category == core.CategoryLight {
				if !strings.Contains(en.DeviceName, "彩") && !strings.Contains(en.DeviceName, "夜灯") {
					continue
				}
			}

			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessStepPct: -10,
				},
				Target: &targetLightData{DeviceId: en.DeviceID},
			})
		}

		if len(script.Sequence) > 0 {
			CreateScript(c, script)
		}
	}
}
