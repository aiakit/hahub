package intelligent

import (
	"errors"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

//空调逻辑
//1.有人超过5分钟，且温度超过26度，启动空调，制冷，自动模式
//2.无人超过10分钟，停止空调,防止忘记关闭空调

func WalkPresenceSensorAir(c *ava.Context) {
	entity, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
	if !ok {
		return
	}

	for _, v := range entity {

		//卧室自动控制空调，应对昼夜温差大，防止感冒。
		if strings.Contains(v.AreaName, "卧室") {
			autoOn, err := presenceSensorOnAir(v)
			if err != nil {
				c.Debugf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				continue
			}

			if autoOn != nil {
				AddAutomation2Queue(c, autoOn)
			}
		}

		if strings.Contains(v.AreaName, "客厅") {
			autoOn, err := presenceSensorOnAirNotBedRoom(v)
			if err != nil {
				c.Debugf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				continue
			}

			if autoOn != nil {
				AddAutomation2Queue(c, autoOn)
			}
		}

		autoOff, err := presenceSensorOffAir(v)
		if err != nil {
			c.Debug(err)
			continue
		}
		if autoOff != nil {
			AddAutomation2Queue(c, autoOff)
		}
	}
}

func presenceSensorOnAir(entity *data.Entity) (*Automation, error) {
	auto := &Automation{
		Alias:       data.SpiltAreaName(entity.AreaName) + "自动打开空调",
		Description: "当温度大于27度，或者温度小于20度自动打开空调，应对昼夜温差变化大。",
		Triggers: []*Triggers{{
			Type:     "occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For: &For{
				Hours:   0,
				Minutes: 5,
				Seconds: 0,
			},
		}},
		Mode: "single",
	}

	//找到当前区域的温度传感器
	temp, ok := data.GetEntityCategoryMap()[data.CategoryTemperatureSensor]
	if !ok {
		return nil, errors.New("没有温度传感器")
	}

	var tempSensor *data.Entity

	for _, e := range temp {
		if e.AreaID == entity.AreaID {
			tempSensor = e
			break
		}
	}

	if tempSensor == nil {
		return nil, errors.New("没有温度传感器")
	}

	auto.Conditions = append(auto.Conditions, &Conditions{Condition: "numeric_state", EntityID: tempSensor.EntityID, Above: 28, Below: 20})

	entities, ok := data.GetEntityCategoryMap()[data.CategoryAirConditioner]
	if ok {
		for _, e := range entities {
			if e.AreaID != entity.AreaID {
				continue
			}
			//制冷
			func() {
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					Condition: "numeric_state",
					EntityId:  tempSensor.EntityID,
					Above:     28,
				})
				act.Then = append(act.Then, ActionService{
					Action: "climate.turn_on",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})

				act.Then = append(act.Then, ActionService{
					Action: "climate.set_temperature",
					Data:   map[string]interface{}{"temperature": 26},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})

				act.Then = append(act.Then, ActionService{
					Action: "climate.set_hvac_mode",
					Data:   map[string]interface{}{"hvac_mode": "cool"},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})
				auto.Actions = append(auto.Actions, act)
			}()

			//制热
			func() {
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					Condition: "numeric_state",
					EntityId:  tempSensor.EntityID, Below: 20,
				})
				act.Then = append(act.Then, ActionService{
					Action: "climate.turn_on",
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})

				act.Then = append(act.Then, ActionService{
					Action: "climate.set_temperature",
					Data:   map[string]interface{}{"temperature": 26},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})

				act.Then = append(act.Then, ActionService{
					Action: "climate.set_hvac_mode",
					Data:   map[string]interface{}{"hvac_mode": "heat"},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: e.EntityID},
				})
				auto.Actions = append(auto.Actions, act)
			}()
		}
	}

	if len(auto.Actions) > 0 {
		return auto, nil
	}

	return nil, errors.New("no device for air")
}

