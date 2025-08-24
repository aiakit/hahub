package intelligent

import (
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// 晚安场景
func GoodNightScript(c *ava.Context) {
	var entities = data.GetEntityAreaMap()

	if len(entities) == 0 {
		return
	}

	for areaId, v := range entities {
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

		// 查找晚晚安景开关
		var switchEntities []*data.Entity
		for _, e := range v {
			if e.Category == data.CategorySwitchScene {
				if strings.Contains(e.OriginalName, "晚安") || strings.Contains(e.OriginalName, "睡觉") {
					switchEntities = append(switchEntities, e)
				}
			}
		}

		// 创建场景部分
		script := &Script{
			Alias:       areaName + "晚安场景",
			Description: "执行" + areaName + "晚安场景，包括播放音乐、打开窗帘、调节灯光",
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
						var message = `主人，晚安。愿这宁静的时刻带给你无尽的安宁与温暖。愿你与美好相遇，拥抱每一个温馨的瞬间。愿明天的阳光带给你新的希望与活力。`
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
								script.Sequence = append(script.Sequence, ExecuteTextCommand(e.DeviceID, "告诉我明天早上天气怎么样出门需要注意什么", false))
								script.Sequence = append(script.Sequence, ActionTimerDelay{
									Delay: &delay{
										Hours:        0,
										Minutes:      0,
										Seconds:      5,
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
		// 1. 关闭窗帘
		for _, e := range v {
			if e.Category == data.CategoryCurtain {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "close",
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

		action := turnOnLights(entityFilter, 50, 3000, false)
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

		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: &delay{
				Hours:        0,
				Minutes:      0,
				Seconds:      10,
				Milliseconds: 0,
			},
		})

		for _, e := range action {
			e.Data.BrightnessStepPct = 25
			actions = append(actions, e)
		}

		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: &delay{
				Hours:        0,
				Minutes:      0,
				Seconds:      10,
				Milliseconds: 0,
			},
		})

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

		//如果有床，设置床角度
		for _, e := range v {
			if e.Category == data.CategoryBed && strings.Contains(e.OriginalName, "腿部") && strings.HasPrefix(e.EntityID, "number.") {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    20,
				})

				script.Sequence = append(script.Sequence, ActionTimerDelay{
					Delay: &delay{
						Hours:        0,
						Minutes:      0,
						Seconds:      5,
						Milliseconds: 0,
					},
				})

				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    0,
				})

			}

			if e.Category == data.CategoryBed && strings.Contains(e.OriginalName, "靠背") && strings.HasPrefix(e.EntityID, "number.") {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    20,
				})

				script.Sequence = append(script.Sequence, ActionTimerDelay{
					Delay: &delay{
						Hours:        0,
						Minutes:      0,
						Seconds:      5,
						Milliseconds: 0,
					},
				})
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "set_value",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "number",
					Value:    0,
				})
			}
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
										Minutes:      2,
										Seconds:      0,
										Milliseconds: 0,
									},
								})

								script.Sequence = append(script.Sequence, ActionService{
									Action: "media_player.media_pause",
									Target: &struct {
										EntityId string `json:"entity_id"`
									}{EntityId: e.EntityID}})
							}
							break
						}
					}
				}
			}
		}()

		// 创建自动化部分
		if len(script.Sequence) > 0 {

			AddScript2Queue(c, script)

			auto := &Automation{
				Alias:       areaName + "晚安自动化",
				Description: "执行" + areaName + "晚安场景，包括播放音乐、关闭窗帘、调节灯光和控制空调",
				Mode:        "single",
			}

			//条件：名字中带有"晚安"/“晚安”的开关按键和场景按键
			func() {
				for bName, v := range switchSelectSameName {
					bns := strings.Split(bName, "_")
					if len(bns) != 2 {
						continue
					}
					buttonName := bns[1]
					if strings.Contains(buttonName, "睡觉") || strings.Contains(buttonName, "晚安") {
						//按键触发和条件
						for _, e := range v {
							auto.Triggers = append(auto.Triggers, &Triggers{
								EntityID: e.EntityID,
								Trigger:  "state",
							})

							if e.Category == data.CategorySwitchClickOnce {
								auto.Conditions = append(auto.Conditions, &Conditions{
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

			// 执行场景
			auto.Actions = script.Sequence

			// 创建自动化
			if len(auto.Triggers) > 0 && len(auto.Actions) > 0 {
				AddAutomation2Queue(c, auto)
			}
		}
	}
}
