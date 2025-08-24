package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

//馨光灯带处理
//音乐律动
//静态模式
//彩光模式

const (
	// BoxModeMovie 馨光主机
	//光随影动
	BoxModeMovie = 0
	BoxModeVideo = 1
	BoxModeGame  = 2

	//音乐律动
	BoxModeA = 18 //古典
	BoxModeB = 19 //流行
	BoxModeC = 20 //摇滚
	BoxModeD = 21 //纯享
	BoxModeE = 22 //电子
	BoxModeF = 23 //氛围
)

func InitXinGuang(c *ava.Context) {
	//初始化馨光主机
	func() {
		s := InitModeOne(c)
		if s != nil && len(s.Sequence) > 0 {
			AddScript2Queue(c, s)
		}
	}()

	func() {
		s := InitModeTwo(c)
		if s != nil && len(s.Sequence) > 0 {
			AddScript2Queue(c, s)
		}
	}()

	func() {
		s := InitModeThree(c, 100, 4000)
		if s != nil && len(s.Sequence) > 0 {
			AddScript2Queue(c, s)
		}
	}()
}

// 设置单个馨光设备开机
func turnOnXinGuangDevice(e *data.Entity, BrightnessPct float64) *ActionLight {
	//注意元数据中有空格
	if strings.HasPrefix(e.EntityID, "light.") {
		return &ActionLight{
			Type:          "turn_on",
			DeviceID:      e.DeviceID,
			EntityID:      e.EntityID,
			Domain:        "light",
			BrightnessPct: BrightnessPct,
		}
	}
	return nil
}

// 设置单个馨光设备为静态模式
func setXinGuangDeviceToStaticMode(e *data.Entity) *ActionLight {
	var isHost bool
	if strings.Contains(e.DeviceName, "主机") {
		isHost = true
	}

	if isHost {
		// 主机设置
		if strings.Contains(e.OriginalName, "动态模式效果") {
			return &ActionLight{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 8}
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			return &ActionLight{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100}
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			return &ActionLight{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 40}
		}
	} else {
		// 灯带设置
		if strings.Contains(e.OriginalName, "饱和度") {
			return &ActionLight{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100}
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			return &ActionLight{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100}
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			return &ActionLight{Domain: "select", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "select_option", Option: "静态模式"}
		}
	}
	return nil
}

// 光随影动
func InitModeOne(c *ava.Context) *Script {
	var script = &Script{
		Alias:       "馨光光随影动场景",
		Description: "馨光光随影动模式，光影模式，动感模式场景",
	}

	//初始化馨光主机
	entities, ok := data.GetEntityCategoryMap()[data.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var areaId = entities[0].AreaID

	////关闭相同区域其他所有灯
	//areaId := entities[0].AreaID
	//allLight, ok := data.GetEntityAreaMap()[areaId]
	//if ok {
	//	for _, e := range allLight {
	//		if strings.HasPrefix(e.EntityID, "light.") && e.Category == data.CategoryXinGuang {
	//			script.Sequence = append(script.Sequence, ActionLight{
	//				Type:     "turn_off",
	//				DeviceID: e.DeviceID,
	//				EntityID: e.EntityID,
	//				Domain:   "light",
	//			})
	//		}
	//	}
	//}
	//关闭所有灯
	func() {
		entitiesSpeakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entitiesSpeakers {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.HasPrefix(e.EntityID, "notify.") && e.AreaID == areaId {
					act := ExecuteTextCommand(e.DeviceID, "关闭所有灯", true)
					script.Sequence = append(script.Sequence, ActionTimerDelay{
						Delay: &delay{
							Hours:        0,
							Minutes:      0,
							Seconds:      3,
							Milliseconds: 0,
						},
					})
					script.Sequence = append(script.Sequence, act)
					break
				}
			}
		}
	}()

	//先开机
	for _, e := range entities {
		if action := turnOnXinGuangDevice(e, 1); action != nil {
			script.Sequence = append(script.Sequence, *action)
		}
	}

	script.Sequence = append(script.Sequence, ActionTimerDelay{
		Delay: &delay{
			Hours:        0,
			Minutes:      0,
			Seconds:      3,
			Milliseconds: 0,
		},
	})

	var areaName string

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 2})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 80})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 40})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 80})
		}

		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "switch", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "turn_on"})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		script.Sequence = append(script.Sequence, ActionLight{
			Type:          "turn_on",
			DeviceID:      e.DeviceID,
			EntityID:      e.EntityID,
			Domain:        "light",
			BrightnessPct: 100,
		})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 58})
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 80})
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 50})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "select", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "select_option", Option: "动态模式"})
		}
	}
	script.Alias = areaName + script.Alias

	return script
}

