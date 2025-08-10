package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

// 洗澡，楼层，热水器开关，温度检测，浴霸
func TakeAShower(c *ava.Context) {

	//判断是否有热水器
	v, ok := data.GetEntityCategoryMap()[data.CategoryWaterHeater]
	//判断温度，随便找到家里的一个温度设备作为判断标准
	vv, ok1 := data.GetEntityCategoryMap()[data.CategoryTemperatureSensor]
	var action []interface{}
	var alias string

	if ok {
		//打开热水器开关，开启零冷水
		for _, e := range v {
			if strings.Contains(e.OriginalName, "开关") {
				alias = data.SpiltAreaName(e.AreaID) + e.DeviceName + "热水"
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					State:     "off",
					EntityId:  e.EntityID,
					Condition: "state",
				})
				act.Then = append(act.Then, ActionCommon{
					Type:     "turn_on",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
				action = append(action, act)
			}

			if strings.Contains(e.OriginalName, "零冷水") && !strings.Contains(e.OriginalName, "点动") {
				var act IfThenELSEAction
				act.If = append(act.If, ifCondition{
					State:     "off",
					EntityId:  e.EntityID,
					Condition: "state",
				})
				act.Then = append(act.Then, ActionCommon{
					Type:     "turn_on",
					DeviceID: e.DeviceID,
					EntityID: e.EntityID,
					Domain:   "switch",
				})
				action = append(action, act)
			}
		}

		for _, e := range v {
			if strings.HasPrefix(e.EntityID, "water_heater.") {
				var isExsitTemperature bool
				if ok1 {
					for _, e1 := range vv {
						if strings.HasPrefix(e1.EntityID, "sensor.") && strings.Contains(e.OriginalName, "温度") {
							isExsitTemperature = true
							var act IfThenELSEAction
							act.If = append(act.If, ifCondition{
								Type:      "is_temperature",
								DeviceId:  e1.DeviceID,
								EntityId:  e1.EntityID,
								Domain:    "sensor",
								Condition: "device",
								Above:     24, //冬天和夏天分割线
							})

							act.Then = append(act.Then, ActionService{
								Action: "water_heater.set_temperature",
								Data:   map[string]interface{}{"temperature": 39},
								Target: &struct {
									EntityId string `json:"entity_id"`
								}{EntityId: e.EntityID},
							})

							act.Else = append(act.Else, ActionService{
								Action: "water_heater.set_temperature",
								Data:   map[string]interface{}{"temperature": 42},
								Target: &struct {
									EntityId string `json:"entity_id"`
								}{EntityId: e.EntityID},
							})

							action = append(action, act)

							break
						}
					}
				}

				if !isExsitTemperature {
					action = append(action, ActionService{
						Action: "water_heater.set_temperature",
						Data:   map[string]interface{}{"temperature": 39},
						Target: &struct {
							EntityId string `json:"entity_id"`
						}{EntityId: e.EntityID},
					})
				}
			}
		}
	}

	var script = &Script{Sequence: make([]interface{}, 0, 2)}

	if len(action) > 0 {
		script.Sequence = append(script.Sequence, action...)
	}

	var isExsit bool
	//判断是否有浴霸
	var act IfThenELSEAction
	for _, e1 := range vv {
		if strings.HasPrefix(e1.EntityID, "sensor.") && strings.Contains(e.OriginalName, "温度") {
			act.If = append(act.If, ifCondition{
				Type:      "is_temperature",
				DeviceId:  e1.DeviceID,
				EntityId:  e1.EntityID,
				Domain:    "sensor",
				Condition: "device",
				Above:     24, //冬天和夏天分割线
			})
		}
	}

	//浴霸
	for _, e := range data.GetEntityCategoryMap()[data.CateroyBathroomHeater] {
		isExsit = true
		alia := alias + e.DeviceName + "洗澡场景"
		if strings.HasPrefix(e.EntityID, "climate.") {
			act.Else = append(act.Else, ActionService{
				Action: "climate.set_temperature",
				Data:   map[string]interface{}{"temperature": 40},
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})

			act.Else = append(act.Else, ActionService{
				Action: "climate.turn_on",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "换气低档") {
			act.Then = append(act.Then, ActionCommon{
				Type:     "press",
				DeviceID: e.DeviceID,
				EntityID: e.EntityID,
				Domain:   "button",
			})
		}
		script.Alias = alia
		script.Description = data.SpiltAreaName(e.AreaName) + alias
		CreateScript(c, script)
	}

	if !isExsit {
		CreateScript(c, script)
	}
}
