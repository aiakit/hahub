package automation

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

func initLevingHome(c *ava.Context) {
	// 创建离家场景
	script := levingHomeScript()
	if script != nil && len(script.Sequence) > 0 {
		scriptId := CreateScript(c, script)

		// 基于场景创建自动化
		if scriptId != "" {
			auto := levingHomeAutomation(scriptId)
			if auto != nil {
				CreateAutomation(c, auto)
			}
		}
	}
}

// 离家场景场景
func levingHomeScript() *Script {
	var script = &Script{
		Alias:       "离家场景",
		Description: "离家场景执行场景",
	}

	// 关闭所有灯
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
		if ok {
			for _, v := range entities {
				script.Sequence = append(script.Sequence, ActionLight{
					Action: "light.turn_off",
					Target: &targetLightData{DeviceId: v.DeviceID},
				})
			}
		}
	}()

	// 关闭插座
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategorySocket]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "turn_off",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
			}
		}
	}()

	// 关闭窗帘
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryCurtain]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "close",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "cover",
				})
			}
		}
	}()

	// 关闭电视
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryIrTV]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "红外电视控制") && strings.Contains(e.OriginalName, "关机") {
					script.Sequence = append(script.Sequence, ActionCommon{
						Type:     "press",
						DeviceID: e.DeviceID,
						EntityID: e.EntityID,
						Domain:   "button",
					})
				}
			}
		}
	}()

	// 关闭空调
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryAirConditioner]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionService{
					Action: "climate.turn_off",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})
			}
		}
	}()

	// 关闭热水器
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryWaterHeater]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionService{
					Action: "water_heater.turn_off",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})
			}
		}
	}()

	// 播放离家提醒（通过音箱）
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") {
					script.Sequence = append(script.Sequence, ActionNotify{
						Action: "notify.send_message",
						Data: struct {
							Message string `json:"message,omitempty"`
							Title   string `json:"title,omitempty"`
						}{Message: "主人即将离家，请检查是否关闭所有电器设备", Title: ""},
					})
				}
			}
		}
	}()

	//关闭插座
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategorySocket]
		if ok {
			for _, e := range entities {
				script.Sequence = append(script.Sequence, ActionCommon{
					Type:     "turn_off",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
			}
		}
	}()

	// 布防
	script.Sequence = append(script.Sequence, ActionService{
		Action: "automation.turn_on",
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "automation.li_jia_bu_fang"},
	})

	return script
}

// 离家自动化
func levingHomeAutomation(scriptId string) *Automation {
	var automation = &Automation{
		Alias:       "离家自动化",
		Description: "门锁关闭/或者开关按键触发用或条件，判断是否所有设备已关闭并启动安防",
		Mode:        "single",
	}

	// 条件：检查是否存在开着的灯
	func() {
		entities, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
		if ok {
			for _, e := range entities {
				automation.Conditions = append(automation.Conditions, Conditions{
					Condition: "device",
					Type:      "is_on",
					DeviceID:  e.DeviceID,
					EntityID:  e.EntityID,
					Domain:    "light",
				})
			}
		}
	}()

	// 条件：名字中带有"离家"的开关按键和场景按键
	func() {
		for bName, v := range switchSelectSameName {
			bns := strings.Split(bName, "_")
			if len(bns) < 2 {
				continue
			}
			buttonName := bns[len(bns)-1]
			if strings.Contains(buttonName, "离家") {
				//按键触发和条件
				for _, e := range v {
					automation.Triggers = append(automation.Triggers, Triggers{
						EntityID: e.EntityID,
						Trigger:  "state",
					})

					if e.Category == core.CategorySwitchClickOnce {
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

	// 执行离家场景
	automation.Actions = append(automation.Actions, ActionService{
		Action: "script.turn_on",
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "script." + scriptId},
	})

	return automation
}
