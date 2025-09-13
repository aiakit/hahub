package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func InitHoming(c *ava.Context) {
	// 创建回家场景
	script, auto := homingScript()
	if script != nil && len(script.Sequence) > 0 {
		AddScript2Queue(c, script)
	}

	if auto != nil && len(auto.Actions) > 0 && len(auto.Triggers) > 0 {
		AddAutomation2Queue(c, auto)
	}
}

// 回家场景场景
func homingScript() (*Script, *Automation) {
	var script = &Script{
		Alias:       "回家场景",
		Description: "回家场景执行场景",
	}

	var action = make([]interface{}, 0)

	// 撤防
	action = append(action, ActionService{
		Action: "automation.turn_off",
		Data:   map[string]interface{}{"stop_actions": true},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "automation.li_jia_bu_fang"},
	})

	var xiaomiHomeSpeakerDeviceId string
	func() {
		entities, _ := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		for _, e := range entities {
			if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") && strings.HasPrefix(e.EntityID, "text.") {
				action = append(action, ActionNotify{
					Action: "notify.send_message",
					Data: struct {
						Message string `json:"message,omitempty"`
						Title   string `json:"title,omitempty"`
					}{Message: "[播放一段轻音乐,true]"},
					Target: &struct {
						DeviceID string `json:"device_id,omitempty"`
					}{DeviceID: e.DeviceID},
				})
				xiaomiHomeSpeakerDeviceId = e.DeviceID
				break
			}
		}
	}()

	var areaId string

	//打开客厅所有灯
	func() {
		//判断是否有展示脚本,如果有，使用展示脚本
		if displayEntityId != "" {
			action = append(action, ActionService{
				Action: "script.turn_on",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: displayEntityId},
			})
		}

		if displayEntityId == "" {
			entities, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
			if ok {
				//先开氛围灯
				for _, v := range entities {
					if strings.Contains(v.AreaName, "客厅") {
						areaId = v.AreaID
						if strings.Contains(v.DeviceName, "氛围") {
							action = append(action, ActionLight{
								Action: "light.turn_on",
								Data: &actionLightData{
									ColorTempKelvin: 5800,
									BrightnessPct:   100,
								},
								Target: &targetLightData{DeviceId: v.DeviceID},
							})
							action = append(action, ActionLight{
								Delay: &delay{
									Hours:        0,
									Minutes:      0,
									Seconds:      5,
									Milliseconds: 0,
								},
							})
						}

					}
				}

				for _, v := range entities {
					if strings.Contains(v.AreaName, "客厅") {
						if !strings.Contains(v.DeviceName, "氛围") {
							action = append(action, ActionLight{
								Action: "light.turn_on",
								Data: &actionLightData{
									ColorTempKelvin: 5800,
									BrightnessPct:   100,
								},
								Target: &targetLightData{DeviceId: v.DeviceID},
							})
						}

					}
				}
			}
		}
	}()

	//打开插座
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategorySocket]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.AreaName, "客厅") {
					action = append(action, ActionCommon{
						Type:     "turn_on",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "switch",
					})
				}
			}
		}
	}()

	//打开窗帘
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryCurtain]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.AreaName, "客厅") {
					action = append(action, ActionCommon{
						Type:     "open",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "cover",
					})
				}
			}
		}
	}()

	func() {
		result := TurnOnTv(areaId)
		if len(result) > 0 {
			action = append(action, result...)
		}
	}()

	func() {

		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			action = append(action, ActionTimerDelay{
				Delay: &delay{
					Hours:        0,
					Minutes:      0,
					Seconds:      30,
					Milliseconds: 0,
				},
			})
			for _, e := range entities {
				if e.DeviceID == xiaomiHomeSpeakerDeviceId {
					if strings.HasPrefix(e.EntityID, "media_player.") && strings.Contains(e.AreaName, "客厅") {
						action = append(action, ActionService{
							Action: "media_player.media_pause",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: e.EntityID}})
					}
				}
			}
		}
	}()

	//var act IfThenELSEAction

	//检查客厅存在传感器是否检测到有人
	//暂时不用传感器作为条件
	//func() {
	//	entities, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
	//	if ok {
	//		for _, e := range entities {
	//			if strings.Contains(e.AreaName, "客厅") {
	//				act.If = append(act.If, ifCondition{
	//					Type:      "is_not_occupied",
	//					DeviceId:  e.DeviceID,
	//					EntityId:  e.EntityID,
	//					Condition: "device",
	//					Domain:    "binary_sensor",
	//				})
	//			}
	//		}
	//	}
	//}()

	//判断客厅灯是否亮
	//if len(act.If) == 0 {
	//entities, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
	//if ok {
	//	for _, e := range entities {
	//		if strings.Contains(e.AreaName, "客厅") {
	//			act.If = append(act.If, ifCondition{
	//				Condition: "device",
	//				Type:      "is_on",
	//				DeviceId:  e.DeviceID,
	//				EntityId:  e.EntityID,
	//				For: &For{
	//					Hours:   0,
	//					Minutes: 5,
	//					Seconds: 0,
	//				},
	//			})
	//		}
	//	}
	//}
	//}

	//if len(act.If) > 0 && len(action) > 0 {
	//	act.Then = append(act.Then, action...)
	//	script.Sequence = append(script.Sequence, act)
	//}
	if len(action) > 0 {
		script.Sequence = append(script.Sequence, &ActionService{
			Action: "automation.turn_off",
			Target: &struct {
				EntityId string `json:"entity_id"`
			}{EntityId: "automation.hui_jia_zi_dong_hua"},
		})
		script.Sequence = append(script.Sequence, action...)
	}

	if len(action) > 0 {
		auto := homingAutomation(action)
		if auto != nil {
			auto.Actions = action
		}
		return script, auto
	}

	return nil, nil
}

// 回家自动化
func homingAutomation(action []interface{}) *Automation {
	var automation = &Automation{
		Alias:       "回家自动化",
		Description: "门锁打开/或者开关按键触发用或条件，判断客厅所有灯是否打开，存在传感器是否检车到人",
		Mode:        "single",
	}

	var condition = make([]*Conditions, 0)

	//条件：名字中带有"回家"的开关按键和场景按键
	func() {
		for bName, v := range switchSelectSameName {
			// 使用SplitN分割，确保只分割成两部分，保留最后一个_后的字符作为buttonName
			bns := strings.Split(bName, "_")
			if len(bns) < 2 {
				continue
			}
			buttonName := bns[len(bns)-1]

			if strings.Contains(buttonName, "回家") {
				//按键触发和条件
				var con = &Conditions{
					Condition: "or",
				}

				for _, e := range v {
					automation.Triggers = append(automation.Triggers, &Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})

					if e.Category == data.CategorySwitchClickOnce && e.SeqButton > 0 {
						con.ConditionChild = append(con.ConditionChild, &Conditions{
							Condition: "state",
							EntityID:  e.EntityID,
							Attribute: e.Attribute,
							State:     e.SeqButton,
						})
					}
				}

				if len(con.ConditionChild) > 0 {
					automation.Conditions = append(automation.Conditions, con)
				}
			}
		}
	}()

	trigger := virtualEventAction("回家自动化")
	if trigger != nil {
		automation.Triggers = append(automation.Triggers, trigger)
	}

	if len(automation.Triggers) > 0 {
		automation.Actions = append(automation.Actions, action...)
		var con = &Conditions{
			Condition: "or",
		}

		for _, v := range condition {
			con.ConditionChild = append(con.ConditionChild, v)
		}

		automation.Conditions = append(automation.Conditions, con)
	}

	return automation
}
