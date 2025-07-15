package main

import (
	"hahub/hub/core"
	"hahub/hub/intelligent/automation"

	"github.com/aiakit/ava"
)

func main() {
	// 等待 chaos.go 的初始化完成
	core.WaitForInit()

	automation.Chaos()
	ava.Debugf("Starting Hahub ok!")
	select {}
}
