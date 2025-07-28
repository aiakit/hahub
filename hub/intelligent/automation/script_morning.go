package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

// 早安场景
func goodMorningScript(c *ava.Context) {
	var entities = core.GetEntityAreaMap()

	if len(entities) == 0 {
		return
	}

	for areaId, v := range entities {
		areaName := core.SpiltAreaName(core.GetAreaName(areaId))

		// 检查是否是卧室区域
		isBedroom := strings.Contains(areaName, "卧室")

		// 如果不是卧室，则跳过
		if !isBedroom {
			continue
		}

		// 检查当前区域是否有灯组，如果没有则不创建场景
		hasLightGroup := false
		for _, e := range v {
			if e.Category == core.CategoryLightGroup {
				hasLightGroup = true
				break
			}
		}

		// 如果没有灯组，则不创建任何场景
		if !hasLightGroup {
			continue
		}

		// 查找早安场景开关
		var switchEntities []*core.Entity
		for _, e := range v {
			if e.Category == core.CategorySwitchScene {
				if strings.Contains(e.OriginalName, "早安") || strings.Contains(e.OriginalName, "起床") {
					switchEntities = append(switchEntities, e)
				}
			}
		}

		// 创建脚本部分
		script := &Script{
			Alias:       areaName + "早安脚本",
			Description: "执行" + areaName + "早安操作，包括播放音乐、打开窗帘、调节灯光和控制空调",
		}

		// 1. 播放30秒轻音乐
		func() {
			entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
			if ok {
				for _, e := range entities {
					if strings.Contains(e.AreaName, areaName) && strings.Contains(e.OriginalName, "执行文本指令") {
						script.Sequence = append(script.Sequence, ExecuteTextCommand(e.EntityID, "播放一段轻快的音乐", true))
						break
					}
				}
			}
		}()

		// 添加设备操作到脚本中
		// 1. 打开窗帘
		for _, e := range v {
			if e.Category == core.CategoryCurtain {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "open",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "cover",
				})
			}
		}

		// 2. 设置灯的色温
		for _, e := range v {
			if e.Category == core.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct:   75,   // 调至75%亮度
						ColorTempKelvin: 5000, // 清醒色温
					},
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
		}

		// 3. 设置空调
		for _, e := range v {
			if e.Category == core.CategoryAirConditioner {
				data := make(map[string]interface{})
				data["temperature"] = 24
				data["mode"] = "auto"

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
			if e.Category == core.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct: 100, // 提高亮度到100%
					},
					Target: &targetLightData{DeviceId: e.DeviceID},
				})
			}
		}

		// 创建自动化部分
		if len(script.Sequence) > 0 {

			CreateScript(c, script)

			auto := &Automation{
				Alias:       areaName + "早安自动化",
				Description: "执行" + areaName + "早安脚本，包括播放音乐、打开窗帘、调节灯光和控制空调",
				Mode:        "single",
			}

			//条件：名字中带有"早安"或"起床"的开关按键和场景按键
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

							if e.Category == core.CategorySwitchClickOnce {
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

			// 执行脚本
			auto.Actions = script.Sequence

			// 创建自动化
			if len(auto.Triggers) > 0 && len(auto.Actions) > 0 {
				CreateAutomation(c, auto)
			}
		}
	}
}
