package core

import "hahub/intelligent"

func InitALL(message, aiMessage, deviceId string) string {
	intelligent.Chaos()
	return "系统初始化已完成，为你创建了自动化和场景以及灯光、开关的参数设置。"
}
