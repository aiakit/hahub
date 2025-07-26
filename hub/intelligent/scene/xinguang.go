package scene

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
		if len(s.Entities) > 0 {
			CreateScene(c, s)
		}
	}()

	func() {
		s := InitModeTwo(c)
		if len(s.Entities) > 0 {
			CreateScene(c, s)
		}
	}()

	func() {
		s := InitModeThree(c)
		if len(s.Entities) > 0 {
			CreateScene(c, s)
		}
	}()

	//func() {
	//	s := InitModeFour(c)
	//	if len(s.Entities) > 0 {
	//		CreateScene(c, s)
	//	}
	//}()
}

// 光随影动
func InitModeOne(c *ava.Context) *Scene {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var s = &Scene{
		Name: "馨光光随影动",
	}

	var en = make(map[string]interface{})
	var meta = make(map[string]interface{})

	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}
		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			enTmp["state"] = "2"
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			enTmp["state"] = "80"
		}
		if strings.Contains(e.OriginalName, "柔和度设置") {
			enTmp["state"] = "40"
		}

		if strings.Contains(e.OriginalName, "速度") {
			enTmp["state"] = "80"
		}

		//fmt.Printf("------%s---\n", e.OriginalName)
		//fmt.Println("------1---", len(e.OriginalName))
		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}

	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			enTmp["state"] = "58"
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			enTmp["state"] = "80"
		}
		if strings.Contains(e.OriginalName, "动态亮度") {
			enTmp["state"] = "100"
		}
		if strings.Contains(e.OriginalName, "速度") {
			enTmp["state"] = "50"
		}
		if strings.Contains(e.OriginalName, "LED运行模式") {
			enTmp["state"] = "动态模式"
		}

		//注意元数据中有空格
		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}
	}

	s.Metadata = meta
	s.Entities = en

	return s
}

// 音乐律动
func InitModeTwo(c *ava.Context) *Scene {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var s = &Scene{
		Name: "馨光音乐律动",
	}

	var en = make(map[string]interface{})
	var meta = make(map[string]interface{})

	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}

		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			enTmp["state"] = "22"
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			enTmp["state"] = "80"
		}
		if strings.Contains(e.OriginalName, "柔和度设置") {
			enTmp["state"] = "40"
		}
		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
			enTmp["state"] = "on"
		}
		if strings.Contains(e.OriginalName, "速度") {
			enTmp["state"] = "50"
		}

		if strings.Contains(e.OriginalName, "放大等级") {
			enTmp["state"] = "5"
		}

		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}

	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			enTmp["state"] = "5"
		}

		if strings.Contains(e.OriginalName, "饱和度") {
			enTmp["state"] = "80"
		}
		if strings.Contains(e.OriginalName, "动态亮度") {
			enTmp["state"] = "100"
		}
		if strings.Contains(e.OriginalName, "速度") {
			enTmp["state"] = "50"
		}
		if strings.Contains(e.OriginalName, "LED运行模式") {
			enTmp["state"] = "律动模式"
		}

		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}
	}

	s.Metadata = meta
	s.Entities = en

	return s
}

// 静态模式
func InitModeThree(c *ava.Context) *Scene {
	//初始化馨光主机
	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
	if !ok {
		return nil
	}

	if len(entities) == 0 {
		return nil
	}

	var s = &Scene{
		Name: "馨光静态模式",
	}

	var en = make(map[string]interface{})
	var meta = make(map[string]interface{})

	for _, e := range entities {
		if !strings.Contains(e.DeviceName, "主机") {
			continue
		}
		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "动态模式效果") {
			enTmp["state"] = "8"
		}

		if strings.Contains(e.OriginalName, "饱和度设置") {
			enTmp["state"] = "100"
		}
		if strings.Contains(e.OriginalName, "柔和度设置") {
			enTmp["state"] = "40"
		}

		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}

	}

	//灯带设置
	for _, e := range entities {
		if strings.Contains(e.DeviceName, "主机") {
			continue
		}

		var enTmp = make(map[string]interface{})

		if strings.Contains(e.OriginalName, "饱和度") {
			enTmp["state"] = "80"
		}

		if strings.Contains(e.OriginalName, "动态亮度") {
			enTmp["state"] = "100"
		}

		if strings.Contains(e.OriginalName, "LED运行模式") {
			enTmp["state"] = "静态模式"
		}

		if e.OriginalName == " 灯" {
			enTmp["state"] = "on"
			enTmp["brightness"] = "255"
		}

		if len(enTmp) > 0 {
			enTmp["friendly_name"] = e.OriginalName
			en[e.EntityID] = enTmp
			meta[e.EntityID] = map[string]interface{}{
				"entity_only": true,
			}
		}
	}

	s.Metadata = meta
	s.Entities = en

	return s
}

//// 彩光模式
//func InitModeFour(c *ava.Context) *Scene {
//	//初始化馨光主机
//	entities, ok := core.GetEntityCategoryMap()[core.CategoryXinGuang]
//	if !ok {
//		return nil
//	}
//
//	if len(entities) == 0 {
//		return nil
//	}
//
//	var s = &Scene{
//		DeviceName: "馨光彩光模式",
//	}
//
//	var en = make(map[string]interface{})
//	var meta = make(map[string]interface{})
//
//	for _, e := range entities {
//		if !strings.Contains(e.DeviceName, "主机") {
//			continue
//		}
//		var enTmp = make(map[string]interface{})
//
//		if strings.Contains(e.OriginalName, "动态模式效果") {
//			enTmp["state"] = "13"
//		}
//
//		if strings.Contains(e.OriginalName, "饱和度设置") {
//			enTmp["state"] = "100"
//		}
//		if strings.Contains(e.OriginalName, "柔和度设置") {
//			enTmp["state"] = "100"
//		}
//		if strings.Contains(e.OriginalName, "律动和场景同步到分控") {
//			enTmp["state"] = "on"
//		}
//		if strings.Contains(e.OriginalName, "速度") {
//			enTmp["state"] = "50"
//		}
//
//		if e.OriginalName == " 灯" {
//			enTmp["state"] = "on"
//		}
//
//		if len(enTmp) > 0 {
//			enTmp["friendly_name"] = e.OriginalName
//			en[e.EntityID] = enTmp
//			meta[e.EntityID] = map[string]interface{}{
//				"entity_only": true,
//			}
//		}
//
//	}
//
//	//灯带设置
//	for _, e := range entities {
//		if strings.Contains(e.DeviceName, "主机") {
//			continue
//		}
//
//		var enTmp = make(map[string]interface{})
//
//		if strings.Contains(e.OriginalName, "动态模式效果") {
//			enTmp["state"] = "2"
//		}
//
//		if strings.Contains(e.OriginalName, "饱和度") {
//			enTmp["state"] = "100"
//		}
//		if strings.Contains(e.OriginalName, "动态亮度") {
//			enTmp["state"] = "100"
//		}
//		if strings.Contains(e.OriginalName, "速度") {
//			enTmp["state"] = "80"
//		}
//		if strings.Contains(e.OriginalName, "LED运行模式") {
//			enTmp["state"] = "动态模式"
//		}
//
//		if e.OriginalName == " 灯" {
//			enTmp["state"] = "on"
//		}
//
//		if len(enTmp) > 0 {
//			enTmp["friendly_name"] = e.OriginalName
//			en[e.EntityID] = enTmp
//			meta[e.EntityID] = map[string]interface{}{
//				"entity_only": true,
//			}
//		}
//	}
//	s.Metadata = meta
//	s.Entities = en
//
//	return s
//}
