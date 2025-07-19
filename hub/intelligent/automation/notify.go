package automation

import (
	"hahub/hub/core"
	"strings"
)

// 火灾，燃气，漏水通知
func notify() {
	core.RegisterDataHandler(handleFire)
}

// 烟雾报警
func handleFire(event *core.StateChanged, data []byte) {
	if strings.Contains(event.Event.NewState.Attributes.FriendlyName, "检测到高浓度烟雾") {

	}
}

// 漏水报警
func handleWater(event *core.StateChanged, data []byte) {
	if strings.Contains(event.Event.NewState.Attributes.FriendlyName, "水浸传感器 浸没状态 变湿") {

	}
}

// 燃气报警
func handleGas(event *core.StateChanged, data []byte) {
	if strings.Contains(event.Event.NewState.Attributes.FriendlyName, "天然气浓度") {

	}
}
