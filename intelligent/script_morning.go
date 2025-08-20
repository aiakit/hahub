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

		// 查找晚起床景开关
		var switchEntities []*data.Entity
		for _, e := range v {
			if e.Category == data.CategorySwitchScene {
				if strings.Contains(e.OriginalName, "起床") || strings.Contains(e.OriginalName, "早安") {
					switchEntities = append(switchEntities, e)
				}
			}
		}

		// 创建场景部分
		script := &Script{
			Alias:       areaName + "早安场景",
			Description: "执行" + areaName + "早安场景，包括播放音乐、打开窗帘、调节灯光",
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
								Delay: struct {
									Hours        int `json:"hours"`
									Minutes      int `json:"minutes"`
									Seconds      int `json:"seconds"`
									Milliseconds int `json:"milliseconds"`
								}{Seconds: 5},
							})
							xiaomiHomeSpeakerDeviceId = e.DeviceID
							break
						}
					}
				}

				for _, e := range speakers {
					if e.AreaID == areaId {
						var message = `主人，早安！愿这清新的晨光带给您无限的活力与希望。愿您在新的一天中与美好相遇，拥抱每一个温馨的瞬间。相信今天会是充满机遇的一天！`
						if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "播放文本") && strings.HasPrefix(e.EntityID, "notify.") {
							if e.DeviceID == xiaomiHomeSpeakerDeviceId {
								script.Sequence = append(script.Sequence, PlayText(e.EntityID, message))
								script.Sequence = append(script.Sequence, ActionTimerDelay{
									Delay: struct {
										Hours        int `json:"hours"`
										Minutes      int `json:"minutes"`
										Seconds      int `json:"seconds"`
										Milliseconds int `json:"milliseconds"`
									}{Seconds: int(x.GetPlaybackDuration(message).Seconds())},
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
									Delay: struct {
										Hours        int `json:"hours"`
										Minutes      int `json:"minutes"`
										Seconds      int `json:"seconds"`
										Milliseconds int `json:"milliseconds"`
									}{Seconds: 3},
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

		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct:   10,   // 先调至50%亮度
						ColorTempKelvin: 3000, // 温馨色温
					},
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
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
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 10},
		})

		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct: 25, // 降低亮度到25%
					},
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
		}

		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 10},
		})

		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_off",
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
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
					Delay: struct {
						Hours        int `json:"hours"`
						Minutes      int `json:"minutes"`
						Seconds      int `json:"seconds"`
						Milliseconds int `json:"milliseconds"`
					}{Minutes: 5},
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
					Delay: struct {
						Hours        int `json:"hours"`
						Minutes      int `json:"minutes"`
						Seconds      int `json:"seconds"`
						Milliseconds int `json:"milliseconds"`
					}{Minutes: 5},
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
									Delay: struct {
										Hours        int `json:"hours"`
										Minutes      int `json:"minutes"`
										Seconds      int `json:"seconds"`
										Milliseconds int `json:"milliseconds"`
									}{Seconds: 120},
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

			CreateScript(c, script)

			auto := &Automation{
				Alias:       areaName + "早安自动化",
				Description: "执行" + areaName + "早安场景，包括播放音乐、打开窗帘、调节灯光",
				Mode:        "single",
			}

			//条件：名字中带有"起床"/“早安”的开关按键和场景按键
			func() {
				for bName, v := range switchSelectSameName {
					bns := strings.Split(bName, "_")
					if len(bns) != 2 {
						continue
					}
					buttonName := bns[1]
					if strings.Contains(buttonName, "早安") || strings.Contains(buttonName, "起床") {
						//按键触发和条件
						for _, e := range v {
							auto.Triggers = append(auto.Triggers, Triggers{
								EntityID: e.EntityID,
								Trigger:  "state",
							})

							if e.Category == data.CategorySwitchClickOnce {
								auto.Conditions = append(auto.Conditions, Conditions{
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
				CreateAutomation(c, auto)
			}
		}
	}
}
