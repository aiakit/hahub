package core

import (
	"hahub/internal/chat"

	"github.com/aiakit/ava"
)

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
