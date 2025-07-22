package automation

import (
	"errors"
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

// -符号是人体感应专属
// 遍历所有人体传感器，生成自动化
// 如果相同区域有多个前缀相同的传感器触发器，则要互斥
// 开灯30秒后自动关灯
func walkBodySensor(c *ava.Context) {
	// 查询所有实体，找到名字中带有'-'的实体
	allEntities := core.GetEntityIdMap()
	var sensors []*core.Entity
	for _, e := range allEntities {
		if strings.Contains(e.Name, "-") && (e.Category == core.CategoryLight || e.Category == core.CategoryHumanBodySensor) {
			sensors = append(sensors, e)
		}
	}

	for _, v := range sensors {

		autoOn, err := bodySensorOn(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", core.MustMarshal2String(v), err)
			continue
		}

		if autoOn != nil {
			CreateAutomation(c, autoOn)
		}
	}
}

// 人体传感器有人时自动开灯/开关
func bodySensorOn(entity *core.Entity) (*Automation, error) {

	if strings.HasPrefix(entity.EntityID, "event.") && !strings.Contains(entity.OriginalName, "有人") {
		return nil, nil
	}

	areaID := entity.AreaID
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.Name
	suffix := ""
	if idx := strings.Index(prefix, "-"); idx > 0 {
		suffix = prefix[idx+1:]
		prefix = prefix[:idx]
	}

	var (
		actions            []interface{}
		atmosphereSwitches []*core.Entity
		normalSwitches     []*core.Entity
		atmosphereLights   []*core.Entity
		normalLights       []*core.Entity
	)

	//优化逻辑：除了卧室只开夜灯，其他区域打开所有灯
	for _, e := range entities {
		if e.Category == core.CategoryWiredSwitch && strings.Contains(e.OriginalName, prefix) {
			if strings.Contains(e.Name, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}
		if e.Category == core.CategoryLight && strings.Contains(e.Name, prefix) {
			if strings.Contains(e.Name, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
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

		if strings.Contains(e.Name, "彩") {
			act.Data = &actionLightData{}
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
		parallel1["parallel"] = append(parallel1["parallel"], &ActionSwitch{
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
	}

	// 3. 延迟3秒
	if len(atmosphereLights) > 0 || len(atmosphereSwitches) > 0 {
		actions = append(actions, ActionTimerDelay{Delay: struct {
			Hours        int `json:"hours"`
			Minutes      int `json:"minutes"`
			Seconds      int `json:"seconds"`
			Milliseconds int `json:"milliseconds"`
		}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
	}

	var parallel2 = make(map[string][]interface{})
	// 4. 再开非氛围灯
	for _, e := range normalLights {
		if e.EntityID == entity.EntityID {
			continue
		}
		// 夜灯/护眼灯特殊逻辑
		if strings.Contains(e.Name, "夜灯") || strings.Contains(e.Name, "护眼") {
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
		parallel2["parallel"] = append(parallel2["parallel"], &ActionSwitch{
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

	// 30秒后关灯
	actions = append(actions, ActionTimerDelay{Delay: struct {
		Hours        int `json:"hours"`
		Minutes      int `json:"minutes"`
		Seconds      int `json:"seconds"`
		Milliseconds int `json:"milliseconds"`
	}{Hours: 0, Minutes: 0, Seconds: 30, Milliseconds: 0}})
	var parallel3 = make(map[string][]interface{})

	// 关灯逻辑，先关非氛围灯、非氛围开关，延迟3秒再关氛围灯、氛围开关
	for _, e := range normalLights {
		parallel3["parallel"] = append(parallel3["parallel"], &ActionLight{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "light",
		})
	}
	for _, e := range normalSwitches {
		parallel3["parallel"] = append(parallel3["parallel"], &ActionSwitch{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
	}
	if len(atmosphereLights) > 0 || len(atmosphereSwitches) > 0 {
		parallel3["parallel"] = append(parallel3["parallel"], ActionTimerDelay{Delay: struct {
			Hours        int `json:"hours"`
			Minutes      int `json:"minutes"`
			Seconds      int `json:"seconds"`
			Milliseconds int `json:"milliseconds"`
		}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
	}
	for _, e := range atmosphereLights {
		parallel3["parallel"] = append(parallel3["parallel"], &ActionLight{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "light",
		})
	}
	for _, e := range atmosphereSwitches {
		parallel3["parallel"] = append(parallel3["parallel"], &ActionSwitch{
			Type:     "turn_off",
			DeviceID: e.DeviceID,
			EntityID: e.EntityID,
			Domain:   "switch",
		})
	}

	if len(parallel3) > 0 {
		actions = append(actions, parallel3)
	}

	areaName := core.SpiltAreaName(entity.AreaName)
	sensorPrefixStr := prefix

	//todo 需要买一个人体传感器才能测试occupied
	triggerType := "occupied"
	triggerDomain := "binary_sensor"
	triggerTrigger := "device"
	triggerDeviceId := entity.DeviceID
	if strings.HasPrefix(entity.EntityID, "event.") {
		triggerType = ""
		triggerDomain = ""
		triggerDeviceId = ""
		triggerTrigger = "state"
	}

	if entity.Category == core.CategoryLight {
		triggerType = "turned_on"
		triggerDomain = "light"
	}

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	suffixStr := suffix
	if suffixStr != "" {
		suffixStr = strings.TrimSpace(suffixStr)
	}

	auto := &Automation{
		Alias:       areaName + prefix + suffixStr + "人来亮灯",
		Description: "当检测到有人，自动打开" + areaName + "下同名前缀的灯和开关",
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

	// 增加光照条件
	lxConfig := getLxConfig(areaID)
	if lxConfig != nil {
		auto.Conditions = append(auto.Conditions, Conditions{
			Condition: "numeric_state",
			EntityID:  lxConfig.EntityId,
			Below:     lxConfig.Lx, // 设置光照阈值
		})
	}
	auto.Conditions = append(auto.Conditions, condition...)

	return auto, nil
}
