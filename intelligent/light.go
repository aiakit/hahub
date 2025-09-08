package intelligent

import (
	"hahub/data"
	"strings"
)

func spiltCondition(src *data.Entity, entities []*ActionLight) (condition []*Conditions, action []interface{}) {
	for _, e := range entities {
		if e.Condition != "" {
			if src != nil && e.EntityID == src.EntityID {
				continue
			}
			condition = append(condition, &Conditions{
				Condition: "device",
				Type:      "is_off",
				DeviceID:  e.DeviceID,
				EntityID:  e.EntityID,
				Domain:    e.Domain,
			})
		} else {
			action = append(action, e)
		}
	}

	return
}

func turnOnLight(e *data.Entity, brightnessPct float64, kelvin int, openCondition bool) []*ActionLight {
	if e.Category != data.CategoryLight && e.Category != data.CategoryLightGroup {
		return nil
	}

	if strings.Contains(e.DeviceName, "浴霸") {
		return nil
	}
	var result = make([]*ActionLight, 0, 2)

	//只打开开关
	if e.SubCategory == data.CategoryWiredSwitch {
		result = append(result, &ActionLight{
			Action: "switch.turn_on",
			Target: &targetLightData{EntityId: e.EntityID, DeviceId: e.DeviceID},
		})
		if openCondition {
			result = append(result, &ActionLight{
				Condition: "device",
				Type:      "is_off",
				DeviceID:  e.DeviceID,
				EntityID:  e.EntityID,
				Domain:    "switch",
			})
		}

		return result
	}

	var rgbTmp []int

	if strings.Contains(e.DeviceName, "氛围") {
		kelvin = 3000
	}

	if strings.Contains(e.DeviceName, "夜灯") {
		brightnessPct = 5
	}

	if e.SubCategory == data.CategoryLightRgb || e.SubCategory == data.CategoryLightRgbAndTemp {
		// 处理彩色灯但不包含"彩"字的情况
		if !strings.Contains(e.DeviceName, "彩") {
			rgbTmp = GetRgbColor(kelvin)
		}
	}

	var action = &ActionLight{
		Action: "light.turn_on",
		Data: &actionLightData{
			BrightnessStepPct: brightnessPct,
		},
		Target: &targetLightData{EntityId: e.EntityID, DeviceId: e.DeviceID},

		subCategory: e.SubCategory,
	}

	if len(rgbTmp) > 0 {
		action.Data.RgbColor = rgbTmp
	} else if !strings.Contains(e.DeviceName, "彩") {
		action.Data.ColorTempKelvin = kelvin
	}

	result = append(result, action)
	if openCondition {
		result = append(result, &ActionLight{
			Condition: "device",
			Type:      "is_off",
			DeviceID:  e.DeviceID,
			EntityID:  e.EntityID,
			Domain:    "light",
		})
	}

	return result
}

// 只打开灯组和特殊单灯，如果没有灯组就打开所有灯
func findLightsWithOutLightCategory(prefix string, entities []*data.Entity) []*data.Entity {
	var result = make([]*data.Entity, 0, 4)
	for _, e := range entities {
		if prefix != "" && !strings.Contains(e.DeviceName, prefix) {
			continue
		}

		if e.Category == data.CategoryLightGroup {
			result = append(result, e)
			continue
		}

		if e.Category == data.CategoryLight && (strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜")) {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		for _, e := range entities {
			if prefix != "" && !strings.Contains(e.DeviceName, prefix) {
				continue
			}

			if e.Category == data.CategoryLightGroup || e.Category == data.CategoryLight {
				result = append(result, e)
			}
		}
	}

	return result
}

