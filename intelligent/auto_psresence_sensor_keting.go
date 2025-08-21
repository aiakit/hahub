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

	entityLx, ok := data.GetEntityCategoryMap()[data.CategoryLxSensor]
	if !ok {
		return
	}

	var l *data.Entity
	for _, lx := range entityLx {
		if strings.Contains(lx.AreaName, "客厅") {
			l = lx
		}
	}

	if l == nil {
		return
	}

	for _, v := range entity {
		if !strings.Contains(v.AreaName, "客厅") {
			continue
		}

		func() {
			autoOn, err := presenceSensorOnKeting(v, l, 200, 999, []string{""}, 3000)
			if err != nil {
				c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				return
			}
			CreateAutomation(c, autoOn)
		}()

		func() {
			autoOn, err := presenceSensorOnKeting(v, l, 150, 200, []string{"次灯"}, 4000)
			if err != nil {
				c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				return
			}
			CreateAutomation(c, autoOn)
		}()
		func() {
			autoOn, err := presenceSensorOnKeting(v, l, 80, 150, []string{"次灯", "主灯"}, 4700)
			if err != nil {
				c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				return
			}
			CreateAutomation(c, autoOn)
		}()
		func() {
			autoOn, err := presenceSensorOnKeting(v, l, 0, 80, []string{"所有"}, 6000)
			if err != nil {
				c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
				return
			}
			CreateAutomation(c, autoOn)
		}()

		autoOff, err := presenceSensorOffKeting(v)
		if err != nil {
			c.Error(err)
			continue
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
func presenceSensorOnKeting(entity, lumen *data.Entity, lxMin, lxMax float64, during []string, kelvin int) (*Automation, error) {
	var (
		areaID             = entity.AreaID
		atmosphereSwitches []*data.Entity
		normalSwitches     []*data.Entity
		atmosphereLights   []*data.Entity
		normalLights       []*data.Entity
		xinguangLights     []*data.Entity
	)

	//// 1. 取entity.Name中'-'前的前缀
	//prefix := entity.DeviceName
	//if idx := strings.Index(prefix, "-"); idx > 0 {
	//	prefix = prefix[:idx]
	//}

	var duringName string
	// 查找同区域所有实体
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	for _, e := range entities {
		if e.Category == data.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				var exist bool
				for k, d := range during {
					if (strings.Contains(e.DeviceName, d) || d == "所有") && !strings.Contains(e.DeviceName, "氛围") {
						exist = true
					}
					if k == len(during)-1 {
						duringName = d
					}
				}
				if exist {
					normalSwitches = append(normalSwitches, e)
				}
			}
		}
		if e.Category == data.CategoryLightGroup {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				var exist bool
				for k, d := range during {
					if (strings.Contains(e.DeviceName, d) || d == "所有") && !strings.Contains(e.DeviceName, "氛围") {
						exist = true
					}
					if k == len(during)-1 {
						duringName = d
					}
				}
				if exist {
					normalLights = append(normalLights, e)
				}
			}
		}

		if e.Category == data.CategoryLight {
			if strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜") || strings.Contains(e.DeviceName, "馨光") {
				atmosphereLights = append(atmosphereLights, e)
			}
		}

		if e.Category == data.CategoryXinGuang {
			if !strings.Contains(e.DeviceName, "主机") && strings.HasPrefix(e.EntityID, "light.") {
				xinguangLights = append(xinguangLights, e)
			}
		}
	}

	//var isNull = false

	if len(atmosphereLights) == 0 && len(normalLights) == 0 && len(normalSwitches) == 0 && len(atmosphereSwitches) == 0 {
		//isNull = true
		//寻找当前区域所有单灯
		for _, e := range entities {
			if e.Category == data.CategoryLight {
				normalLights = append(normalLights, e)
			}
		}
	}

	if duringName == "" {
		duringName = "氛围"
	}

	var actions []interface{}
	var parallel1 = make(map[string][]interface{})
	// 1. 先开氛围灯
	for _, l := range atmosphereLights {
		//if prefix != "" && !isNull {
		//	if !strings.Contains(l.DeviceName, prefix) {
		//		continue
		//	}
		//}
		var act = &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: kelvin,
				BrightnessPct:   100,
			},
			Target: &targetLightData{DeviceId: l.DeviceID},
		}

		if strings.Contains(l.DeviceName, "彩") {
			act.Data = &actionLightData{BrightnessStepPct: 100}
		}

		parallel1["parallel"] = append(parallel1["parallel"], act)
	}

	// 1. 开馨光
	for _, l := range xinguangLights {
		if data.GetXinGuang(l.DeviceID) == "" {
			continue
		}

		//改为静态模式,不能并行执行，必须优先执行
		actions = append(actions, &ActionLight{
			DeviceID: l.DeviceID,
			Domain:   "select",
			EntityID: data.GetXinGuang(l.DeviceID),
			Type:     "select_option",
			Option:   "静态模式",
		})

		//修改颜色
		parallel1["parallel"] = append(parallel1["parallel"], &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				BrightnessPct: 100,
				RgbColor:      GetRgbColor(kelvin),
			},
			Target: &targetLightData{DeviceId: l.DeviceID},
		})
	}

	// 2. 先开氛围开关
	for _, s := range atmosphereSwitches {
		//if prefix != "" && !isNull {
		//	if !strings.Contains(s.DeviceName, prefix) {
		//		continue
		//	}
		//}

		parallel1["parallel"] = append(parallel1["parallel"], &ActionCommon{
			Type:     "turn_on",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}
	if len(parallel1) > 0 {
		actions = append(actions, parallel1)
	}
	// 3. 延迟3秒
	if len(normalLights) > 0 || len(normalSwitches) > 0 {
		actions = append(actions, ActionTimerDelay{Delay: struct {
			Hours        int `json:"hours"`
			Minutes      int `json:"minutes"`
			Seconds      int `json:"seconds"`
			Milliseconds int `json:"milliseconds"`
		}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
	}
	var parallel2 = make(map[string][]interface{})
	// 4. 再开非氛围灯
	for _, l := range normalLights {
		//if prefix != "" && !isNull {
		//	if !strings.Contains(l.DeviceName, prefix) {
		//		continue
		//	}
		//}

		if strings.Contains(l.DeviceName, "夜") {
			parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					ColorTempKelvin: kelvin,
					BrightnessPct:   30,
				},
				Target: &targetLightData{DeviceId: l.DeviceID},
			})
			continue
		}
		parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: kelvin,
				BrightnessPct:   100,
			},
			Target: &targetLightData{DeviceId: l.DeviceID},
		})
	}
	// 5. 再开非氛围开关
	for _, s := range normalSwitches {
		//if prefix != "" && !isNull {
		//	if !strings.Contains(s.DeviceName, prefix) {
		//		continue
		//	}
		//}
		parallel2["parallel"] = append(parallel2["parallel"], &ActionCommon{
			Type:     "turn_on",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}
	if len(parallel2) > 0 {
		actions = append(actions, parallel2)
	}

	if len(actions) == 0 {
		return nil, errors.New("没有设备")
	}

	areaName := data.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "存在传感器光感自动亮灯" + duringName,
		Description: fmt.Sprintf("当光照条件为大于%.2f小于%.2f,且人体传感器检测到有人，自动开灯", lxMin, lxMax),
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

	auto.Conditions = append(auto.Conditions, Conditions{
		Condition: "numeric_state",
		EntityID:  lumen.EntityID,
		Above:     lxMin,
		Below:     lxMax,
	})

	ss, ok := data.GetEntityCategoryMap()[data.CategoryScript]

	if ok {
		for _, e := range ss {
			if (strings.Contains(e.OriginalName, "晚安") || strings.Contains(e.OriginalName, "睡觉") || strings.Contains(e.OriginalName, "开/关") ||
				strings.Contains(e.OriginalName, "开关")) && strings.Contains(e.OriginalName, areaName) {
				auto.Conditions = append(auto.Conditions, Conditions{
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

	return auto, nil
}

func presenceSensorOffKeting(entity *data.Entity) (*Automation, error) {
	var (
		areaID             = entity.AreaID
		atmosphereSwitches []*data.Entity
		normalSwitches     []*data.Entity
		atmosphereLights   []*data.Entity
		normalLights       []*data.Entity
	)

	// 查找同区域所有实体
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		if e.Category == data.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}
		if e.Category == data.CategoryLightGroup {
			if strings.Contains(e.DeviceName, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
		}

		if e.Category == data.CategoryLight {
			if strings.Contains(e.DeviceName, "彩") || strings.Contains(e.DeviceName, "夜") {
				atmosphereLights = append(atmosphereLights, e)
			}
		}
	}

	var actions []interface{}
	var parallel1 = make(map[string][]interface{})
	// 先关非氛围灯
	for _, l := range normalLights {
		parallel1["parallel"] = append(parallel1["parallel"], &ActionLight{
			Type:     "turn_off",
			DeviceID: l.DeviceID,
			EntityID: l.EntityID,
			Domain:   "light",
		})
	}
	// 先关非氛围开关
	for _, s := range normalSwitches {
		parallel1["parallel"] = append(parallel1["parallel"], &ActionCommon{
			Type:     "turn_off",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}
	if len(parallel1) > 0 {
		actions = append(actions, parallel1)
	}
	// 延迟3秒
	if len(atmosphereLights) > 0 || len(atmosphereSwitches) > 0 {
		actions = append(actions, ActionTimerDelay{Delay: struct {
			Hours        int `json:"hours"`
			Minutes      int `json:"minutes"`
			Seconds      int `json:"seconds"`
			Milliseconds int `json:"milliseconds"`
		}{Hours: 0, Minutes: 0, Seconds: 3, Milliseconds: 0}})
	}
	var parallel2 = make(map[string][]interface{})
	// 再关氛围灯
	for _, l := range atmosphereLights {
		parallel2["parallel"] = append(parallel2["parallel"], &ActionLight{
			Type:     "turn_off",
			DeviceID: l.DeviceID,
			EntityID: l.EntityID,
			Domain:   "light",
		})
	}
	// 再关氛围开关
	for _, s := range atmosphereSwitches {
		parallel2["parallel"] = append(parallel2["parallel"], &ActionCommon{
			Type:     "turn_off",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}
	if len(parallel2) > 0 {
		actions = append(actions, parallel2)
	}

	areaName := data.SpiltAreaName(entity.AreaName)

	auto := &Automation{
		Alias:       areaName + "无人关灯",
		Description: "当人体传感器检测到无人，自动关闭" + areaName + "灯组和有线开关",
		Triggers: []Triggers{{
			Type:     "not_occupied",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For: &For{
				Hours:   0,
				Minutes: 5,
				Seconds: 0,
			},
		}},
		Actions: actions,
		Mode:    "single",
	}

	return auto, nil
}
