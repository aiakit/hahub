package core

import "hahub/intelligent"

func InitAutomation(message, aiMessage, deviceId string) string {
	intelligent.ChaosAutomation()
	return "已根据你家里的设备和房屋信息规划好了自动化"
}