func turnOnLights(entities []*data.Entity, brightnessPct float64, kelvin int, openCondition bool) []*ActionLight {
	if len(entities) == 0 {
		return nil
	}

	var actions = make([]*ActionLight, 0, 4)
	var priorityParallel = make([]*ActionLight, 0, 4)
	var otherParallel = make([]*ActionLight, 0, 4)

	var xinguangArea = make(map[string]bool)
	//先找出所有的设备名称中带有"氛围"和"馨光"的灯，优先打开
	for _, e := range entities {

		if e.Category == data.CategoryLightGroup && strings.Contains(e.DeviceName, "馨光") {
			xinguangArea[e.AreaID] = true
			continue
		}

		if e.Category != data.CategoryLight {
			continue
		}

		// 对于馨光灯带非灯组，需要先设置为静态模式
		if strings.Contains(e.DeviceName, "馨光") {
			e1, ok := data.GetEntitiesByDeviceId()[e.DeviceID]
			if !ok {
				continue
			}
			for _, e2 := range e1 {
				a := setXinGuangDeviceToStaticWithout(e2)
				if a != nil {
					actions = append(actions, a)
				}
			}
		}
	}
	//处理灯组没有单灯的情况
	if len(xinguangArea) > 0 {
		lights, ok := data.GetEntityCategoryMap()[data.CategoryLight]
		if ok {
			for _, e := range lights {
				if e.Category == data.CategoryLight && strings.Contains(e.DeviceName, "馨光") {
					if _, ok := xinguangArea[e.AreaID]; ok {
						// 对于馨光设备，需要先设置为静态模式
						e1, ok := data.GetEntitiesByDeviceId()[e.DeviceID]
						if !ok {
							continue
						}
						for _, e2 := range e1 {
							a := setXinGuangDeviceToStaticWithout(e2)
							if a != nil {
								actions = append(actions, a)
							}
						}
					}
				}
			}
		}
	}

	if len(actions) > 0 {
		actions = append(actions, &ActionLight{
			Delay: &delay{
				Hours:        0,
				Minutes:      0,
				Seconds:      2,
				Milliseconds: 0,
			},
		})
	}

	for _, e := range entities {
		if e.Category != data.CategoryLightGroup && e.Category != data.CategoryLight {
			continue
		}

		if strings.Contains(e.DeviceName, "氛围") {
			// 添加到优先处理列表
			lightAction := turnOnLight(e, brightnessPct, kelvin, openCondition)
			if lightAction != nil {
				priorityParallel = append(priorityParallel, lightAction...)
			}
		}
	}

	// 处理其他非优先设备
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "氛围") {
			lightAction := turnOnLight(e, brightnessPct, kelvin, openCondition)
			if lightAction != nil {
				otherParallel = append(otherParallel, lightAction...)
			}
		}
	}

	// 先添加优先设备的并行操作
	if len(priorityParallel) > 0 {
		actions = append(actions, priorityParallel...)
	}

	// 再添加其他设备的并行操作
	if len(otherParallel) > 0 {
		actions = append(actions, otherParallel...)
	}

	return actions
}

func turnOffLight(e *data.Entity) *ActionLight {
	if e.Category != data.CategoryLight && e.Category != data.CategoryLightGroup {
		return nil
	}

	if e.SubCategory == data.CategoryWiredSwitch {
		return &ActionLight{
			Action: "switch.turn_off",
			Target: &targetLightData{EntityId: e.EntityID, DeviceId: e.DeviceID},
		}
	}

	return &ActionLight{
		Action: "light.turn_off",
		Target: &targetLightData{EntityId: e.EntityID, DeviceId: e.DeviceID},
	}
}

func turnOffLights(entities []*data.Entity) []*ActionLight {
	var actions = make([]*ActionLight, 0, 4)
	var priorityParallel = make([]*ActionLight, 0, 4)
	var otherParallel = make([]*ActionLight, 0, 4)

	for _, e := range entities {
		if e.Category != data.CategoryLightGroup && e.Category != data.CategoryLight {
			continue
		}
		lightAction := turnOffLight(e)

		if strings.Contains(e.DeviceName, "氛围") {
			if lightAction != nil {
				priorityParallel = append(priorityParallel, lightAction)
			}
		} else {
			otherParallel = append(otherParallel, lightAction)
		}
	}

	if len(otherParallel) > 0 {
		actions = append(actions, otherParallel...)
	}

	if len(otherParallel) > 0 && len(priorityParallel) > 0 {
		actions = append(actions, &ActionLight{
			Delay: &delay{Seconds: 5},
		})
	}

	if len(priorityParallel) > 0 {
		actions = append(actions, priorityParallel...)
	}

	return actions
}
