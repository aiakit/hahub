package main

import (
	"hahub/hub/core"
	"hahub/hub/intelligent/automation"
)

func main() {
	// 等待 chaos.go 的初始化完成
	core.WaitForInit()

	automation.Chaos()
	select {}
}
