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
		Alias:       "馨光光随影动脚本",
		Description: "馨光光随影动模式设置脚本",
	}

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "2",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "40",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		//注意元数据中有空格
		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "58",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "100",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "50",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "动态模式",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		//注意元数据中有空格
		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
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
		Alias:       "馨光音乐律动脚本",
		Description: "馨光音乐律动模式设置脚本",
	}

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "22",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "40",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "on",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "50",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "放大等级") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "5",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000)},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "5",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "100",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "速度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "50",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "律动模式",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
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
		Alias:       "馨光静态模式脚本",
		Description: "馨光静态模式设置脚本",
	}

	// 主机设置
	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "动态模式效果") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "8",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "100",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "柔和度设置") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "40",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}
	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "80",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "100",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "set_state",
				Data: &actionLightData{
					State: "静态模式",
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}

		if e.OriginalName == " 灯" {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{EntityId: e.EntityID},
			})
		}
	}

	return script
}

// // 彩光模式
// func InitModeFour(c *ava.Context) *Script {
// 	//初始化馨光主机
// 	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
// 	if !ok {
// 		return nil
// 	}
//
// 	if len(entities) == 0 {
// 		return nil
// 	}
//
// 	var script = &Script{
// 		Alias:       "馨光彩光模式脚本",
// 		Description: "馨光彩光模式设置脚本",
// 	}
//
// 	// 主机设置
// 	for _, e := range entities {
// 		if !strings.Contains(e.DeviceName, "主机") {
// 			continue
// 		}
//
// 		if strings.Contains(e.OriginalName, "动态模式效果") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "13",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "饱和度设置") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "100",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "柔和度设置") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "100",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "on",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "速度") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "50",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if e.OriginalName == " 灯" {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "light.turn_on",
// 				Data: &actionLightData{
// 					Brightness: 255,
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
// 	}
//
// 	//灯带设置
// 	for _, e := range entities {
// 		if strings.Contains(e.DeviceName, "主机") {
// 			continue
// 		}
//
// 		if strings.Contains(e.OriginalName, "动态模式效果") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "2",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "饱和度") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "100",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "动态亮度") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "100",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "速度") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "80",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if strings.Contains(e.OriginalName, "LED运行模式") {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "set_state",
// 				Data: &actionLightData{
// 					State: "动态模式",
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
//
// 		if e.OriginalName == " 灯" {
// 			script.Sequence = append(script.Sequence, ActionLight{
// 				Action: "light.turn_on",
// 				Data: &actionLightData{
// 					Brightness: 255,
// 				},
// 				Target: &targetLightData{EntityId: e.EntityID},
// 			})
// 		}
// 	}
//
// 	return script
// }