func presenceSensorOnAirNotBedRoom(entity *data.Entity) (*Automation, error) {
	if !strings.Contains(entity.AreaName, "客厅") {
		return nil, errors.New("当前区域不是客厅")
	}

	auto := &Automation{
		Alias:       data.SpiltAreaName(entity.AreaName) + "询问是否打开空调",
		Description: "当温度大于27度，或者温度小于20度自动打开空调，应对昼夜温差变化大。",
		Triggers: []*Triggers{{
			Type:     "occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For: &For{
				Hours:   0,
				Minutes: 5,
				Seconds: 0,
			},
		}},
		Mode: "single",
	}

	//找到当前区域的温度传感器
	temp, ok := data.GetEntityCategoryMap()[data.CategoryTemperatureSensor]
	if !ok {
		return nil, errors.New("没有温度传感器")
	}

	var tempSensor *data.Entity

	for _, e := range temp {
		if e.AreaID == entity.AreaID {
			tempSensor = e
			break
		}
	}

	if tempSensor == nil {
		return nil, errors.New("没有温度传感器")
	}

	auto.Conditions = append(auto.Conditions, &Conditions{Condition: "numeric_state", EntityID: tempSensor.EntityID, Above: 28, Below: 20})

	entitiesSpeaker, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if !ok {
		return nil, errors.New("没有音箱设备")
	}

	var speaderPlayTextEntityId string
	for _, v := range entitiesSpeaker {
		if strings.Contains(v.AreaName, "客厅") && strings.Contains(v.EntityID, "_play_text") && strings.HasPrefix(v.EntityID, "notify.") {
			speaderPlayTextEntityId = v.EntityID
			break
		}
	}

	entities, ok := data.GetEntityCategoryMap()[data.CategoryAirConditioner]
	if ok {
		for _, e := range entities {
			if e.AreaID != entity.AreaID {
				continue
			}
			//制冷
			func() {
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					Condition: "numeric_state",
					EntityId:  tempSensor.EntityID,
					Above:     28,
				})
				act.Then = append(act.Then, ActionService{
					Action: "notify.send_message",
					Data:   map[string]interface{}{"message": "亲爱的宿主，检测到客厅温度过低，是否打开空调"},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: speaderPlayTextEntityId},
				})

				auto.Actions = append(auto.Actions, act)
			}()

			//制热
			func() {
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					Condition: "numeric_state",
					EntityId:  tempSensor.EntityID, Below: 20,
				})
				act.Then = append(act.Then, ActionService{
					Action: "notify.send_message",
					Data:   map[string]interface{}{"message": "亲爱的宿主，检测到客厅温度过高，是否打开空调"},
					Target: &struct {
						EntityId string `json:"entity_id"`
					}{EntityId: speaderPlayTextEntityId},
				})
				auto.Actions = append(auto.Actions, act)
			}()
		}
	}

	if len(auto.Actions) > 0 {
		return auto, nil
	}

	return nil, errors.New("no device for air")
}

func presenceSensorOffAir(entity *data.Entity) (*Automation, error) {
	auto := &Automation{
		Alias:       data.SpiltAreaName(entity.AreaName) + "自动关闭空调",
		Description: "无人超过30分钟自动关闭空调",
		Triggers: []*Triggers{{
			Type:     "not_occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For: &For{
				Hours:   0,
				Minutes: 30,
				Seconds: 0,
			},
		}},
		Mode: "single",
	}

	entities, ok := data.GetEntityCategoryMap()[data.CategoryAirConditioner]
	if ok {
		for _, e := range entities {
			if e.AreaID != entity.AreaID {
				continue
			}

			var condition = Conditions{
				Condition: "state",
				EntityID:  e.EntityID,
				State:     "off",
				Attribute: "hvac_action",
			}

			auto.Conditions = append(
				auto.Conditions,
				&Conditions{Condition: "not", ConditionChild: []interface{}{condition}},
			)
			auto.Actions = append(auto.Actions, ActionService{
				Action: "climate.turn_off",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})
		}
	}

	if len(auto.Actions) > 0 {
		return auto, nil
	}

	return nil, errors.New("no device for air")
}
