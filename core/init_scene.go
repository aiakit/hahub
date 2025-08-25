package core

import (
	"hahub/intelligent"
)

func InitScene(message, aiMessage, deviceId string) string {
	intelligent.ScriptChaos()
	return "已根据你家里的设备和房屋信息规划好了场景"
}
