package automation

import (
	"hahub/hub/core"

	"github.com/aiakit/ava"
)

// 布防
// 存在传感器，人体传感器，发送通知
func defense(c *ava.Context) {
	var automation = &Automation{
		Alias:       "离家布防",
		Description: "离家之后如果存在传感器，人体传感器感应到人，发送通知",
		Mode:        "single",
	}

	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entities {
				automation.Triggers = append(automation.Triggers, Triggers{
					Type:     "occupied",
					EntityID: e.EntityID,
					DeviceID: e.DeviceID,
					Domain:   "binary_sensor",
					Trigger:  "device",
					For: &For{
						Hours:   0,
						Minutes: 0,
						Seconds: 3,
					},
				})
			}
		}
	}()

	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entities {
				automation.Triggers = append(automation.Triggers, Triggers{
					EntityID: e.EntityID,
					Trigger:  "state",
				})
			}
		}
	}()

	doNotify("离家布防", "请注意，家里有人！！！", automation)

	if len(automation.Triggers) > 0 {
		CreateAutomation(c, automation)
	}
}
