package main

import (
	"hahub/data"
	"hahub/intelligent"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aiakit/ava"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			ava.Fatalf("revocer |err=%v", err)
		}
	}()

	now := time.Now()
	// 等待 chaos.go 的初始化完成
	data.WaitForInit()

	//必须先创建脚本再创建自动化，这里不打开，改为ai驱动
	//intelligent.GoodMorningScript(ava.Background())
	intelligent.Chaos()

	//启动音箱ai驱动
	//core.CoreChaos()

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
