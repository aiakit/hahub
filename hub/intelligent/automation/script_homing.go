package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

func initHoming(c *ava.Context) {
	// 创建回家脚本
	script := homingScript()
	if script != nil && len(script.Sequence) > 0 {
		scriptId := CreateScript(c, script)

		// 基于脚本创建自动化
		if scriptId != "" {
			auto := homingAutomation(scriptId)
			if auto != nil && len(auto.Triggers) > 0 {
				CreateAutomation(c, auto)
			}
		}
	}
}

// 回家场景脚本
func homingScript() *Script {
	var script = &Script{
		Alias:       "回家脚本",
		Description: "回家场景执行脚本",
	}

	//打开客厅所有灯
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
		if ok {
			//先开氛围灯
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					if strings.Contains(v.DeviceName, "氛围") {
						script.Sequence = append(script.Sequence, ActionLight{
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
						script.Sequence = append(script.Sequence, ActionLight{
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
	}()

	//打开插座
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategorySocket]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionCommon{
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
				script.Sequence = append(script.Sequence, ActionCommon{
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
					script.Sequence = append(script.Sequence, ActionCommon{
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
					script.Sequence = append(script.Sequence, ActionNotify{
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
					script.Sequence = append(script.Sequence, ActionNotify{
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
					script.Sequence = append(script.Sequence, ActionCommon{
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
					script.Sequence = append(script.Sequence, ActionNotify{
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
					script.Sequence = append(script.Sequence, ActionCommon{
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
					script.Sequence = append(script.Sequence, ActionNotify{
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
	script.Sequence = append(script.Sequence, ActionService{
		Action: "automation.turn_off",
		Data:   map[string]interface{}{"stop_actions": true},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "automation.li_jia_bu_fang"},
	})

	// 5分钟后关闭音乐
	func() {
		script.Sequence = append(script.Sequence, ActionTimerDelay{
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
					script.Sequence = append(script.Sequence, ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	return script
}

// 回家自动化
func homingAutomation(scriptId string) *Automation {
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
				for _, e := range v {
					automation.Triggers = append(automation.Triggers, Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})

					if e.Category == core.CategorySwitchClickOnce {
						automation.Conditions = append(automation.Conditions, Conditions{
							Condition: "state",
							EntityID:  e.EntityID,
							Attribute: e.Attribute,
							State:     e.SeqButton,
						})
					}
				}
			}

		}
	}()

	// 执行回家脚本
	automation.Actions = append(automation.Actions, ActionService{
		Action: "script.turn_on",
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "script." + scriptId},
	})

	return automation
}
