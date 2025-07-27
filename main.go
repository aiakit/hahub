package main

import (
	"hahub/hub/core"
	"hahub/hub/intelligent/automation"
	"time"

	"github.com/aiakit/ava"
)

func main() {
	now := time.Now()
	// 等待 chaos.go 的初始化完成
	core.WaitForInit()

	//必须先创建脚本再创建自动化
	automation.ScriptChaos()

	automation.Chaos()
	ava.Debugf("Starting Hahub ok! |latency=%.2fs", time.Since(now).Seconds())
	select {}
}
