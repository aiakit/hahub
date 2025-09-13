package main

import (
	"hahub/core"
	_ "hahub/intelligent"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aiakit/ava"
)

func main() {
	now := time.Now()
	// 等待 chaos.go 的初始化完成

	//必须先创建脚本再创建自动化，这里不打开，改为ai驱动
	//intelligent.Display(ava.Background())
	//intelligent.InitSwitchSelect(ava.Background())
	//intelligent.InitHoming(ava.Background())
	//intelligent.InitLevingHome(ava.Background())
	//intelligent.CreateAutomation(ava.Background())
	//intelligent.Chaos()

	//启动音箱ai驱动
	core.CoreChaos()

	ava.Debugf("Starting Hahub ok! |latency=%.2fs", time.Since(now).Seconds())

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		shutdown()
	}
}

func shutdown() {
	ava.Debug("-------------------exit------------------")
	time.Sleep(time.Second * 2)
}
