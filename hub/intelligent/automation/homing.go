package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

func initHoming(c *ava.Context) {
	a := homing()
	if a != nil && len(a.Actions) > 0 {
		CreateAutomation(c, a)
	}
}

// 回家场景:门需要出发按键
// 1.门锁打开/或者开关按键触发用或条件，判断客厅所有灯是否打开，如果没有打开，则打开,优先打开门所在位置的灯，再打开客厅氛围灯和其他灯，播放欢迎，播放音乐，打开电视，判断问题打开空调，询问是否需要洗澡。撤防：关闭人体传感器，存在传感器检测到人的报警
func homing() *Automation {
	var automation = &Automation{
		Alias:       "回家自动化",
		Description: "门锁打开/或者开关按键触发用或条件，判断客厅所有灯是否打开，存在传感器是否检车到人",
		Mode:        "single",
	}

	//检查全屋灯是否打开
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
		if ok {
			for _, e := range entities {
				automation.Conditions = append(automation.Conditions, Conditions{
					Condition: "device",
					Type:      "is_off",
					DeviceID:  e.DeviceID,
					EntityID:  e.EntityID,
					Domain:    "light",
				})
			}

		}
	}()

	//检查存在传感器是否检测到有人
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entities {
				automation.Conditions = append(automation.Conditions, Conditions{
					Type:      "is_not_occupied",
					DeviceID:  e.DeviceID,
					EntityID:  e.EntityID,
					Domain:    "binary_sensor",
					For:       &For{Minutes: 10},
					Condition: "device",
				})
			}
		}
	}()
	//开关状态切换,转无线模式下不能触发
	//func() {
	//	entities, ok := core.GetEntityCategoryMap()[core.CategorySwitchToggle]
	//	if ok {
	//		for _, e := range entities {
	//			if strings.Contains(e.OriginalName, "回家") {
	//				automation.Triggers = append(automation.Triggers, Triggers{
	//					EntityID: e.EntityID,
	//					Trigger:  "state",
	//				})
	//			}
	//		}
	//	}
	//
	//}()

	//条件：名字中带有“回家”的开关按键被按下,这个是单击事件，缺点是自动化不能作为动作
	/*
		alias: 回家自动化
		description: 门锁打开/或者开关按键触发用或条件，判断客厅所有灯是否打开，存在传感器是否检车到人
		triggers:
		  - entity_id: event.linp_cn_1137815613_qh2db4_click_e_10_1
		    trigger: state
		conditions:
		  - condition: state
		    entity_id: event.linp_cn_1137815613_qh2db4_click_e_10_1
		    attribute: 按键类型
		    state: 3
	*/
	//开关单击事件是否有这个场景
	func() {
		entitiesSenor, ok1 := core.GetEntityCategoryMap()[core.CategorySwitchScene]
		if ok1 {
			for _, e := range entitiesSenor {
				if strings.Contains(e.OriginalName, "回家") {
					//找到按键
					automation.Triggers = append(automation.Triggers, Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})
				}
			}
		}
	}()

	//场景按键是否有这个场景
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryScene]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "回家") {
					automation.Triggers = append(automation.Triggers, Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})
				}
			}
		}

	}()

	//打开客厅所有灯
	func() {
		var parallel = make(map[string][]interface{})

		entities, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
		if ok {
			//先开氛围灯
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					if strings.Contains(v.DeviceName, "氛围") {
						parallel["parallel"] = append(parallel["parallel"], &ActionLight{
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

			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					if !strings.Contains(v.DeviceName, "氛围") {
						parallel["parallel"] = append(parallel["parallel"], &ActionLight{
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

			if len(parallel) > 0 {
				automation.Actions = append(automation.Actions, parallel)
			}

		}
	}()

	//打开插座
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategorySocket]
		if ok {
			for _, e := range entities {
				automation.Actions = append(automation.Actions, &ActionCommon{
					Type:     "turn_on",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
			}
		}
	}()

	//关闭窗帘
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryCurtain]
		if ok {
			for _, e := range entities {
				automation.Actions = append(automation.Actions, &ActionCommon{
					Type:     "close",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "cover",
				})
			}
		}
	}()

	//打开电视
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryIrTV]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "红外电视控制") && strings.Contains(e.OriginalName, "开机") {
					automation.Actions = append(automation.Actions, &ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	//播放当前温度、湿度
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "室内温度和湿度是多少，如果检查不到室内情况，就播报室外温度和湿度", Title: ""},
					})
				}
			}
		}
	}()

	//todo 需要接入AI
	//是否打开空调
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		entitiesAir, okAir := core.GetEntityCategoryMap()[core.CategoryAirConditioner]

		var existAir bool

		if okAir {
			for _, e := range entitiesAir {
				if strings.Contains(e.AreaName, "客厅") {
					existAir = true
					break
				}
			}
		}

		if ok && existAir {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "播放文本") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "是否需要为你打开空调", Title: ""},
						Target: struct {
							DeviceID string `json:"device_id,omitempty"`
						}{DeviceID: e.DeviceID},
					})
				}

				if strings.Contains(e.OriginalName, "唤醒") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	//是否打开热水
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		_, okAir := core.GetEntityCategoryMap()[core.CategoryWaterHeater]

		if ok && okAir {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "播放文本") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "是否需要为你打开空调", Title: ""},
						Target: struct {
							DeviceID string `json:"device_id,omitempty"`
						}{DeviceID: e.DeviceID},
					})
				}

				if strings.Contains(e.OriginalName, "唤醒") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	//播放30秒音乐
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "播放音乐", Title: ""},
						Target: struct {
							DeviceID string `json:"device_id,omitempty"`
						}{DeviceID: e.DeviceID},
					})
				}
			}
		}
	}()
	// 撤防
	func() {
		automation.Actions = append(automation.Actions, &ActionService{
			Action: "automation.turn_off",
			Data:   map[string]interface{}{"stop_actions": true},
			Target: &struct {
				EntityId string `json:"entity_id"`
			}{EntityId: "automation.li_jia_bu_fang"},
		})
	}()

	// 5分钟后关闭音乐
	func() {
		automation.Actions = append(automation.Actions, &ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Minutes: 5},
		})
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "暂停") && strings.Contains(e.AreaName, "客厅") {
					automation.Actions = append(automation.Actions, &ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	return automation
}
