package scene

import (
	"hahub/hub/core"
	"hahub/hub/intelligent/automation"
	"strings"

	"github.com/aiakit/ava"
)

func lightSceneSetting(c *ava.Context) {
	lightScene(c, "会客", 80, 6000)
	lightScene(c, "观影", 20, 3800)
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

func lightScene(c *ava.Context, simpleName string, brightness, kelvin int) {

	var entities = core.GetEntityAreaMap()

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
		areaName := core.SpiltAreaName(core.GetAreaName(areaId))

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
			if e.Category == core.CategoryLightGroup {
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
		var en = make(map[string]interface{})
		var meta = make(map[string]interface{})
		var switchEntityId string
		var scene = &Scene{}

		//先找开关按键名称
		for _, e := range v {
			//判断当前区域是否有开关命名中带有simpleName的
			if e.Category == core.CategorySwitchScene {
				if strings.Contains(e.OriginalName, simpleName) {
					switchEntityId = e.EntityID
					break
				}

			}
		}

		for _, e1 := range v {
			if e1.Category == core.CategoryLightGroup {
				en[e1.EntityID] = map[string]interface{}{
					"state":             "on",
					"brightness":        core.ConvertBrightnessToPercentage(brightness),
					"color_temp_kelvin": kelvin,
					"friendly_name":     e1.OriginalName,
				}
				meta[e1.EntityID] = map[string]interface{}{
					"entity_only": true,
				}
			}

			if e1.Category == core.CategoryLight {
				if strings.Contains(e1.DeviceName, "彩") {
					en[e1.EntityID] = map[string]interface{}{
						"state":         "on",
						"brightness":    core.ConvertBrightnessToPercentage(brightness),
						"friendly_name": e1.OriginalName,
					}
					meta[e1.EntityID] = map[string]interface{}{
						"entity_only": true,
					}
				}

				if strings.Contains(e1.DeviceName, "夜灯") {
					en[e1.EntityID] = map[string]interface{}{
						"state":             "on",
						"brightness":        core.ConvertBrightnessToPercentage(brightness),
						"color_temp_kelvin": kelvin,
						"friendly_name":     e1.OriginalName,
					}
					meta[e1.EntityID] = map[string]interface{}{
						"entity_only": true,
					}
				}

			}
		}

		if len(en) > 0 {
			scene.Name = areaName + simpleName + "场景"
			scene.Entities = en
			scene.Metadata = meta
			sceneId := CreateScene(c, scene)

			if sceneId != "" && switchEntityId != "" {
				auto := &automation.Automation{
					Alias:       areaName + simpleName + "自动化",
					Description: "点击开关按键执行" + areaName + simpleName + "自动化",
					Triggers: []automation.Triggers{{
						EntityID: switchEntityId,
						Trigger:  "state",
					}},
					Mode: "single",
				}

				auto.Actions = append(auto.Actions, automation.ActionService{
					Action: "scene.apply",
					Data:   map[string]interface{}{"entities": sceneId},
					Target: nil,
				})
				automation.CreateAutomation(c, auto)
			}
		}
	}
}
