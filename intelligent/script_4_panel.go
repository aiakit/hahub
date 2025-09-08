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

	var allOffLight = make([]*ActionLight, 0, 10)

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
				turnOffScriptId = AddScript2Queue(c, s)
			}

			if entityId != "" {
				//两个自动化
				//1.按下开关执行脚本

				if turnOnScriptId != "" {
					var a = &Automation{
						Alias:       "HomePanel" + areaName + "开灯",
						Description: "3D面板上按下按键开灯",
						Mode:        "single",
						Triggers: []*Triggers{{
							EntityID: entityId,
							Trigger:  "state",
							To:       "on",
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

				if turnOffScriptId != "" {
					var a = &Automation{
						Alias:       "HomePanel" + areaName + "关灯",
						Description: "3D面板上按下按键关灯",
						Mode:        "single",
						Triggers: []*Triggers{{
							EntityID: entityId,
							Trigger:  "state",
							To:       "off",
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

				//1.区域任意一个灯开或者灯关，打开/关闭开关
				func() {
					var a = &Automation{
						Alias:       "HomePanel" + areaName + "3d面板按键为开",
						Description: areaName + "灯开修改3d面板按键为开",
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
					})

					a.Triggers = triggers
					if len(a.Triggers) > 0 {
						a.Actions = append(a.Actions, &ActionLight{
							Action: "input_boolean.turn_on",
							Target: &targetLightData{
								EntityId: entityId,
							},
						})
						AddAutomation2Queue(c, a)
					}
				}()

				func() {
					var a = &Automation{
						Alias:       "HomePanel" + areaName + "3d面板按键为关",
						Description: areaName + "灯开修改3d面板按键为关",
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
					})

					a.Triggers = triggers
					if len(a.Triggers) > 0 {
						a.Actions = append(a.Actions, &ActionLight{
							Action: "input_boolean.turn_off",
							Target: &targetLightData{
								EntityId: entityId,
							},
						})
						AddAutomation2Queue(c, a)
					}
				}()
			}
		}
	}

	if len(allOffLight) > 0 {
		var s1 = &Script{
			Alias:       "全屋关灯",
			Description: "全屋全部区域直接关灯",
		}
		for _, a := range allOffLight {
			s1.Sequence = append(s1.Sequence, a)
		}

		if len(s1.Sequence) > 0 {
			AddScript2Queue(c, s1)
		}
	}
}
