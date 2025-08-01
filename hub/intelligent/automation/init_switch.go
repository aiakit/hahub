package automation

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
	if len(s.Sequence) > 0 {
		CreateScript(c, s)
	}
}

func switchMode(entities []*core.Entity) *Script {
	var s = &Script{
		Alias:       "初始化开关模式-慎用",
		Description: "初始化开关模式，包括有线和无线开关配置",
		Sequence:    []interface{}{},
	}

	for _, v := range entities {
		selectSome := "select_first"
		if !strings.Contains(v.DeviceName, "#") {
			selectSome = "select_last"
		}

		// 将原本Scene中的实体配置转换为Script中的动作序列
		s.Sequence = append(s.Sequence, ActionCommon{
			Type:     selectSome,
			DeviceID: v.DeviceID,
			EntityID: v.EntityID,
			Domain:   "select",
		})
	}

	return s
}
