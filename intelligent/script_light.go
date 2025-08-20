package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func lightScriptSetting(c *ava.Context) {
	lightScene(c, "会客", 80, 6000)
	lightScene(c, "观影", 20, 3800)
	lightScene(c, "游戏", 100, 4800)
	lightScene(c, "棋牌", 100, 4800)
	lightScene(c, "喝茶", 80, 3500)
	lightScene(c, "休息", 60, 3000)
	lightScene(c, "日常", 80, 4500)
	lightScene(c, "阅读", 70, 4000)
	lightScene(c, "清洁", 100, 6000)
	lightScene(c, "就餐", 70, 3500)
	lightScene(c, "日光", 100, 5500)
	lightScene(c, "月光", 20, 2700)
	lightScene(c, "黄昏", 100, 2700)
	lightScene(c, "温馨", 50, 3000)
	lightScene(c, "冬天", 80, 5500)
	lightScene(c, "夏天", 80, 3500)
	lightScene(c, "音乐", 60, 3200)
	lightScene(c, "娱乐", 80, 4000)
}

func lightScene(c *ava.Context, simpleName string, brightness float64, kelvin int) {

	var entities = data.GetEntityAreaMap()

	if len(entities) == 0 {
		return
	}

	// 定义卧室限定场景
	bedroomScenes := map[string]bool{
		"日光": true,
		"温馨": true,
		"阅读": true,
		"休息": true,
	}

	for areaId, v := range entities {
		areaName := data.SpiltAreaName(data.GetAreaName(areaId))

		// 检查是否是客厅区域或者卧室区域
		isLivingRoom := strings.Contains(areaName, "客厅")
		isBedroom := strings.Contains(areaName, "卧室")

		// 如果不是客厅也不是卧室，则跳过
		if !isLivingRoom && !isBedroom {
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

		// 如果是卧室区域，检查是否是限定的场景
		if isBedroom {
			if !bedroomScenes[simpleName] {
				continue
			}
		}

		// 为每个区域创建独立的en和meta map
		var script = &Script{}
		var automation = &Automation{}

		//条件：名字中带有"回家"的开关按键和场景按键
		func() {
			for bName, v := range switchSelectSameName {
				bns := strings.Split(bName, "_")
				if len(bns) != 2 {
					continue
				}
				buttonName := bns[1]
				if strings.Contains(buttonName, simpleName) {
					//按键触发和条件
					for _, e := range v {
						automation.Triggers = append(automation.Triggers, Triggers{
							EntityID: e.EntityID,
							Trigger:  "state",
						})

						if e.Category == data.CategorySwitchClickOnce {
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

		var actions []interface{}
		var parallel2 = make(map[string][]interface{})

		// 添加设备操作到场景中
		for _, e1 := range v {
			if e1.Category == data.CategoryLightGroup {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct:   brightness,
						ColorTempKelvin: kelvin,
					},
					Target: &targetLightData{DeviceId: e1.DeviceID},
				})
			}

			if strings.HasPrefix(e1.EntityID, "light.") && e1.Category == data.CategoryXinGuang && !strings.Contains(e1.DeviceName, "主机") {
				//改为静态模式,不能并行执行，必须优先执行
				actions = append(actions, &ActionLight{
					DeviceID: e1.DeviceID,
					Domain:   "select",
					EntityID: data.GetXinGuang(e1.DeviceID),
					Type:     "select_option",
					Option:   "静态模式",
				})

				//修改颜色
				parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
					Action: "light.turn_on",
					Data: &actionLightData{
						BrightnessPct: brightness,
						RgbColor:      GetRgbColor(kelvin),
					},
					Target: &targetLightData{DeviceId: e1.DeviceID},
				})
				continue
			}

			if e1.Category == data.CategoryLight {
				if strings.Contains(e1.DeviceName, "彩") {
					script.Sequence = append(script.Sequence, ActionLight{
						Action: "light.turn_on",
						Data: &actionLightData{
							BrightnessPct: brightness,
						},
						Target: &targetLightData{DeviceId: e1.DeviceID},
					})
				}

				if strings.Contains(e1.DeviceName, "夜灯") {
					script.Sequence = append(script.Sequence, ActionLight{
						Action: "light.turn_on",
						Data: &actionLightData{
							BrightnessPct:   brightness,
							ColorTempKelvin: kelvin,
						},
						Target: &targetLightData{DeviceId: e1.DeviceID},
					})
				}
			}
		}

		if len(actions) > 0 {
			script.Sequence = append(script.Sequence, actions)
		}

		if len(parallel2) > 0 {
			script.Sequence = append(script.Sequence, parallel2)
		}

		if len(script.Sequence) > 0 {
			script.Alias = areaName + simpleName + "场景"
			script.Description = "点击开关按键执行" + areaName + simpleName + "场景"
			scriptId := CreateScript(c, script)

			if scriptId != "" && len(automation.Triggers) > 0 {
				automation.Alias = areaName + simpleName + "自动化"
				automation.Description = "点击开关按键执行" + areaName + simpleName + "自动化"
				automation.Mode = "single"

				automation.Actions = append(automation.Actions, ActionService{
					Action: "script.execute",
					Data:   map[string]interface{}{"script_id": scriptId},
					Target: nil,
				})
				CreateAutomation(c, automation)
			}
		}
	}
}
