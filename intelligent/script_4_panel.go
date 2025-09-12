package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 针对面板，其他可用
func Panel(c *ava.Context) {
	entitiesBoolean := data.GetEntityCategoryMap()[data.CategoryInputBoolean]
	area := data.GetAreaMap() //key:areaid value:areaname
	entities := data.GetEntityAreaMap()
	if len(entities) == 0 {
		return
	}

	for areaId, areaName := range area {
		var entityId string
		areaName = data.SpiltAreaName(areaName)
		for _, v := range entitiesBoolean {
			if !strings.Contains(v.OriginalName, "电视") {
				continue
			}

			if strings.Contains(v.OriginalName, areaName) {
				entityId = v.EntityID
				break
			}
		}

		if entityId != "" {
			turnOn := TurnOnTv(areaId)
			if len(turnOn) > 0 {
				var s = &Script{
					Alias:       areaName + "打开电视",
					Description: areaName + "区域灯直接打开电视",
				}
				s.Sequence = turnOn
				turnOnScriptId := AddScript2Queue(c, s)

				var a = &Automation{
					Alias:       "视图" + areaName + "打开电视",
					Description: "视图中按下按键打开电视",
					Mode:        "single",
					Triggers: []*Triggers{{
						EntityID: entityId,
						Trigger:  "state",
						To:       "on",
						From:     "off",
					}},
				}
				a.Actions = append(a.Actions, &ActionService{
					Action: "script.turn_on",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: turnOnScriptId},
				})
				AddAutomation2Queue(c, a)
			}

			turnOff := TurnOffTv(areaId)
			if len(turnOff) > 0 {
				var s = &Script{
					Alias:       areaName + "关闭电视",
					Description: areaName + "区域灯直接关闭电视",
				}
				s.Sequence = turnOff

				turnOffScriptId := AddScript2Queue(c, s)

				var a = &Automation{
					Alias:       "视图" + areaName + "关闭电视",
					Description: "视图中按下按键关闭电视",
					Mode:        "single",
					Triggers: []*Triggers{{
						EntityID: entityId,
						Trigger:  "state",
						To:       "off",
						From:     "on",
					}},
				}
				a.Actions = append(a.Actions, &ActionService{
					Action: "script.turn_on",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: turnOffScriptId},
				})
				AddAutomation2Queue(c, a)
			}
		}

	}

	var allOffLight = make([]*ActionLight, 0, 10)
	var allOnLight = make([]*ActionLight, 0, 10)

	for areaId, areaName := range area {
		var entityId string
		areaName = data.SpiltAreaName(areaName)
		for _, v := range entitiesBoolean {
			if !strings.Contains(v.OriginalName, "灯") {
				continue
			}

			if strings.Contains(v.OriginalName, areaName) {
				entityId = v.EntityID
				break
			}
		}

		if v, _ := entities[areaId]; len(v) > 0 {
			filter := findLightsWithOutLightCategory("", v)
			if len(filter) == 0 {
				continue
			}
			actions := turnOnLights(filter, 100, 4800, false)
			if len(actions) == 0 {
				continue
			}

			var s = &Script{
				Alias:       areaName + "开灯",
				Description: areaName + "区域灯直接打开",
			}
			for _, a := range actions {
				s.Sequence = append(s.Sequence, a)
				allOnLight = append(allOnLight, a)
			}

			var turnOnScriptId string
			if len(s.Sequence) > 0 {
				turnOnScriptId = AddScript2Queue(c, s)
			}

			actionsOff := turnOffLights(filter)
			if len(actionsOff) == 0 {
				continue
			}

			var s1 = &Script{
				Alias:       areaName + "关灯",
				Description: areaName + "区域灯直接关灯",
			}
			for _, a := range actionsOff {
				s1.Sequence = append(s1.Sequence, a)
				allOffLight = append(allOffLight, a)
			}

			var turnOffScriptId string
			if len(s1.Sequence) > 0 {
				turnOffScriptId = AddScript2Queue(c, s1)
			}

			var turnOnAutoId string
			var turnOffAutoId string
			if entityId != "" {
				//两个自动化
				//1.按下开关执行脚本

				if turnOnScriptId != "" {
					var a = &Automation{
						Alias:       "视图" + areaName + "开灯",
						Description: "视图中按下按键开灯",
						Mode:        "single",
						Triggers: []*Triggers{{
							EntityID: entityId,
							Trigger:  "state",
							To:       "on",
							From:     "off",
						}},
					}

					a.Actions = append(a.Actions, &ActionService{
						Action: "script.turn_on",
						Target: &struct {
							EntityId string `json:"entity_id"`
						}{EntityId: turnOnScriptId},
					})
					turnOnAutoId = AddAutomation2Queue(c, a)
				}

				if turnOffScriptId != "" {
					var a = &Automation{
						Alias:       "视图" + areaName + "关灯",
						Description: "视图中按下按键关灯",
						Mode:        "single",
						Triggers: []*Triggers{{
							EntityID: entityId,
							Trigger:  "state",
							To:       "off",
							From:     "on",
						}},
					}

					a.Actions = append(a.Actions, &ActionService{
						Action: "script.turn_on",
						Target: &struct {
							EntityId string `json:"entity_id"`
						}{EntityId: turnOffScriptId},
					})
					turnOffAutoId = AddAutomation2Queue(c, a)
				}
			}

			if entityId != "" {
				func() {
					var a = &Automation{
						Alias:       "视图" + areaName + "修改按键为开",
						Description: areaName + "灯开修改按键为开",
						Mode:        "single",
					}
					var triggers []*Triggers
					for _, act := range actions {
						if !strings.Contains(act.Action, "light") && !strings.Contains(act.Action, "switch") {
							continue
						}
						if act.Target == nil {
							continue
						}

						if act.Target.DeviceId == "" && act.Target.EntityId == "" {
							continue
						}

						triggers = append(triggers, &Triggers{
							Type:     "turned_on",
							DeviceID: act.Target.DeviceId,
							EntityID: act.Target.EntityId,
							Domain:   "light",
							Trigger:  "device",
						})
					}
					a.Conditions = append(a.Conditions, &Conditions{
						Condition: "state",
						EntityID:  entityId,
						State:     "off",
						For: &For{
							Hours:   0,
							Minutes: 0,
							Seconds: 1,
						},
					})

					a.Triggers = triggers
					if len(a.Triggers) > 0 {
						a.Actions = append(a.Actions, &ActionService{
							Action: "automation.turn_off",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: turnOnAutoId},
						})
						a.Actions = append(a.Actions, &ActionLight{
							Action: "input_boolean.turn_on",
							Target: &targetLightData{
								EntityId: entityId,
							},
						})
						a.Actions = append(a.Actions, &ActionTimerDelay{Delay: &delay{
							Hours:        0,
							Minutes:      0,
							Seconds:      1,
							Milliseconds: 0,
						}})
						a.Actions = append(a.Actions, &ActionService{
							Action: "automation.turn_on",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: turnOnAutoId},
						})
						AddAutomation2Queue(c, a)
					}
				}()

				func() {
					var a = &Automation{
						Alias:       "视图" + areaName + "修改按键为关",
						Description: areaName + "灯开修改按键为关",
						Mode:        "single",
					}
					var triggers []*Triggers
					for _, act := range actionsOff {
						if !strings.Contains(act.Action, "light") && !strings.Contains(act.Action, "switch") {
							continue
						}
						if act.Target == nil {
							continue
						}

						if act.Target.DeviceId == "" && act.Target.EntityId == "" {
							continue
						}

						triggers = append(triggers, &Triggers{
							Type:     "turned_off",
							DeviceID: act.Target.DeviceId,
							EntityID: act.Target.EntityId,
							Domain:   "light",
							Trigger:  "device",
						})
					}

					a.Conditions = append(a.Conditions, &Conditions{
						Condition: "state",
						EntityID:  entityId,
						State:     "on",
						For: &For{
							Hours:   0,
							Minutes: 0,
							Seconds: 1,
						},
					})

					a.Triggers = triggers
					if len(a.Triggers) > 0 {
						a.Actions = append(a.Actions, &ActionService{
							Action: "automation.turn_off",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: turnOffAutoId},
						})
						a.Actions = append(a.Actions, &ActionLight{
							Action: "input_boolean.turn_off",
							Target: &targetLightData{
								EntityId: entityId,
							},
						})
						a.Actions = append(a.Actions, &ActionTimerDelay{Delay: &delay{
							Hours:        0,
							Minutes:      0,
							Seconds:      1,
							Milliseconds: 0,
						}})
						a.Actions = append(a.Actions, &ActionService{
							Action: "automation.turn_on",
							Target: &struct {
								EntityId string `json:"entity_id"`
							}{EntityId: turnOffAutoId},
						})

						AddAutomation2Queue(c, a)
					}
				}()
			}
		}
	}

	if len(allOffLight) > 0 {
		var s1 = &Script{
			Alias:       "关闭所有灯",
			Description: "全屋全部区域直接关灯",
		}
		for _, a := range allOffLight {
			s1.Sequence = append(s1.Sequence, a)
		}

		if len(s1.Sequence) > 0 {
			AddScript2Queue(c, s1)
			//var a = &Automation{
			//	Alias:       "视图" + "全屋关灯",
			//	Description: "视图中按下按键全屋关灯",
			//	Mode:        "single",
			//	Triggers: []*Triggers{{
			//		EntityID: "input_boolean.quan_wu_guan_deng",
			//		Trigger:  "state",
			//		To:       "on",
			//		From:     "off",
			//	}},
			//}
			//a.Actions = append(a.Actions, &ActionService{
			//	Action: "script.turn_on",
			//	Target: &struct {
			//		EntityId string `json:"entity_id"`
			//	}{EntityId: id},
			//})
			//AddAutomation2Queue(c, a)
		}
	}

	if len(allOnLight) > 0 {
		var s1 = &Script{
			Alias:       "打开所有灯",
			Description: "全屋全部区域直接开灯",
		}
		for _, a := range allOnLight {
			s1.Sequence = append(s1.Sequence, a)
		}

		if len(s1.Sequence) > 0 {
			AddScript2Queue(c, s1)
			//id := AddScript2Queue(c, s1)
			//var a = &Automation{
			//	Alias:       "视图" + "全屋开灯",
			//	Description: "视图中按下按键全屋开灯",
			//	Mode:        "single",
			//	Triggers: []*Triggers{{
			//		EntityID: "input_boolean.quan_wu_kai_deng",
			//		Trigger:  "state",
			//		To:       "on",
			//		From:     "off",
			//	}},
			//}
			//a.Actions = append(a.Actions, &ActionService{
			//	Action: "script.turn_on",
			//	Target: &struct {
			//		EntityId string `json:"entity_id"`
			//	}{EntityId: id},
			//})
			//AddAutomation2Queue(c, a)
		}
	}
}
