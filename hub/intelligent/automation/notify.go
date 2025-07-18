package automation

import "hahub/hub/core"

// 火灾，燃气，漏水通知
func notify() {
	core.RegisterDataHandler(handleFire)
}

// 烟雾报警
func handleFire(event *core.StateChanged, data []byte) {

}

//漏水报警

//燃气报警
