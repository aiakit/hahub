package intelligent

import (
	"hahub/data"
	"strings"
)

// 使用红外控制控制电视，使用电视接入ha的插件判断电视状态,红外电视名称和ha插件电视名称必须一致
func turnOffTv(areaId string) []interface{} {
	var result = make([]interface{}, 0, 2)

	areanEntities, ok := data.GetEntityAreaMap()[areaId]
	if !ok {
		return result
	}

	var okIr, okHa bool
	for _, e := range areanEntities {
		if e.Category == data.CategoryIrTV {
			okIr = true
		}

		if e.Category == data.CategoryHaTV {
			okHa = true
		}
	}

	var act *IfThenELSEAction
	var sameEntity = make(map[string]bool)

	// 同时存在红外和HA电视设备的情况
	if okIr && okHa {
		for _, e := range areanEntities {
			// 匹配相同名称的设备
			if e.Category == data.CategoryIrTV && strings.Contains(e.OriginalName, "关机") {
				for _, ha := range areanEntities {
					if ha.Category == data.CategoryHaTV && ha.Name == e.DeviceName {
						// 查找红外电视的关机按钮
						act = new(IfThenELSEAction)
						act.If = append(act.If, ifCondition{
							Condition: "state",
							State:     "on",
							EntityId:  ha.EntityID,
						})
						act.Then = append(act.Then, ActionCommon{
							Type:     "press",
							DeviceID: e.DeviceID,
							EntityID: e.EntityID,
							Domain:   "button",
						})
						sameEntity[ha.EntityID] = true
						sameEntity[e.EntityID] = true

						result = append(result, act)
					}
				}
			}
		}
	}

	for _, e := range areanEntities {
		if sameEntity[e.EntityID] {
			continue
		}

		if e.Category == data.CategoryHaTV {
			act = new(IfThenELSEAction)
			act.If = append(act.If, ifCondition{
				Condition: "state",
				State:     "on",
				EntityId:  e.EntityID,
			})
			act.Then = append(act.Then, ActionService{
				Action: "media_player.turn_off",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})
			result = append(result, act)
		}
	}

	// 只有红外电视设备的情况
	for _, e := range areanEntities {
		if sameEntity[e.EntityID] {
			continue
		}

		if e.Category == data.CategoryIrTV && strings.Contains(e.OriginalName, "关机") {
			result = append(result, &ActionCommon{
				DeviceID: e.DeviceID,
				Domain:   "button",
				EntityID: e.EntityID,
				Type:     "press"})
		}
	}

	return result
}

// 使用红外控制控制电视，使用电视接入ha的插件判断电视状态,红外电视名称和ha插件电视名称必须一致
func turnOnTv(areaId string) []interface{} {
	var result = make([]interface{}, 0)

	areanEntites, ok := data.GetEntityAreaMap()[areaId]
	if !ok {
		return result
	}

	var okIr, okHa bool
	for _, e := range areanEntites {
		if e.Category == data.CategoryIrTV {
			okIr = true
		}

		if e.Category == data.CategoryHaTV {
			okHa = true
		}
	}

	var act *IfThenELSEAction
	var sameEntity = make(map[string]bool)

	// 同时存在红外和HA电视设备的情况
	if okIr && okHa {
		for _, e := range areanEntites {
			// 匹配相同名称的设备
			if e.Category == data.CategoryIrTV && strings.Contains(e.OriginalName, "开机") {
				for _, ha := range areanEntites {
					if ha.Category == data.CategoryHaTV && ha.Name == e.DeviceName {
						// 查找红外电视的开机按钮
						act = new(IfThenELSEAction)
						act.If = append(act.If, ifCondition{
							Condition: "state",
							State:     "off",
							EntityId:  ha.EntityID,
						})
						act.Then = append(act.Then, ActionCommon{
							Type:     "press",
							DeviceID: e.DeviceID,
							EntityID: e.EntityID,
							Domain:   "button",
						})
						sameEntity[ha.EntityID] = true
						sameEntity[e.EntityID] = true

						result = append(result, act)
					}
				}
			}
		}
	}

	for _, e := range areanEntites {
		if sameEntity[e.EntityID] {
			continue
		}

		if e.Category == data.CategoryHaTV {
			act = new(IfThenELSEAction)
			act.If = append(act.If, ifCondition{
				Condition: "state",
				State:     "off",
				EntityId:  e.EntityID,
			})
			act.Then = append(act.Then, ActionService{
				Action: "media_player.turn_on",
				Target: &struct {
					EntityId string `json:"entity_id"`
				}{EntityId: e.EntityID},
			})
			result = append(result, act)
		}
	}

	// 只有红外电视设备的情况
	for _, e := range areanEntites {
		if sameEntity[e.EntityID] {
			continue
		}

		if e.Category == data.CategoryIrTV && strings.Contains(e.OriginalName, "开机") {
			result = append(result, &ActionCommon{
				DeviceID: e.DeviceID,
				Domain:   "button",
				EntityID: e.EntityID,
				Type:     "press"})
		}
	}

	return result
}
