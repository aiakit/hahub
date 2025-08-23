package intelligent

import (
	"errors"
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

func WalkPresenceSensor(c *ava.Context) {
	entity, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
	if !ok {
		return
	}

	for _, v := range entity {
		if strings.Contains(v.AreaName, "客厅") {
			continue
		}

		autoOn, autoOff, err := presenceSensorOn(v)
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

// 人体存在传感器
// 人在亮灯,人走灭灯
// 1.遍历所有和人在传感器区域相同的灯
// 2.对客厅、卧室区域，判断光照条件,时间条件,是否执行晚安场景
// 3.被主动关闭后，人来灯亮的所有自动化都实效，持续到被主动开启,晚安，起床
func presenceSensorOn(entity *data.Entity) (*Automation, *Automation, error) {
	var (
		areaID   = entity.AreaID
		areaName = data.SpiltAreaName(entity.AreaName)
	)

	// 查找同区域所有实体
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, nil, errors.New("entity area not found")
	}
	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.DeviceName
	if idx := strings.Index(prefix, "-"); idx > 0 {
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

	auto := &Automation{
		Alias:       areaName + "有人亮灯",
		Description: "当人体传感器检测到有人，自动打开" + areaName + "灯组和有线开关",
		Triggers: []*Triggers{{
			Type:     "occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
		}},
		Conditions: condition,
		Actions:    action,
		Mode:       "single",
	}

	// 增加光照条件
	lxConfig := getLxConfig(areaID)
	if lxConfig != nil {
		auto.Conditions = append(auto.Conditions, &Conditions{
			Condition: "numeric_state",
			EntityID:  lxConfig.EntityId,
			Below:     lxConfig.Lx, // 设置光照阈值
		})
	}

	//时间条件
	if strings.Contains(entity.AreaName, "卧室") {
		auto.Conditions = append(auto.Conditions, &Conditions{
			Condition: "time",
			After:     "16:00:00",
			Before:    "22:00:00",
			Weekday:   []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
		})
	}

	ss, ok := data.GetEntityCategoryMap()[data.CategoryScript]

	if ok {
		for _, e := range ss {
			if (strings.Contains(e.OriginalName, "晚安") || strings.Contains(e.OriginalName, "睡觉") || strings.Contains(e.OriginalName, "开/关") ||
				strings.Contains(e.OriginalName, "开关")) && strings.Contains(e.OriginalName, areaName) {

				auto.Conditions = append(auto.Conditions, &Conditions{
					Condition: "state",
					EntityID:  e.EntityID,
					State:     "on",
					Attribute: "last_triggered",
					For: &For{
						Hours:   9,
						Minutes: 0,
						Seconds: 0,
					},
				})
			}
		}
	}
	au, err := presenceSensorOff(areaName, entity, entitiesFilter)
	if err != nil {
		ava.Error(err)
		return nil, nil, err
	}

	return auto, au, nil
}

func presenceSensorOff(areaName string, entity *data.Entity, entities []*data.Entity) (*Automation, error) {

	var f *For
	f = &For{
		Hours:   0,
		Minutes: 20,
		Seconds: 0,
	}

	actions := turnOffLights(entities)
	if len(actions) == 0 {
		return nil, fmt.Errorf("%s区域没有设备", areaName)
	}

	var result = make([]interface{}, 0, 2)
	for _, e := range actions {
		result = append(result, e)
	}

	auto := &Automation{
		Alias:       areaName + "无人关灯",
		Description: "当人体传感器检测到无人，自动关闭" + areaName + "灯组和有线开关",
		Triggers: []*Triggers{{
			Type:     "not_occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For:      f,
		}},
		Actions: result,
		Mode:    "single",
	}

	return auto, nil
}

// 根据色温选颜色
func GetRgbColor(kelvin int) []int {
	if 2000 < kelvin && kelvin < 3500 {
		return []int{255, 195, 17}
	}
	if 3500 <= kelvin && kelvin < 4800 {
		return []int{0, 255, 55}
	}

	if 4800 <= kelvin {
		return []int{20, 255, 47}
	}

	return []int{255, 195, 17}
}
