package intelligent

import (
	"errors"
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// -符号是人体感应专属
// 遍历所有人体传感器，生成自动化
// 如果相同区域有多个前缀相同的传感器触发器，则要互斥
// 开灯30秒后自动关灯
func WalkBodySocketSensor(c *ava.Context) {
	// 查询所有实体，找到名字中带有'-'的实体
	allEntities, ok := data.GetEntityCategoryMap()[data.CategorySocket]
	if !ok {
		return
	}

	var sensors []*data.Entity
	for _, e := range allEntities {
		if strings.Contains(e.DeviceName, "-") {
			sensors = append(sensors, e)
		}
	}

	for _, v := range sensors {

		autoOn, autoOff, err := bodySocketSensorOn(v)
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
func bodySocketSensorOn(entity *data.Entity) (*Automation, *Automation, error) {

	var (
		areaID   = entity.AreaID
		areaName = data.SpiltAreaName(entity.AreaName)
	)

	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, nil, errors.New("entity area not found")
	}

	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.DeviceName
	suffix := ""
	if idx := strings.Index(prefix, "-"); idx > 0 {
		suffix = prefix[idx+1:]
		prefix = prefix[:idx]
	}

	entitiesFilter := findLightsWithOutLightCategory(prefix, entities)
	if len(entitiesFilter) == 0 {
		return nil, nil, fmt.Errorf("%s区域没有发现灯", areaName)
	}

	actions := turnOnLights(entitiesFilter, 100, 4800, true)

	if len(actions) == 0 {
		return nil, nil, fmt.Errorf("%s区域没有发现灯", areaName)
	}

	condition, action := spiltCondition(entity, actions)

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
		Triggers: []*Triggers{{
			Type:     triggerType,
			DeviceID: triggerDeviceId,
			EntityID: entity.EntityID,
			Domain:   triggerDomain,
			Trigger:  triggerTrigger,
		}},
		Actions: action,
		Mode:    "single",
	}

	// 插座是关闭的
	auto.Conditions = append(auto.Conditions, &Conditions{
		Type:      "is_off",
		EntityID:  entity.EntityID,
		DeviceID:  entity.DeviceID,
		Domain:    "switch",
		Condition: "device",
	})
	auto.Conditions = append(auto.Conditions, condition...)

	au, err := presenceSensorOff(areaName, entity, entitiesFilter)
	if err != nil {
		ava.Error(err)
		return nil, nil, err
	}

	return auto, au, nil
}

func bodySocketSensorOff(prefix, suffix, areaName string, entity *data.Entity, entities []*data.Entity) (*Automation, error) {

	actions := turnOffLights(entities)
	if len(actions) == 0 {
		return nil, fmt.Errorf("%s区域没有设备", areaName)
	}

	var result = make([]interface{}, 0, 2)
	for _, e := range actions {
		result = append(result, e)
	}

	sensorPrefixStr := prefix

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	suffixStr := suffix
	if suffixStr != "" {
		suffixStr = strings.TrimSpace(suffixStr)
	}

	auto := &Automation{
		Alias:       areaName + "插座关闭就关灯",
		Description: "当插座关闭时，关闭" + areaName + "下同名前缀的灯和开关",
		Triggers: []*Triggers{{
			Type:     "power",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "sensor",
			Trigger:  "device",
			Below:    10,
		}},
		Actions: result,
		Mode:    "single",
	}

	auto.Actions = append(auto.Actions, delay{
		Hours:        0,
		Minutes:      0,
		Seconds:      5,
		Milliseconds: 0,
	})

	auto.Actions = append(auto.Actions, ActionCommon{
		Type:     "turn_off",
		DeviceID: entity.DeviceID,
		EntityID: entity.EntityID,
		Domain:   "switch",
	})

	return auto, nil
}
