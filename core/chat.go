package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/internal/x"

	"github.com/aiakit/ava"
)

var systemPrompts = `你是一个智能家居助理音箱，名字叫做'小爱同学'，以下是我们最近的对话记录%s。`

func chatCompletion(msgInput []*chat.ChatMessage) (string, error) {
	historyData := GetHistory()
	var message = make([]*chat.ChatMessage, 0, 5)

	var history = make([]*chat.ChatMessage, 0, 5)
	if historyData != nil {
		for _, v := range historyData {
			for _, v1 := range v.Conversation {
				history = append(history, &chat.ChatMessage{
					Role:    v1.Role,
					Content: v1.Content,
				})
			}
		}
	}

	if len(history) == 0 {
		history = []*chat.ChatMessage{}
	}

	message = append(message, &chat.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf(systemPrompts, x.MustMarshal2String(history)),
	})

	message = append(message, msgInput...)
	ava.Debugf("req=%s", x.MustMarshal2String(message))

	return chat.ChatCompletionMessage(message)
}
