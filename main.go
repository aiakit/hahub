package main

import (
	"hahub/data"
	"hahub/intelligent"
	"time"

	"github.com/aiakit/ava"
)

func main() {
	now := time.Now()
	// 等待 chaos.go 的初始化完成
	data.WaitForInit()

	intelligent.WalkPresenceSensorAir(ava.Background())
	//必须先创建脚本再创建自动化，这里不打开，改为ai驱动
	//intelligent.Chaos()

	//启动音箱ai驱动
	//core.CoreChaos()
	ava.Debugf("Starting Hahub ok! |latency=%.2fs", time.Since(now).Seconds())
	select {}
}
