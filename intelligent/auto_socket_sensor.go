package intelligent

import (
	"errors"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// -符号是人体感应专属
// 遍历所有人体传感器，生成自动化
// 如果相同区域有多个前缀相同的传感器触发器，则要互斥
// 开灯30秒后自动关灯
func walkBodySocketSensor(c *ava.Context) {
	// 查询所有实体，找到名字中带有'-'的实体
	allEntities := data.GetEntityIdMap()
	var sensors []*data.Entity
	for _, e := range allEntities {
		if strings.Contains(e.DeviceName, "-") && e.Category == data.CategorySocket {
			sensors = append(sensors, e)
		}
	}

	for _, v := range sensors {

		autoOn, err := bodySocketSensorOn(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
			continue
		}

		autoOff, err := bodySocketSensorOff(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
			continue
		}

		if autoOn != nil {
			CreateAutomation(c, autoOn)
		}

		if autoOff != nil {
			CreateAutomation(c, autoOff)
		}
	}
}

// 插座有人时自动开灯/开关
func bodySocketSensorOn(entity *data.Entity) (*Automation, error) {

	areaID := entity.AreaID
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.DeviceName
	suffix := ""
	if idx := strings.Index(prefix, "-"); idx > 0 {
		suffix = prefix[idx+1:]
		prefix = prefix[:idx]
	}

	var (
		actions            []interface{}
		atmosphereSwitches []*data.Entity
		normalSwitches     []*data.Entity
		atmosphereLights   []*data.Entity
		normalLights       []*data.Entity
	)

	//优化逻辑：除了卧室只开夜灯，其他区域打开所有灯
	for _, e := range entities {
		if e.Category == data.CategoryWiredSwitch && strings.Contains(e.OriginalName, prefix) {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}

		if e.Category == data.CategoryLightGroup && strings.Contains(e.DeviceName, prefix) {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
		}

		if e.Category == data.CategoryLight && (strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜灯")) {
			atmosphereLights = append(atmosphereLights, e)
		}
	}

	var condition = make([]Conditions, 0, 2)

	var parallel1 = make(map[string][]interface{})
	// 开灯逻辑
	// 1. 先开氛围灯
	for _, e := range atmosphereLights {
		act := &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: 3000,
				BrightnessPct:   100,
			},
			Target: &targetLightData{DeviceId: e.DeviceID},
		}

		if strings.Contains(e.DeviceName, "彩") {
			act.Data = &actionLightData{BrightnessStepPct: 100}
		}

		condition = append(condition, Conditions{
			Condition: "device",
			Type:      "is_off",
			DeviceID:  e.DeviceID,
			EntityID:  e.EntityID,
			Domain:    "light",
		})
		parallel1["parallel"] = append(parallel1["parallel"], act)
	}
	// 2. 先开氛围开关
	for _, e := range atmosphereSwitches {
		parallel1["parallel"] = append(parallel1["parallel"], &ActionCommon{
			Type:     "turn_on",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
		condition = append(condition, Conditions{
			Condition: "device",
			Type:      "is_off",
			DeviceID:  e.DeviceID,
			EntityID:  e.EntityID,
			Domain:    "switch",
		})
	}
	if len(parallel1) > 0 {
		actions = append(actions, parallel1)
		if len(atmosphereLights) > 0 || len(atmosphereSwitches) > 0 {
			actions = append(actions, ActionTimerDelay{Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
		}
	}

	var parallel2 = make(map[string][]interface{})
	// 4. 再开非氛围灯
	for _, e := range normalLights {
		if e.EntityID == entity.EntityID {
			continue
		}
		// 夜灯特殊逻辑
		if strings.Contains(e.DeviceName, "夜灯") {
			parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					ColorTempKelvin: 3000,
					BrightnessPct:   10,
				},
				Target: &targetLightData{DeviceId: e.DeviceID},
			})
			continue
		}
		condition = append(condition, Conditions{
			Condition: "device",
			Type:      "is_off",
			DeviceID:  e.DeviceID,
			EntityID:  e.EntityID,
			Domain:    "light",
		})

		parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: 3000,
				BrightnessPct:   100,
			},
			Target: &targetLightData{DeviceId: e.DeviceID},
		})
	}
	// 5. 再开非氛围开关
	for _, e := range normalSwitches {
		parallel2["parallel"] = append(parallel2["parallel"], &ActionCommon{
			Type:     "turn_on",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
		condition = append(condition, Conditions{
			Condition: "device",
			Type:      "is_off",
			DeviceID:  e.DeviceID,
			EntityID:  e.EntityID,
			Domain:    "switch",
		})
	}

	if len(parallel2) > 0 {
		actions = append(actions, parallel2)
	}

	if len(actions) == 0 {
		return nil, errors.New("没有设备")
	}

	areaName := data.SpiltAreaName(entity.AreaName)
	sensorPrefixStr := prefix

	triggerType := "turned_on"
	triggerDomain := "switch"
	triggerTrigger := "device"
	triggerDeviceId := entity.DeviceID

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	suffixStr := suffix
	if suffixStr != "" {
		suffixStr = strings.TrimSpace(suffixStr)
	}

	auto := &Automation{
		Alias:       areaName + prefix + suffixStr + "插座打开就开灯",
		Description: "当插座打开" + areaName + "下同名前缀的灯和开关",
		Triggers: []Triggers{{
			Type:     triggerType,
			DeviceID: triggerDeviceId,
			EntityID: entity.EntityID,
			Domain:   triggerDomain,
			Trigger:  triggerTrigger,
		}},
		Actions: actions,
		Mode:    "single",
	}
	auto.Actions = actions

	// 插座是关闭的
	auto.Conditions = append(auto.Conditions, Conditions{
		Type:      "is_off",
		EntityID:  entity.EntityID,
		DeviceID:  entity.DeviceID,
		Domain:    "switch",
		Condition: "device",
	})
	auto.Conditions = append(auto.Conditions, condition...)

	return auto, nil
}

func bodySocketSensorOff(en *data.Entity) (*Automation, error) {
	ess, ok := data.GetEntityCategoryMap()[data.CategoryPowerconsumption]
	if !ok {
		return nil, nil
	}

	var entity *data.Entity
	for _, e := range ess {
		if e.DeviceID == en.DeviceID {
			entity = e
		}
	}

	if entity == nil {
		return nil, nil
	}

	areaID := entity.AreaID
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.DeviceName
	suffix := ""
	if idx := strings.Index(prefix, "-"); idx > 0 {
		suffix = prefix[idx+1:]
		prefix = prefix[:idx]
	}

	var (
		actions            []interface{}
		atmosphereSwitches []*data.Entity
		normalSwitches     []*data.Entity
		atmosphereLights   []*data.Entity
		normalLights       []*data.Entity
	)

	//优化逻辑：除了卧室只开夜灯，其他区域打开所有灯
	for _, e := range entities {
		if e.Category == data.CategoryWiredSwitch && strings.Contains(e.OriginalName, prefix) {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}

		if e.Category == data.CategoryLightGroup && strings.Contains(e.DeviceName, prefix) {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
		}

		if e.Category == data.CategoryLight && (strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜灯")) {
			atmosphereLights = append(atmosphereLights, e)
		}
	}

	var parallel1 = make(map[string][]interface{})
	// 开灯逻辑
	// 1. 先开氛围灯
	for _, e := range atmosphereLights {
		act := &ActionLight{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "light",
		}

		parallel1["parallel"] = append(parallel1["parallel"], act)
	}
	// 2. 先开氛围开关
	for _, e := range atmosphereSwitches {
		parallel1["parallel"] = append(parallel1["parallel"], &ActionCommon{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
	}
	if len(parallel1) > 0 {
		actions = append(actions, parallel1)
		// 3. 延迟3秒
		if len(atmosphereLights) > 0 || len(atmosphereSwitches) > 0 {
			actions = append(actions, ActionTimerDelay{Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
		}
	}

	var parallel2 = make(map[string][]interface{})
	for _, e := range normalLights {
		if e.EntityID == entity.EntityID {
			continue
		}

		parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "light",
		})
	}
	for _, e := range normalSwitches {
		parallel2["parallel"] = append(parallel2["parallel"], &ActionCommon{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
	}

	if len(parallel2) > 0 {
		actions = append(actions, parallel2)
	}

	if len(actions) == 0 {
		return nil, errors.New("没有设备")
	}

	actions = append(actions, ActionCommon{
		Type:     "turn_off",
		DeviceID: en.DeviceID,
		EntityID: en.EntityID,
		Domain:   "switch",
	})

	areaName := data.SpiltAreaName(entity.AreaName)
	sensorPrefixStr := prefix

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	suffixStr := suffix
	if suffixStr != "" {
		suffixStr = strings.TrimSpace(suffixStr)
	}

	auto := &Automation{
		Alias:       areaName + prefix + suffixStr + "插座关闭就关灯",
		Description: "当插座关闭时，关闭" + areaName + "下同名前缀的灯和开关",
		Triggers: []Triggers{{
			Type:     "power",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "sensor",
			Trigger:  "device",
			Below:    10,
		}},
		Actions: actions,
		Mode:    "single",
	}
	auto.Actions = actions

	return auto, nil
}
