package core

import (
	"hahub/internal/chat"

	"github.com/aiakit/ava"
)

func Conversation(message, aiMessage, deviceId string) string {
	result, err := chatCompletionInternal([]*chat.ChatMessage{{
		Role:    "user",
		Content: message,
	}})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了"
	}

	return result
}
