package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func InitLevingHome(c *ava.Context) {
	// 创建离家场景
	script, auto := levingHomeScript()
	if script != nil && len(script.Sequence) > 0 {
		AddScript2Queue(c, script)
	}

	if auto != nil && len(auto.Actions) > 0 && len(auto.Triggers) > 0 {
		AddAutomation2Queue(c, auto)
	}
}

// 离家场景场景
func levingHomeScript() (*Script, *Automation) {
	var script = &Script{
		Alias:       "离家场景",
		Description: "离家场景执行场景",
	}

	var action = make([]interface{}, 0)

	// 关闭所有灯
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
		if ok {
			for _, v := range entities {
				action = append(action, ActionLight{
					Action: "light.turn_off",
					Target: &targetLightData{DeviceId: v.DeviceID},
				})
			}
		}
	}()

	// 关闭插座
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategorySocket]
		if ok {
			for _, e := range entities {
				action = append(action, ActionCommon{
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
		entities, ok := data.GetEntityCategoryMap()[data.CategoryCurtain]
		if ok {
			for _, e := range entities {
				action = append(action, ActionCommon{
					Type:     "close",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "cover",
				})
			}
		}
	}()

	// 关闭电视逻辑，找到电视机
	func() {
		area := data.GetAreas()
		if len(area) > 0 {
			for _, v := range area {
				result := TurnOffTv(v)
				for _, v1 := range result {
					action = append(action, v1)
				}
			}
		}
	}()

	// 关闭空调
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategoryAirConditioner]
		if ok {
			for _, e := range entities {
				action = append(action, ActionService{
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
		entities, ok := data.GetEntityCategoryMap()[data.CategoryWaterHeater]
		if ok {
			for _, e := range entities {
				if strings.Contains(e.OriginalName, "开关") {
					action = append(action, ActionService{
						Action: "water_heater.turn_off",
						Target: &struct {
							EntityId string `json:"entity_id"`
						}{EntityId: e.EntityID},
					})
				}
			}
		}
	}()

	//// 播放离家提醒（通过音箱）
	//func() {
	//	entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	//	if ok {
	//		for _, e := range entities {
	//			if strings.Contains(e.OriginalName, "执行文本指令") && strings.Contains(e.AreaName, "客厅") {
	//				action = append(action, ActionNotify{
	//					Action: "notify.send_message",
	//					Data: struct {
	//						Message string `json:"message,omitempty"`
	//						Title   string `json:"title,omitempty"`
	//					}{Message: "主人即将离家，请检查是否关闭所有电器设备", Title: ""},
	//				})
	//			}
	//		}
	//	}
	//}()

	//关闭插座
	func() {
		entities, ok := data.GetEntityCategoryMap()[data.CategorySocket]
		if ok {
			for _, e := range entities {
				action = append(action, ActionCommon{
					Type:     "turn_off",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
			}
		}
	}()

	entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if ok {
		for _, e := range entities {
			if strings.HasPrefix(e.EntityID, "media_player.") {
				action = append(action, ActionService{
					Action: "media_player.media_pause",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID}})
			}
		}
	}

	// 布防
	action = append(action, ActionService{
		Action: "automation.turn_on",
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: "automation.li_jia_bu_fang"},
	})

	var act IfThenELSEAction

	//检查存在传感器是否检测到有人
	func() {
		entitiesSensor, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
		if ok {
			for _, e := range entitiesSensor {
				act.If = append(act.If, ifCondition{
					Type:      "is_not_occupied",
					DeviceId:  e.DeviceID,
					EntityId:  e.EntityID,
					Condition: "device",
					Domain:    "binary_sensor",
					For: &For{
						Hours:   0,
						Minutes: 5,
						Seconds: 0,
					},
				})
			}
		}
	}()

	if len(act.If) > 0 && len(action) > 0 {
		act.Then = append(act.Then, action...)
		script.Sequence = append(script.Sequence, act)
	} else if len(action) > 0 {
		script.Sequence = append(script.Sequence, action...)
	}

	if len(action) > 0 {
		auto := levingHomeAutomation(action)
		if auto != nil {
			auto.Actions = action
		}
		return script, auto
	}

	return nil, nil
}

// 离家自动化
func levingHomeAutomation(action []interface{}) *Automation {
	var automation = &Automation{
		Alias:       "离家自动化",
		Description: "门锁关闭/或者开关按键触发用或条件，判断是否所有设备已关闭并启动安防",
		Mode:        "single",
	}

	var condition = make([]*Conditions, 0)

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
				var con = &Conditions{
					Condition: "or",
				}

				for _, e := range v {
					automation.Triggers = append(automation.Triggers, &Triggers{
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
					automation.Conditions = append(automation.Conditions, con)
				}
			}
		}
	}()

	if len(condition) > 0 {
		automation.Actions = append(automation.Actions, action...)
		var con = &Conditions{
			Condition: "or",
		}

		for _, v := range condition {
			con.ConditionChild = append(con.ConditionChild, v)
		}

		automation.Conditions = append(automation.Conditions, con)
	}

	return automation
}
