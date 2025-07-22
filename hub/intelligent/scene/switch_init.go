package scene

import (
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

func InitSwitch(c *ava.Context) {
	e, ok := core.GetEntityCategoryMap()[core.CategoryWiredSwitch]
	if !ok {
		return
	}
	s := switchMode(e)
	if len(s.Entities) > 0 {
		CreateScene(c, s)
	}
}

func switchMode(entities []*core.Entity) *Scene {
	var s = &Scene{
		Name: "初始化开关模式-慎用",
	}

	var en = make(map[string]interface{})
	var meta = make(map[string]interface{})

	for _, v := range entities {
		option := "有线和无线开关"
		if !strings.Contains(v.Name, "#") {
			option = "无线开关"
		}
		if strings.Contains(v.Name, "#") {
			option = "有线和无线开关"
		}
		en[v.EntityID] = map[string]interface{}{
			"friendly_name": v.OriginalName,
			"state":         option}
		meta[v.EntityID] = map[string]interface{}{
			"entity_only": true,
		}
	}

	s.Metadata = meta
	s.Entities = en

	return s
}
