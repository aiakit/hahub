package main

import (
	"hahub/hub/data"
	"time"

	"github.com/aiakit/ava"
)

func main() {
	now := time.Now()
	// 等待 chaos.go 的初始化完成
	data.WaitForInit()

	//必须先创建脚本再创建自动化
	//intelligent.Chaos()
	ava.Debugf("Starting Hahub ok! |latency=%.2fs", time.Since(now).Seconds())
	select {}
}
