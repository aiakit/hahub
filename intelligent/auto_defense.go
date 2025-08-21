package intelligent

import (
	"hahub/data"

	"github.com/aiakit/ava"
)

// 布防
// 存在传感器，人体传感器，发送通知
func Defense(c *ava.Context) {
	var automation = &Automation{
		Alias:       "离家布防",
		Description: "离家之后如果存在传感器，人体传感器感应到人，发送通知",
		Mode:        "single",
	}

	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
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
		entities, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entities {
				automation.Triggers = append(automation.Triggers, Triggers{
					EntityID: e.EntityID,
					Trigger:  "state",
				})
			}
		}
	}()

	doNotify("家里有人", "请注意，家里有人！！！", automation)

	if len(automation.Triggers) > 0 {
		entitryId := CreateAutomation(c, automation)
		s := &Script{
			Alias:       "撤防",
			Description: "撤去家里的防御机制",
			Sequence:    make([]interface{}, 0),
		}

		s.Sequence = append(s.Sequence, &ActionService{
			Action: "automation.turn_off",
			Data:   map[string]interface{}{"stop_actions": true},
			Target: &struct {
				EntityId string `json:"entity_id"`
			}{EntityId: entitryId},
		})
		CreateScript(c, s)
	}
}
