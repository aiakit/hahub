package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func InitHoming(c *ava.Context) {
	// 创建回家场景
	script := homingScript()
	if script != nil && len(script.Sequence) > 0 {
		scriptId := CreateScript(c, script)

		// 基于场景创建自动化
		if scriptId != "" {
			auto := homingAutomation(scriptId)
			if auto != nil && len(auto.Triggers) > 0 {
				CreateAutomation(c, auto)
			}
		}
	}
}

// 回家场景场景
func homingScript() *Script {
	var script = &Script{
		Alias:       "回家场景",
		Description: "回家场景执行场景",
	}

	// 撤防
	script.Sequence = append(script.Sequence, ActionService{
		Action: "automation.turn_off",
		Data:   map[string]interface{}{"stop_actions": true},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "automation.li_jia_bu_fang"},
	})

	var sequence = make(map[string][]interface{})

	var xiaomiHomeSpeakerDeviceId string
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "播放文本") && strings.Contains(e.AreaName, "客厅") && strings.HasPrefix(e.EntityID, "notify.") {
					script.Sequence = append(script.Sequence, ActionTimerDelay{
						Delay: struct {
							Hours        int `json:"hours"`
							Minutes      int `json:"minutes"`
							Seconds      int `json:"seconds"`
							Milliseconds int `json:"milliseconds"`
						}{Seconds: 3},
					})

					script.Sequence = append(script.Sequence, ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "欢迎主人回家"},
						Target: struct {
							DeviceID string `json:"device_id,omitempty"`
						}{DeviceID: e.DeviceID},
					})
					script.Sequence = append(script.Sequence, ActionTimerDelay{
						Delay: struct {
							Hours        int `json:"hours"`
							Minutes      int `json:"minutes"`
							Seconds      int `json:"seconds"`
							Milliseconds int `json:"milliseconds"`
						}{Seconds: 3},
					})

					xiaomiHomeSpeakerDeviceId = e.DeviceID
					break
				}
			}
		}

		for _, e := range entities {
			if e.DeviceID == xiaomiHomeSpeakerDeviceId {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") && strings.HasPrefix(e.EntityID, "text.") {
					script.Sequence = append(script.Sequence, ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "[播放一段轻音乐,true]"},
						Target: struct {
							DeviceID string `json:"device_id,omitempty"`
						}{DeviceID: e.DeviceID},
					})
					break
				}
			}
		}
	}()

	//打开客厅所有灯
	func() {
		entitiesScript, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		var isOpen bool
		if ok {
			for _, e := range entitiesScript {
				if strings.Contains(e.OriginalName, "逐个") {
					script.Sequence = append(script.Sequence, ActionService{
						Action: "script.turn_on",
						Target: &struct {
							EntityId string `json:"entity_id"`
						}{EntityId: e.EntityID},
					})
					isOpen = true
				}
			}
		}

		if !isOpen {
			//判断是否有展示脚本,如果有，使用展示脚本
			entities, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
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
		}
	}()

	//打开插座
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategorySocket]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.AreaName, "客厅") {
					script.Sequence = append(script.Sequence, ActionCommon{
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
					script.Sequence = append(script.Sequence, ActionCommon{
						Type:     "open",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "cover",
					})
				}
			}
		}
	}()

	//播放当前温度、湿度
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if e.DeviceID == xiaomiHomeSpeakerDeviceId {
					if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") && strings.HasPrefix(e.EntityID, "text.") {
						sequence["sequence"] = append(sequence["sequence"], ActionNotify{
							Action: "notify.send_message",
							Data: struct {
								Message string `json:"message,omitempty"`
								Title   string `json:"title,omitempty"`
							}{Message: "[室内湿度,false]", Title: ""},
							Target: struct {
								DeviceID string `json:"device_id,omitempty"`
							}{DeviceID: e.DeviceID},
						})

						sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
							Delay: struct {
								Hours        int `json:"hours"`
								Minutes      int `json:"minutes"`
								Seconds      int `json:"seconds"`
								Milliseconds int `json:"milliseconds"`
							}{Seconds: 4},
						})

						sequence["sequence"] = append(sequence["sequence"], ActionNotify{
							Action: "notify.send_message",
							Data: struct {
								Message string `json:"message,omitempty"`
								Title   string `json:"title,omitempty"`
							}{Message: "[室内温度,false]", Title: ""},
							Target: struct {
								DeviceID string `json:"device_id,omitempty"`
							}{DeviceID: e.DeviceID},
						})
						break
					}
				}
			}
		}
	}()

	var turnOnMessage = "是否需要为你打开"

	//打开电视，直接打开，默认离家场景已经关闭电视了，如果存在传感器很久没人，在弄一个关闭电视的自动化
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryIrTV]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.AreaName, "客厅") {
					if strings.Contains(e.OriginalName, "红外电视控制") && strings.Contains(e.OriginalName, "开机") {
						turnOnMessage += "电视，"
					}
				}
			}
		}
	}()

	//是否打开空调
	func() {
		_, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		entitiesAir, okAir := data.GetEntityCategoryMap()[data.CategoryAirConditioner]

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
			for _, e := range entitiesAir {
				if strings.Contains(e.AreaName, "客厅") {
					turnOnMessage += "客厅空调，"
					break
				}
			}
		}
	}()

	//是否打开热水
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		_, okAir := data.GetEntityCategoryMap()[data.CategoryWaterHeater]

		if ok && okAir {
			for _, e := range entities {
				if e.DeviceID == xiaomiHomeSpeakerDeviceId {
					if strings.Contains(e.OriginalName, "播放文本") && strings.Contains(e.AreaName, "客厅") && strings.HasPrefix(e.EntityID, "notify.") {
						sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
							Delay: struct {
								Hours        int `json:"hours"`
								Minutes      int `json:"minutes"`
								Seconds      int `json:"seconds"`
								Milliseconds int `json:"milliseconds"`
							}{Seconds: 4},
						})

						sequence["sequence"] = append(sequence["sequence"], ActionNotify{
							Action: "notify.send_message",
							Data: struct {
								Message string `json:"message,omitempty"`
								Title   string `json:"title,omitempty"`
							}{Message: turnOnMessage + "热水器"},
							Target: struct {
								DeviceID string `json:"device_id,omitempty"`
							}{DeviceID: e.DeviceID},
						})
					}
				}
			}

			for _, e := range entities {
				if e.DeviceID == xiaomiHomeSpeakerDeviceId {
					if strings.Contains(e.OriginalName, "唤醒") && strings.Contains(e.AreaName, "客厅") {
						sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
							Delay: struct {
								Hours        int `json:"hours"`
								Minutes      int `json:"minutes"`
								Seconds      int `json:"seconds"`
								Milliseconds int `json:"milliseconds"`
							}{Seconds: 6},
						})

						sequence["sequence"] = append(sequence["sequence"], ActionCommon{
							Type:     "press",
							DeviceID: e.DeviceID,
							EntityID: e.EntityID,
							Domain:   "button",
						})
					}
				}
			}
		}
	}()

	func() {
		sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 60},
		})
		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if e.DeviceID == xiaomiHomeSpeakerDeviceId {
					if strings.HasPrefix(e.EntityID, "media_player.") && strings.Contains(e.AreaName, "客厅") {
						sequence["sequence"] = append(sequence["sequence"], ActionService{
							Action: "media_player.media_pause",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: e.EntityID}})
					}
				}
			}
		}
	}()

	if len(sequence) > 0 {
		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 6},
		})
		script.Sequence = append(script.Sequence, sequence)
	}

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
		entities, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
		if ok {
			for _, e := range entities {
				automation.Conditions = append(automation.Conditions, &Conditions{
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
		entities, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entities {
				automation.Conditions = append(automation.Conditions, &Conditions{
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
					automation.Triggers = append(automation.Triggers, &Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})

					if e.Category == data.CategorySwitchClickOnce {
						automation.Conditions = append(automation.Conditions, &Conditions{
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

	// 执行回家场景
	automation.Actions = append(automation.Actions, ActionService{
		Action: "script.turn_on",
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "script." + scriptId},
	})

	return automation
}
