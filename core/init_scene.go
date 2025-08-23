package core

import (
	"hahub/data"
	"hahub/intelligent"
)

func InitScene(message, aiMessage, deviceId string) string {
	intelligent.ScriptChaos()
	data.CallService().WaitForCallService()
	return "已根据你家里的设备和房屋信息规划好了场景"
}
