package automation

import (
	"errors"
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

// -符号是人体感应专属
// 遍历所有人体传感器，生成自动化
// 名字中带“感应-”的灯也作为感应器
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
			c.Error(err)
			continue
		}

		if autoOn != nil {
			CreateAutomation(c, autoOn, false, true)
		}

		autoOff, err := bodySensorOff(v)
		if err != nil {
			c.Error(err)
			continue
		}
		if autoOff != nil {
			CreateAutomation(c, autoOff, false, true)
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
	if idx := strings.Index(prefix, "-"); idx > 0 {
		prefix = prefix[:idx]
	}

	var actions []Actions
	for _, e := range entities {
		if (e.Category == core.CategoryLight || e.Category == core.CategoryWiredSwitch) && strings.HasPrefix(e.Name, prefix) {
			act := Actions{
				Type:     "turn_on",
				DeviceID: e.DeviceID,
				EntityID: e.EntityID,
				Domain:   "light",
			}
			if e.Category == core.CategoryLight {
				act.BrightnessPct = 100
			} else {
				act.Domain = "switch"
			}
			actions = append(actions, act)
		}
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

	auto := &Automation{
		Alias:       areaName + sensorPrefixStr + "人来亮灯",
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
	return auto, nil
}

// 人体传感器无人时自动关灯/关开关
func bodySensorOff(entity *core.Entity) (*Automation, error) {
	if strings.HasPrefix(entity.EntityID, "event.") && !strings.Contains(entity.OriginalName, "无人") {
		return nil, nil
	}

	areaID := entity.AreaID
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	prefix := entity.Name
	if idx := strings.Index(prefix, "-"); idx > 0 {
		prefix = prefix[:idx]
	}

	var actions []Actions
	for _, e := range entities {
		if (e.Category == core.CategoryLight || e.Category == core.CategoryWiredSwitch) && strings.HasPrefix(e.Name, prefix) {
			act := Actions{
				Type:     "turn_off",
				DeviceID: e.DeviceID,
				EntityID: e.EntityID,
				Domain:   "light",
			}
			if e.Category == core.CategoryWiredSwitch {
				act.Domain = "switch"
			}
			actions = append(actions, act)
		}
	}

	areaName := core.SpiltAreaName(entity.AreaName)
	sensorPrefixStr := prefix

	triggerType := "not_occupied"
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
		triggerType = "turned_off"
		triggerDomain = "light"
	}

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	auto := &Automation{
		Alias:       areaName + sensorPrefixStr + "人走关灯",
		Description: "当检测到无人，自动关闭" + areaName + "下同名前缀的灯和开关",
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
	return auto, nil
}
