package core

import (
	"hahub/data"
	"hahub/intelligent"
)

func InitAutomation(message, aiMessage, deviceId string) string {
	intelligent.ChaosAutomation()
	data.CallService().WaitForCallService()
	return "已根据你家里的设备和房屋信息规划好了自动化"
}
