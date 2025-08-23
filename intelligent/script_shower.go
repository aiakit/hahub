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

	if ok {
		//打开热水器开关，开启零冷水
		for _, e := range v {
			if strings.Contains(e.OriginalName, "开关") {
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
						if strings.HasPrefix(e1.EntityID, "sensor.") && strings.Contains(e1.OriginalName, "温度") {
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

							//美的插件有bug
							var tem float64 = 39
							if e1.Platform == "midea_ac_lan" {
								tem *= 2
							}

							act.Then = append(act.Then, ActionService{
								Action: "water_heater.set_temperature",
								Data:   map[string]interface{}{"temperature": tem},
								Target: &struct {
									EntityId string `json:"entity_id"`
								}{EntityId: e.EntityID},
							})

							//todo 美的插件有bug
							tem = 42
							if e1.Platform == "midea_ac_lan" {
								tem *= 2
							}

							act.Else = append(act.Else, ActionService{
								Action: "water_heater.set_temperature",
								Data:   map[string]interface{}{"temperature": tem},
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

	//浴霸
	var deviceName = make(map[string]*IfThenELSEAction)
	var areaName string
	for _, e := range data.GetEntityCategoryMap()[data.CateroyBathroomHeater] {
		if strings.HasPrefix(e.EntityID, "climate.") {
			if deviceName[e.DeviceName] == nil {
				deviceName[e.DeviceName] = new(IfThenELSEAction)
			}
			areaName = data.SpiltAreaName(e.AreaName)

			deviceName[e.DeviceName].If = append(deviceName[e.DeviceName].If, ifCondition{
				Attribute: "current_temperature",
				Condition: "numeric_state",
				EntityId:  e.EntityID,
				Above:     24, //冬天和夏天分割线
			})
			deviceName[e.DeviceName].Else = append(deviceName[e.DeviceName].Else, ActionService{
				Action: "climate.set_temperature",
				Data:   map[string]interface{}{"temperature": 40},
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			}, ActionService{
				Action: "climate.turn_on",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "换气低挡") {
			if deviceName[e.DeviceName] == nil {
				deviceName[e.DeviceName] = new(IfThenELSEAction)
			}

			deviceName[e.DeviceName].Then = append(deviceName[e.DeviceName].Then, ActionCommon{
				Type:     "press",
				DeviceID: e.DeviceID,
				EntityID: e.EntityID,
				Domain:   "button",
			})
		}
	}

	if len(deviceName) > 0 {
		for k, v3 := range deviceName {
			script.Alias = areaName + k + "洗澡场景"
			script.Description = "打开热水器和浴霸，使用热水洗澡场景"
			script.Sequence = append(script.Sequence, v3)
			AddScript2Queue(c, script)
		}
	} else if len(script.Sequence) > 0 && len(deviceName) == 0 {
		script.Alias = areaName + "热水场景"
		script.Description = "打开热水器，使用热水场景"
		AddScript2Queue(c, script)
	}
}
