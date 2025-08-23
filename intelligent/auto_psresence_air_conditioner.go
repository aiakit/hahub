package intelligent

import (
	"errors"
	"hahub/data"
	"hahub/x"

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

		autoOn, err := presenceSensorOnAir(v)
		if err != nil {
			c.Debugf("entity=%s |err=%v", x.MustMarshal2String(v), err)
			continue
		}

		AddAutomation2Queue(c, autoOn)

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
		Description: "当温度大于27度，或者温度小于20度自动打开空调",
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
			auto.Conditions = append(auto.Conditions, &Conditions{
				Condition: "numeric_state",
				EntityID:  e.EntityID,
				Above:     27,
				Below:     20,
			})
			break
		}
	}

	if len(auto.Conditions) == 0 || tempSensor == nil {
		return nil, errors.New("没有温度传感器")
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
					Above:     27,
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
			var act IfThenELSEAction
			act.If = append(act.If, ifCondition{
				Condition: "state",
				State:     "off",
				EntityId:  e.EntityID,
				Attribute: "hvac_action",
			})
			act.Else = append(act.Else, ActionService{
				Action: "climate.turn_off",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})

			auto.Actions = append(auto.Actions, act)
		}
	}

	if len(auto.Actions) > 0 {
		return auto, nil
	}

	return nil, errors.New("no device for air")
}