// 音乐律动
func InitModeTwo(c *ava.Context) *Script {
	//初始化馨光主机
	entities, ok := data.GetEntityCategoryMap()[data.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var script = &Script{
		Alias:       "馨光音乐律动场景",
		Description: "馨光律动模式设置场景",
	}

	var areaId = entities[0].AreaID

	////关闭相同区域其他所有灯
	//areaId := entities[0].AreaID
	//allLight, ok := data.GetEntityAreaMap()[areaId]
	//if ok {
	//	for _, e := range allLight {
	//		if strings.HasPrefix(e.EntityID, "light.") && e.Category != data.CategoryXinGuang {
	//			script.Sequence = append(script.Sequence, ActionLight{
	//				Type:     "turn_off",
	//				DeviceID: e.DeviceID,
	//				EntityID: e.EntityID,
	//				Domain:   "light",
	//			})
	//		}
	//	}
	//}

	//关闭所有灯
	func() {
		entitiesSpeakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if ok {
			for _, e := range entitiesSpeakers {
				if strings.Contains(e.OriginalName, "执行文本指令") && strings.HasPrefix(e.EntityID, "notify.") && e.AreaID == areaId {
					act := ExecuteTextCommand(e.DeviceID, "关闭所有灯", true)
					script.Sequence = append(script.Sequence, ActionTimerDelay{
						Delay: &delay{
							Hours:        0,
							Minutes:      0,
							Seconds:      3,
							Milliseconds: 0,
						},
					})
					script.Sequence = append(script.Sequence, act)
					break
				}
			}
		}
	}()

	//先开机
	for _, e := range entities {
		if action := turnOnXinGuangDevice(e, 100); action != nil {
			script.Sequence = append(script.Sequence, *action)
		}
	}

	script.Sequence = append(script.Sequence, ActionTimerDelay{
		Delay: &delay{
			Hours:        0,
			Minutes:      0,
			Seconds:      3,
			Milliseconds: 0,
		},
	})

	var areaName string

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 22})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 80})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 40})
		}

		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "switch", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "turn_on"})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 50})
		}

		if strings.Contains(e.OriginalName, "放大等级") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 8})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 5})
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 80})

		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 50})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "select", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "select_option", Option: "律动模式"})
		}
	}
	script.Alias = areaName + script.Alias

	return script
}

// 静态模式
func InitModeThree(c *ava.Context, BrightnessPct float64, kelvin int) *Script {
	//初始化馨光主机
	entities, ok := data.GetEntityCategoryMap()[data.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var script = &Script{
		Alias:       "馨光静态模式场景",
		Description: "馨光静态模式设置场景",
	}

	//先开机
	for _, e := range entities {
		if action := turnOnXinGuangDevice(e, BrightnessPct); action != nil {
			script.Sequence = append(script.Sequence, *action)
		}
	}

	script.Sequence = append(script.Sequence, ActionTimerDelay{
		Delay: &delay{
			Hours:        0,
			Minutes:      0,
			Seconds:      3,
			Milliseconds: 0,
		},
	})

	var areaName string
	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		result := setXinGuangDeviceToStaticMode(e)
		if result != nil {
			script.Sequence = append(script.Sequence, result)
		}
	}

	//灯带设置
	for _, e := range entities {

		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if areaName == "" {
			areaName = data.SpiltAreaName(e.AreaName)
		}

		result := setXinGuangDeviceToStaticMode(e)
		if result != nil {
			script.Sequence = append(script.Sequence, result)
		}
	}
	script.Alias = areaName + script.Alias

	return script
}
