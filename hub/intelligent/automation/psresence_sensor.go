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
		if strings.Contains(v.AreaName, "客厅") {
			continue
		}

		autoOn, err := presenceSensorOn(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", core.MustMarshal2String(v), err)
			continue
		}

		CreateAutomation(c, autoOn)

		autoOff, err := presenceSensorOff(v)
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
func presenceSensorOn(entity *core.Entity) (*Automation, error) {
	var (
		areaID             = entity.AreaID
		atmosphereSwitches []*core.Entity
		normalSwitches     []*core.Entity
		atmosphereLights   []*core.Entity
		normalLights       []*core.Entity
	)

	// 查找同区域所有实体
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	for _, e := range entities {
		if e.Category == core.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			if strings.Contains(e.Name, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}
		if e.Category == core.CategoryLight {
			if strings.Contains(e.Name, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
		}
	}

	var actions []interface{}
	// 1. 先开氛围灯
	for _, l := range atmosphereLights {
		act := &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: 3000,
				BrightnessPct:   100,
			},
			Target: &targetLightData{DeviceId: l.DeviceID},
		}

		if strings.Contains(l.Name, "彩") {
			act.Data = &actionLightData{}
		}

		actions = append(actions, act)
	}
	// 2. 先开氛围开关
	for _, s := range atmosphereSwitches {
		actions = append(actions, &ActionSwitch{
			Type:     "turn_on",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
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
	// 4. 再开非氛围灯
	for _, l := range normalLights {
		if strings.Contains(l.Name, "馨光") && strings.Contains(l.Name, "主机") {
			continue
		}

		//馨光灯带只打开不调整
		if strings.Contains(l.Name, "馨光") && !strings.Contains(l.Name, "主机") {
			//改为静态模式
			actions = append(actions, &ActionLight{
				DeviceID: l.DeviceID,
				Domain:   "select",
				EntityID: core.GetXinGuang(l.DeviceID),
				Type:     "select_option",
				Option:   "静态模式",
			})

			//修改颜色
			actions = append(actions, &ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 80,
					RgbColor:      GetRgbColor(3000),
				},
				Target: &targetLightData{DeviceId: l.DeviceID},
			})
			continue
		}

		// 护眼灯特殊逻辑
		if strings.Contains(l.Name, "护眼") || strings.Contains(l.Name, "夜灯") {
			actions = append(actions, &ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					ColorTempKelvin: 3000,
					BrightnessPct:   30,
				},
				Target: &targetLightData{DeviceId: l.DeviceID},
			})
			continue
		}
		actions = append(actions, &ActionLight{
			Action: "light.turn_on",
			Data: &actionLightData{
				ColorTempKelvin: 3000,
				BrightnessPct:   80,
			},
			Target: &targetLightData{DeviceId: l.DeviceID},
		})
	}
	// 5. 再开非氛围开关
	for _, s := range normalSwitches {
		actions = append(actions, &ActionSwitch{
			Type:     "turn_on",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}

	if len(actions) == 0 {
		return nil, errors.New("没有设备")
	}

	areaName := core.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "有人亮灯",
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

	// 增加光照条件
	lxConfig := getLxConfig(areaID)
	if lxConfig != nil {
		auto.Conditions = append(auto.Conditions, Conditions{
			Condition: "numeric_state",
			EntityID:  lxConfig.EntityId,
			Below:     lxConfig.Lx, // 设置光照阈值
		})
	}

	return auto, nil
}

func presenceSensorOff(entity *core.Entity) (*Automation, error) {
	var (
		areaID             = entity.AreaID
		atmosphereSwitches []*core.Entity
		normalSwitches     []*core.Entity
		atmosphereLights   []*core.Entity
		normalLights       []*core.Entity
	)

	// 查找同区域所有实体
	entities, ok := core.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}
	for _, e := range entities {
		if e.Category == core.CategoryWiredSwitch && !strings.HasPrefix(e.EntityID, "button.") {
			if strings.Contains(e.Name, "氛围") {
				atmosphereSwitches = append(atmosphereSwitches, e)
			} else {
				normalSwitches = append(normalSwitches, e)
			}
		}
		if e.Category == core.CategoryLight {
			if strings.Contains(e.Name, "氛围") {
				atmosphereLights = append(atmosphereLights, e)
			} else {
				normalLights = append(normalLights, e)
			}
		}
	}

	var actions []interface{}
	// 先关非氛围灯
	for _, l := range normalLights {
		actions = append(actions, &ActionLight{
			Type:     "turn_off",
			DeviceID: l.DeviceID,
			EntityID: l.EntityID,
			Domain:   "light",
		})
	}
	// 先关非氛围开关
	for _, s := range normalSwitches {
		actions = append(actions, &ActionSwitch{
			Type:     "turn_off",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
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
	// 再关氛围灯
	for _, l := range atmosphereLights {
		actions = append(actions, &ActionLight{
			Type:     "turn_off",
			DeviceID: l.DeviceID,
			EntityID: l.EntityID,
			Domain:   "light",
		})
	}
	// 再关氛围开关
	for _, s := range atmosphereSwitches {
		actions = append(actions, &ActionSwitch{
			Type:     "turn_off",
			DeviceID: s.DeviceID,
			EntityID: s.EntityID,
			Domain:   "switch",
		})
	}

	areaName := core.SpiltAreaName(entity.AreaName)

	auto := &Automation{
		Alias:       areaName + "无人关灯",
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

	if strings.Contains(entity.AreaName, "卧室") {
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
