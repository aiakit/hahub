package scene

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

var lightGradientTime = &Scene{
	Name:     "初始化灯光模式设置",
	Entities: make(map[string]interface{}),
	Metadata: make(map[string]interface{}),
}

func init() {
	core.RegisterEntityCallback(registerLightGradientTime)
}

// 初始化灯光
// 创建灯光初始化
func InitLight(c *ava.Context) {
	lowest, ok := core.GetEntityCategoryMap()[core.CategoryLightLowest]
	if ok {
		s := lowestBrightness(lowest)
		CreateScene(c, s)
	}

	if len(lightGradientTime.Entities) > 0 {
		CreateScene(c, lightGradientTime)
	}
}

// 最低亮度设置
func lowestBrightness(entities []*core.Entity) *Scene {
	var s = &Scene{
		Name: "初始化最低亮度设置",
	}

	var en = make(map[string]interface{})
	var meta = make(map[string]interface{})

	for _, v := range entities {
		en[v.EntityID] = map[string]interface{}{
			"step":          1,
			"mode":          "auto",
			"friendly_name": v.OriginalName,
			"state":         "1"} //亮度0.1%
		meta[v.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	s.Metadata = meta
	s.Entities = en

	return s
}

// 灯光时间设置
func registerLightGradientTime(entity *core.Entity) {
	var step = 1
	//柜子灯带
	if (strings.Contains(entity.OriginalName, "开灯渐变时长(单位ms)") || strings.Contains(entity.OriginalName, "灯光调光时长(单位ms)") || strings.Contains(entity.OriginalName, "关灯渐变时长(单位ms)")) && strings.Contains(entity.ID, "number.") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"step":          step,
			"mode":          "auto",
			"friendly_name": entity.OriginalName,
			"state":         "3000"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	if (strings.Contains(entity.OriginalName, "通电默认状态") || strings.Contains(entity.OriginalName, "分段开关") || strings.Contains(entity.OriginalName, "恢复断电前灯光")) && strings.Contains(entity.ID, "switch.") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"device_class":  "switch",
			"friendly_name": entity.OriginalName,
			"state":         "off"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	if strings.Contains(entity.OriginalName, "指数调光") && strings.HasPrefix(entity.EntityID, "switch.") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"device_class":  "switch",
			"friendly_name": entity.OriginalName,
			"state":         "on"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	//4540997
	if strings.Contains(entity.OriginalName, "默认状态 渐变时间设置，字节[0]开灯渐变时间，字节[1]关灯渐变时间，字节[2]模式渐变时间") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"step":          step,
			"mode":          "auto",
			"friendly_name": entity.OriginalName,
			"state":         "4540997"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
		if strings.Contains(entity.OriginalName, "字节3（配置渐变、默认灯光、配置灯光、灯光变化、配置变化）") {
			lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
				"step":          step,
				"mode":          "auto",
				"friendly_name": entity.OriginalName,
				"state":         "21318213"}
		}
	}

	if strings.Contains(entity.OriginalName, "默认上电状态") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"friendly_name": entity.OriginalName,
			"state":         "上电关闭"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	if strings.Contains(entity.OriginalName, "默认状态 灯光变化") && strings.HasPrefix(entity.ID, "select.") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"friendly_name": entity.OriginalName,
			"state":         "Gradient"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	if strings.Contains(entity.OriginalName, "默认状态 默认灯光") && strings.HasPrefix(entity.ID, "select.") {
		lightGradientTime.Entities[entity.EntityID] = map[string]interface{}{
			"friendly_name": entity.OriginalName,
			"state":         "OFF"}
		lightGradientTime.Metadata[entity.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}
}
