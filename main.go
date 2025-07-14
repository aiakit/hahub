package main

import (
	"hahub/hub/intelligent/automation"
	"time"
)

func main() {
	time.Sleep(time.Second * 5)
	automation.Chaos()
	select {}
}
