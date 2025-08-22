package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 调光旋钮工作,每次亮度减少20%
// 只调当前打开的灯，关闭的灯忽略
func dimmmingReduce(c *ava.Context) {

	for areaId, ens := range data.GetEntityAreaMap() {

		var script = &Script{
			Alias:       data.SpiltAreaName(data.GetAreaName(areaId)) + "调低灯光亮度",
			Description: "对" + data.SpiltAreaName(data.GetAreaName(areaId)) + "区域开着的灯进行亮度加调节",
			Sequence:    make([]interface{}, 0, 2),
		}

		for _, en := range ens {
			if en.Category != data.CategoryLight && en.Category != data.CategoryLightGroup {
				continue
			}

			if en.Category == data.CategoryLight {
				if !strings.Contains(en.DeviceName, "彩") && !strings.Contains(en.DeviceName, "夜") {
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
