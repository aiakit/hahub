package intelligent

import (
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// 起床场景
func GoodMorningScript(c *ava.Context) {
	var entities = data.GetEntityAreaMap()

	if len(entities) == 0 {
		return
	}

	var entityMode, ok3 = data.GetEntityCategoryMap()[data.CategoryLightModel]

	for areaId, v := range entities {
		var parallel1 = make(map[string][]interface{})
		if ok3 {
			for _, e := range entityMode {
				for _, e1 := range v {
					if e1.DeviceID == e.DeviceID {
						actionCommon := handleDefaultGradientTimeSettings(e, 3)
						if actionCommon != nil {
							parallel1["parallel"] = append(parallel1["parallel"], actionCommon)
						}
						break
					}
				}
			}
		}

		areaName := data.SpiltAreaName(data.GetAreaName(areaId))

		// 检查是否是卧室区域
		isBedroom := strings.Contains(areaName, "卧室")

		// 如果不是卧室，则跳过
		if !isBedroom {
			continue
		}

		// 检查当前区域是否有灯组，如果没有则不创建场景
		hasLightGroup := false
		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				hasLightGroup = true
				break
			}
		}

		// 如果没有灯组，则不创建任何场景
		if !hasLightGroup {
			continue
		}

		// 创建场景部分
		script := &Script{
			Alias:       areaName + "早安场景",
			Description: "执行" + areaName + "早安场景，包括播放音乐、打开窗帘、调节灯光",
		}

		if len(parallel1) > 0 {
			script.Sequence = append(script.Sequence, parallel1)
		}

		var xiaomiHomeSpeakerDeviceId string
		// 1. 播放轻音乐
		func() {
			speakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
			if ok {
				for _, e := range speakers {
					if e.AreaID == areaId {
						if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "执行文本指令") && strings.HasPrefix(e.EntityID, "notify.") {
							script.Sequence = append(script.Sequence, ExecuteTextCommand(e.DeviceID, "播放一段轻音乐", true))
							script.Sequence = append(script.Sequence, ActionTimerDelay{
								Delay: &delay{
									Hours:        0,
									Minutes:      0,
									Seconds:      5,
									Milliseconds: 0,
								},
							})
							xiaomiHomeSpeakerDeviceId = e.DeviceID
							break
						}
					}
				}

				for _, e := range speakers {
					if e.AreaID == areaId {
						var message = `宿主，早安！愿这清新的晨光带给您无限的活力与希望。愿您在新的一天中与美好相遇，拥抱每一个温馨的瞬间。相信今天会是充满机遇的一天！`
						if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "播放文本") && strings.HasPrefix(e.EntityID, "notify.") {
							if e.DeviceID == xiaomiHomeSpeakerDeviceId {
								script.Sequence = append(script.Sequence, PlayText(e.EntityID, message))
								script.Sequence = append(script.Sequence, ActionTimerDelay{
									Delay: &delay{
										Hours:        0,
										Minutes:      0,
										Seconds:      int(x.GetPlaybackDuration(message).Seconds()),
										Milliseconds: 0,
									},
								})
								break
							}
						}
					}
				}

				for _, e := range speakers {
					if e.AreaID == areaId {
						if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "执行文本指令") && strings.HasPrefix(e.EntityID, "notify.") {
							if e.DeviceID == xiaomiHomeSpeakerDeviceId {
								script.Sequence = append(script.Sequence, ExecuteTextCommand(e.DeviceID, "告诉我今天早上天气怎么样出门需要注意什么", false))
								script.Sequence = append(script.Sequence, ActionTimerDelay{
									Delay: &delay{
										Hours:        0,
										Minutes:      0,
										Seconds:      3,
										Milliseconds: 0,
									},
								})
								break
							}
						}
					}
				}

			}
		}()

		// 添加设备操作到场景中
		for _, e := range v {
			if e.Category == data.CategoryCurtain {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "open",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "cover",
				})
			}
		}

		var actions = make([]interface{}, 0, 2)

		entityFilter := findLightsWithOutLightCategory("", v)
		if len(entityFilter) == 0 {
			continue
		}

		action := turnOnLights(entityFilter, 10, 3000, false)
		if len(action) == 0 {
			continue
		}

		for _, e := range action {
			actions = append(actions, e)
		}

		//// 3. 设置空调
		//for _, e := range v {
		//	if e.Category == data.CategoryAirConditioner {
		//		data := make(map[string]interface{})
		//		data["hvac_mode"] = "cool"
		//		data["temperature"] = 26
		//
		//		script.Sequence = append(script.Sequence, ActionService{
		//			Action: "climate.set_temperature",
		//			Data:   data,
		//			Target: &struct {
		//				EntityId string `json:"entity_id"`
		//			}{EntityId: e.EntityID},
		//		})
		//	}
		//}

		actions = append(actions, ActionTimerDelay{
			Delay: &delay{
				Hours:        0,
				Minutes:      0,
				Seconds:      60,
				Milliseconds: 0,
			},
		})

		for _, e := range action {
			var e1 = *e
			if e1.Data != nil {
				e1.Data.BrightnessStepPct = 20
			}
			actions = append(actions, e1)
		}

		//如果有床，设置床角度
		var exist bool
		for _, e := range v {
			if e.Category == data.CategoryBed && (strings.Contains(e.OriginalName, "腿部") || strings.Contains(e.OriginalName, "靠背")) && strings.HasPrefix(e.EntityID, "number.") {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    20,
				})
				exist = true
			}
		}

		if exist {
			script.Sequence = append(script.Sequence, ActionTimerDelay{
				Delay: &delay{
					Hours:        0,
					Minutes:      0,
					Seconds:      30,
					Milliseconds: 0,
				},
			})
		}

		for _, e := range v {
			if e.Category == data.CategoryBed && (strings.Contains(e.OriginalName, "腿部") || strings.Contains(e.OriginalName, "靠背")) && strings.HasPrefix(e.EntityID, "number.") {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    0,
				})
			}
		}

		actionOff := turnOffLights(entityFilter)
		if len(actionOff) == 0 {
			continue
		}

		for _, e := range actionOff {
			actions = append(actions, e)
		}

		if len(actions) > 0 {
			script.Sequence = append(script.Sequence, actions...)
		}

		func() {

			speakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
			if ok {
				for _, e := range speakers {
					if e.AreaID == areaId {
						if e.DeviceID == xiaomiHomeSpeakerDeviceId {
							if strings.HasPrefix(e.EntityID, "media_player.") {

								script.Sequence = append(script.Sequence, ActionTimerDelay{
									Delay: &delay{
										Hours:        0,
										Minutes:      0,
										Seconds:      10,
										Milliseconds: 0,
									},
								})
								script.Sequence = append(script.Sequence, ActionService{
									Action: "media_player.media_pause",
									Target: &struct {
										EntityId string `json:"entity_id"`
									}{EntityId: e.EntityID}})
								break
							}
						}
					}
				}
			}
		}()

		//改回去
		var parallel2 = make(map[string][]interface{})
		if ok3 {
			for _, e := range entityMode {
				for _, e1 := range v {
					if e1.DeviceID == e.DeviceID {
						actionCommon := handleDefaultGradientTimeSettings(e, 1)
						if actionCommon != nil {
							parallel2["parallel"] = append(parallel2["parallel"], actionCommon)
						}
						break
					}
				}
			}
		}

		if len(parallel2) > 0 {
			script.Sequence = append(script.Sequence, parallel2)
		}

		// 创建自动化部分
		if len(script.Sequence) > 0 {

			AddScript2Queue(c, script)

			auto := &Automation{
				Alias:       areaName + "早安自动化",
				Description: "执行" + areaName + "早安场景，包括播放音乐、打开窗帘、调节灯光",
				Mode:        "single",
			}

			//条件：名字中带有"起床"/“早安”的开关按键和场景按键
			func() {
				for bName, v1 := range switchSelectSameName {
					bns := strings.Split(bName, "_")
					if len(bns) < 2 {
						continue
					}
					buttonName := bns[len(bns)-1]
					if strings.Contains(buttonName, "早安") || strings.Contains(buttonName, "起床") {
						//按键触发和条件
						var con = &Conditions{
							Condition: "or",
						}

						for _, e := range v1 {
							if e.AreaID != areaId {
								continue
							}
							auto.Triggers = append(auto.Triggers, &Triggers{
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
							auto.Conditions = append(auto.Conditions, con)
						}
					}

				}
			}()

			// 执行场景
			auto.Actions = script.Sequence

			// 创建自动化
			if len(auto.Triggers) > 0 && len(auto.Actions) > 0 {
				AddAutomation2Queue(c, auto)
			}
		}
	}
}
