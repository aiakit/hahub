package automation

import (
	"errors"
	"hahub/hub/internal"
	"strings"

	"github.com/aiakit/ava"
)

func walkPresenceSensor(c *ava.Context) {
	entity, ok := internal.GetEntityCategoryMap()[internal.CategoryHumanPresenceSensor]
	if !ok {
		return
	}

	for _, v := range entity {
		autoOn, err := presenceSensorOn(v)
		if err != nil {
			c.Error(err)
			continue
		}

		CreateAutomation(c, autoOn, false, true)

		autoOff, err := presenceSensorOff(v)
		if err != nil {
			c.Error(err)
			continue
		}
		CreateAutomation(c, autoOff, false, true)
	}
}

// 人体存在传感器
// 人在亮灯,人走灭灯
// 1.遍历所有和人在传感器区域相同的灯
func presenceSensorOn(entity *internal.Entity) (*Automation, error) {
	var (
		areaID        = entity.AreaID
		lights        []*internal.Entity
		wiredSwitches []*internal.Entity
	)

	// 查找同区域所有实体
	entities, ok := internal.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		// 灯组:EntityID前缀light.
		if e.Category == internal.CategoryLight {
			lights = append(lights, e)
		}
		// 有线开关: 名字含“开关”和“有线”，EntityID含switch
		if e.Category == internal.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			wiredSwitches = append(wiredSwitches, e)
		}
	}

	var actions []Actions
	for _, l := range lights {
		actions = append(actions, Actions{
			Type:          "turn_on",
			DeviceID:      l.DeviceID,
			EntityID:      l.EntityID,
			Domain:        "light",
			BrightnessPct: 100,
		})
	}

	for _, s := range wiredSwitches {
		actions = append(actions, Actions{
			Type:     "turn_on",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}

	areaName := internal.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "人来亮灯",
		Description: "当人体传感器检测到有人，自动打开" + areaName + "灯组和有线开关",
		Triggers: []Triggers{{
			Type:     "occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
		}},
		Actions: actions,
		Mode:    "single",
	}

	return auto, nil
}

func presenceSensorOff(entity *internal.Entity) (*Automation, error) {
	var (
		areaID        = entity.AreaID
		lights        []*internal.Entity
		wiredSwitches []*internal.Entity
	)

	// 查找同区域所有实体
	entities, ok := internal.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		// 灯组:EntityID前缀light.
		if e.Category == internal.CategoryLight {
			lights = append(lights, e)
		}
		// 有线开关: 名字含“开关”和“有线”，EntityID含switch
		if e.Category == internal.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			wiredSwitches = append(wiredSwitches, e)
		}
	}

	var actions []Actions
	for _, l := range lights {
		actions = append(actions, Actions{
			Type:     "turn_off",
			DeviceID: l.DeviceID,
			EntityID: l.EntityID,
			Domain:   "light",
		})
	}

	for _, s := range wiredSwitches {
		actions = append(actions, Actions{
			Type:     "turn_off",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}

	areaName := internal.SpiltAreaName(entity.AreaName)

	auto := &Automation{
		Alias:       areaName + "人走关灯",
		Description: "当人体传感器检测到无人，自动关闭" + areaName + "灯组和有线开关",
		Triggers: []Triggers{{
			Type:     "not_occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
		}},
		Actions: actions,
		Mode:    "single",
	}

	return auto, nil
}
