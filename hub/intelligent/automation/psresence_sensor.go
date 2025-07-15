package automation

import (
	"errors"
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

func walkPresenceSensor(c *ava.Context) {
	entity, ok := core.GetEntityCategoryMap()[core.CategoryHumanPresenceSensor]
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
// 2.对客厅、卧室区域，判断光照条件,时间条件,是否执行晚安场景
// 3.被主动关闭后，人来灯亮的所有自动化都实效，持续到被主动开启,晚安，起床
func presenceSensorOn(entity *core.Entity) (*Automation, error) {
	var (
		areaID        = entity.AreaID
		lights        []*core.Entity
		wiredSwitches []*core.Entity
	)

	// 查找同区域所有实体
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		// 灯组:EntityID前缀light.
		if e.Category == core.CategoryLight && !strings.Contains(e.OriginalName, "馨光") {
			lights = append(lights, e)
		}
		// 有线开关: 名字含“开关”和“有线”，EntityID含switch
		if e.Category == core.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
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

	areaName := core.SpiltAreaName(entity.AreaName)
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

func presenceSensorOff(entity *core.Entity) (*Automation, error) {
	var (
		areaID        = entity.AreaID
		lights        []*core.Entity
		wiredSwitches []*core.Entity
	)

	// 查找同区域所有实体
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		// 灯组:EntityID前缀light.
		if e.Category == core.CategoryLight {
			lights = append(lights, e)
		}
		// 有线开关: 名字含“开关”和“有线”，EntityID含switch
		if e.Category == core.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
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

	areaName := core.SpiltAreaName(entity.AreaName)

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
