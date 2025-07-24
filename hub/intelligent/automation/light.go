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

	for _, v := range entities {
		for _, e := range v {
			if e.Category == core.CategorySwitchToggle {
				var areaName = core.SpiltAreaName(e.AreaName)

				name := strings.Split(e.OriginalName, " ")
				buttonName := ""

				for _, v1 := range name {
					if v1 != " " && v1 != "" {
						buttonName = v1
						break
					}
				}

				auto := &Automation{
					Alias:       areaName + buttonName + "灯开关",
					Description: "对" + areaName + "区域的" + buttonName + "进行控制",
					Mode:        "single",
				}
				auto.Triggers = append(auto.Triggers, Triggers{
					EntityID: e.EntityID,
					Trigger:  "state",
				})

				for _, l := range v {
					//找到名字相同的灯组
					if l.Category == core.CategoryLightGroup {

						if strings.Contains(l.Name, buttonName) {
							auto.Actions = append(auto.Actions, ActionLight{
								Action: "light.turn_on",
								Data: &actionLightData{
									ColorTempKelvin: 4500,
									BrightnessPct:   100,
								},
								Target: &targetLightData{DeviceId: l.DeviceID},
							})
						}
					}
					if e.Category == core.CategoryLight {
						if strings.Contains(e.Name, "彩") || strings.Contains(e.Name, "夜灯") {
							auto.Actions = append(auto.Actions, ActionLight{
								Action: "light.turn_on",
								Target: &targetLightData{DeviceId: l.DeviceID},
							})
						}
					}
				}

				if len(auto.Actions) > 0 {
					CreateAutomation(c, auto)
				}
			}
		}
	}
}
