package automation

import (
	"hahub/hub/core"
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
			CreateScript(c, s)
		}
	}()

	func() {
		s := InitModeTwo(c)
		if s != nil && len(s.Sequence) > 0 {
			CreateScript(c, s)
		}
	}()

	func() {
		s := InitModeThree(c)
		if s != nil && len(s.Sequence) > 0 {
			CreateScript(c, s)
		}
	}()

	//func() {
	//	s := InitModeFour(c)
	//	if s != nil && len(s.Sequence) > 0 {
	//		CreateScript(c, s)
	//	}
	//}()
}

// 光随影动
func InitModeOne(c *ava.Context) *Script {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var script = &Script{
		Alias:       "馨光光随影动场景",
		Description: "馨光光随影动模式设置场景",
	}

	//先开机
	for _, e := range entities {
		//注意元数据中有空格
		if strings.HasPrefix(e.EntityID, "light.") {
			script.Sequence = append(script.Sequence, ActionLight{
				Type:          "turn_on",
				DeviceID:      e.DeviceID,
				EntityID:      e.EntityID,
				Domain:        "light",
				BrightnessPct: 100,
			})
		}
	}

	//script.Sequence = append(script.Sequence, ActionTimerDelay{
	//	Delay: struct {
	//		Hours        int `json:"hours"`
	//		Minutes      int `json:"minutes"`
	//		Seconds      int `json:"seconds"`
	//		Milliseconds int `json:"milliseconds"`
	//	}{Seconds: 3},
	//})

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
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

	return script
}

// 音乐律动
func InitModeTwo(c *ava.Context) *Script {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var script = &Script{
		Alias:       "馨光音乐律动场景",
		Description: "馨光音乐律动模式设置场景",
	}

	//先开机
	for _, e := range entities {
		//注意元数据中有空格
		if strings.HasPrefix(e.EntityID, "light.") {
			script.Sequence = append(script.Sequence, ActionLight{
				Type:          "turn_on",
				DeviceID:      e.DeviceID,
				EntityID:      e.EntityID,
				Domain:        "light",
				BrightnessPct: 100,
			})
		}
	}

	//script.Sequence = append(script.Sequence, ActionTimerDelay{
	//	Delay: struct {
	//		Hours        int `json:"hours"`
	//		Minutes      int `json:"minutes"`
	//		Seconds      int `json:"seconds"`
	//		Milliseconds int `json:"milliseconds"`
	//	}{Seconds: 3},
	//})

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
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

	return script
}

// 静态模式
func InitModeThree(c *ava.Context) *Script {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
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
		//注意元数据中有空格
		if strings.HasPrefix(e.EntityID, "light.") {
			script.Sequence = append(script.Sequence, ActionLight{
				Type:          "turn_on",
				DeviceID:      e.DeviceID,
				EntityID:      e.EntityID,
				Domain:        "light",
				BrightnessPct: 100,
			})
		}
	}

	//script.Sequence = append(script.Sequence, ActionTimerDelay{
	//	Delay: struct {
	//		Hours        int `json:"hours"`
	//		Minutes      int `json:"minutes"`
	//		Seconds      int `json:"seconds"`
	//		Milliseconds int `json:"milliseconds"`
	//	}{Seconds: 3},
	//})

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 8})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 40})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100})
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "number", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "set_value", Value: 100})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionCommon{Domain: "select", DeviceID: e.DeviceID, EntityID: e.EntityID, Type: "select_option", Option: "静态模式"})
		}
	}

	return script
}
