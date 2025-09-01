package core

import (
	"fmt"
	"hahub/intelligent"
	"hahub/internal/chat"

	"github.com/aiakit/ava"
)

// SOS紧急求助功能
func RunSOS(message, aiMessage, deviceId string) string {
	// 从AI消息中提取SOS内容和相关信息
	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role:    "system",
			Content: `你是一个紧急求助助手，需要从用户的话语中识别紧急情况并组织合适的求助内容。请简洁明了地总结用户的紧急需求。`,
		},
		{
			Role:    "user",
			Content: message,
		},
	})
	if err != nil {
		ava.Error(err)
		return "紧急求助功能暂时不可用"
	}

	// 调用智能家居场景
	intelligent.RunSript("script.sos")

	return fmt.Sprintf("紧急求助已发送：%s。", result)
}
