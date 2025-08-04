package intelligent

import (
	"hahub/hub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 睡觉场景
func goodNightScript(c *ava.Context) {
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

		// 查找晚睡觉景开关
		var switchEntities []*data.Entity
		for _, e := range v {
			if e.Category == data.CategorySwitchScene {
				if strings.Contains(e.OriginalName, "睡觉") || strings.Contains(e.OriginalName, "晚安") {
					switchEntities = append(switchEntities, e)
				}
			}
		}

		// 创建场景部分
		script := &Script{
			Alias:       areaName + "睡觉/晚安场景",
			Description: "执行" + areaName + "睡觉场景，包括播放音乐、关闭窗帘、调节灯光和控制空调",
		}

		// 1. 播放30秒轻音乐
		func() {
			speakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
			if ok {
				for _, e := range speakers {
					if e.AreaID == areaId {
						if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "执行文本指令") {
							script.Sequence = append(script.Sequence, ExecuteTextCommand(e.EntityID, "播放一段轻快的音乐", true))
							break
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

		// 2. 设置灯的色温
		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct:   50,   // 先调至50%亮度
						ColorTempKelvin: 3000, // 温馨色温
					},
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
		}

		// 3. 设置空调
		for _, e := range v {
			if e.Category == data.CategoryAirConditioner {
				data := make(map[string]interface{})
				data["temperature"] = 26
				data["mode"] = "cool"

				script.Sequence = append(script.Sequence, ActionService{
					Action: "climate.turn_on",
					Data:   data,
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})
			}
		}

		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 30}, // 等待30秒
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

		// 关灯
		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 30}, // 再等待30秒
		})

		for _, e := range v {
			if e.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_off",
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
		}

		// 创建自动化部分
		if len(script.Sequence) > 0 {

			CreateScript(c, script)

			auto := &Automation{
				Alias:       areaName + "睡觉自动化",
				Description: "执行" + areaName + "睡觉场景，包括播放音乐、关闭窗帘、调节灯光和控制空调",
				Mode:        "single",
			}

			//条件：名字中带有"睡觉"/“晚安”的开关按键和场景按键
			func() {
				for bName, v := range switchSelectSameName {
					bns := strings.Split(bName, "_")
					if len(bns) != 2 {
						continue
					}
					buttonName := bns[1]
					if strings.Contains(buttonName, "晚安") || strings.Contains(buttonName, "睡觉") {
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
