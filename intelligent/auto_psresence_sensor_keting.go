package intelligent

import (
	"errors"
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

func WalkPresenceSensorKeting(c *ava.Context) {
	entity, ok := data.GetEntityCategoryMap()[data.CategoryHumanPresenceSensor]
	if !ok {
		return
	}

	for _, v := range entity {
		if !strings.Contains(v.AreaName, "客厅") {
			continue
		}

		autoOn, autoOff, err := presenceSensorOnKeting(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
			continue
		}

		if autoOn != nil {
			AddAutomation2Queue(c, autoOn)
		}

		if autoOff != nil {
			AddAutomation2Queue(c, autoOff)
		}
	}
}

// 人体存在传感器
// 人在亮灯,人走灭灯
// 1.遍历所有和人在传感器区域相同的灯
// 2.对客厅、卧室区域，判断光照条件,时间条件,是否执行晚安场景
// 3.被主动关闭后，人来灯亮的所有自动化都实效，持续到被主动开启,晚安，起床
func presenceSensorOnKeting(entity *data.Entity) (*Automation, *Automation, error) {
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
	//氛围灯和其他灯
	var normal = make([]*data.Entity, 0, 4)
	var atmosphere = make([]*data.Entity, 0, 4)
	for _, e := range entitiesFilter {
		if strings.Contains(e.DeviceName, "氛围") {
			atmosphere = append(atmosphere, e)
		} else {
			normal = append(normal, e)
		}
	}

	normalLight := turnOnLights(normal, 100, 4800, false)
	atmosphereLight := turnOnLights(atmosphere, 100, 3000, false)

	if len(normalLight) == 0 && len(atmosphereLight) == 0 {
		return nil, nil, fmt.Errorf("%s区域没有发现灯", areaName)
	}

	var acts = make([]IfThenELSEAction, 0, 2)

	//如果是下午17点到下午19点，打开所有灯，亮度50,色温3000
	//如果是下午19点到下午24点，打开所有灯，亮度100,色温5200
	//如果是0点到17点，只打开氛围灯，亮度50,色温3000

	func() {
		s := setLightSettings(
			normalLight,
			atmosphereLight,
			entity,
			40, 4000, 80, 3000,
			"17:00:00", "18:59:59")
		if len(s.Then) > 0 {
			acts = append(acts, s)
		}
	}()

	func() {
		s := setLightSettings(
			normalLight,
			atmosphereLight,
			entity,
			100, 5200, 100, 5200,
			"19:00:00", "23:59:59")
		if len(s.Then) > 0 {
			acts = append(acts, s)
		}
	}()

	func() {
		s := setLightSettings(
			nil,
			atmosphereLight,
			entity,
			0, 0, 80, 3000,
			"00:00:00", "16:59:59")
		if len(s.Then) > 0 {
			acts = append(acts, s)
		}
	}()

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
		Mode: "single",
	}

	for _, v := range acts {
		auto.Actions = append(auto.Actions, v)
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

	au, err := presenceSensorOffKeting(areaName, entity, entitiesFilter)
	if err != nil {
		ava.Error(err)
		return nil, nil, err
	}

	return auto, au, nil
}

func presenceSensorOffKeting(areaName string, entity *data.Entity, entities []*data.Entity) (*Automation, error) {

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

	resultTv := turnOffTv(entity.AreaID)
	if len(resultTv) > 0 {
		result = append(result, resultTv...)
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

func setLightSettings(
	normalLight,
	atmosphereLight []*ActionLight,
	entity *data.Entity,
	normalBrightness float64, normalKelvin int,
	atmosphereBrightness float64, atmosphereKelvin int,
	after, before string) IfThenELSEAction {

	// 设置 normalLight 的属性
	if normalLight != nil {
		for i, e := range normalLight {
			if e.subCategory == data.CategoryLightTemp {
				normalLight[i].Data.ColorTempKelvin = normalKelvin
				normalLight[i].Data.BrightnessStepPct = normalBrightness
			} else if e.subCategory != data.CategoryWiredSwitch {
				normalLight[i].Data.BrightnessStepPct = 100
			}
		}
	}

	// 设置 atmosphereLight 的属性
	for i, e := range atmosphereLight {
		if e.subCategory == data.CategoryLightTemp {
			atmosphereLight[i].Data.ColorTempKelvin = atmosphereKelvin
			atmosphereLight[i].Data.BrightnessStepPct = atmosphereBrightness
		} else if e.subCategory != data.CategoryWiredSwitch {
			atmosphereLight[i].Data.BrightnessStepPct = atmosphereBrightness
		}
	}

	// 创建条件动作
	var act IfThenELSEAction
	act.If = append(act.If, ifCondition{
		Condition: "time",
		After:     after,
		Before:    before,
		Weekday:   []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
	})

	if normalLight != nil {
		_, actionNormal := spiltCondition(entity, normalLight)
		_, actionAtmosphere := spiltCondition(entity, atmosphereLight)
		act.Then = append(act.Then, actionNormal...)
		act.Then = append(act.Then, actionAtmosphere...)
	}

	if normalLight == nil {
		_, actionAtmosphere := spiltCondition(entity, atmosphereLight)
		act.Then = append(act.Then, actionAtmosphere...)
	}

	return act
}
