package core

import (
	"hahub/internal/chat"

	"github.com/aiakit/ava"
)

// 当用户在客厅音箱喊话，找到需要播报的文本的音箱地址，播报消息给其他用户,相当于对讲机的功能
func Conversation(message, aiMessage, deviceId string) string {
	result, err := chatCompletionHistory([]*chat.ChatMessage{{
		Role:    "user",
		Content: message,
	}}, deviceId)
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	return result
}
