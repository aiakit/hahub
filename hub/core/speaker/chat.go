package speaker

import (
	"fmt"
	"hahub/hub/intelligent"
	"hahub/hub/internal/chat"
	"hahub/hub/internal/x"
	"hahub/proto/phome"

	"github.com/aiakit/ava"
)

var systemPrompts = `你是一个智能家居助理，你的名字叫做'小爱同学'，以下是我们最近的对话记录%s。`

func chatCompletion(msgInput []*phome.ChatMessage) (string, error) {
	historyData := intelligent.GetHistory()
	var message = make([]*phome.ChatMessage, 0, 5)

	var history = make([]*phome.ChatMessage, 0, 5)
	if historyData != nil {
		for _, v := range historyData {
			for _, v1 := range v.Conversation {
				history = append(history, &phome.ChatMessage{
					Role:    v1.Role,
					Content: v1.Content,
				})
			}
		}
	}

	if len(history) == 0 {
		history = []*phome.ChatMessage{}
	}

	message = append(message, &phome.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf(systemPrompts, x.MustMarshal2String(history)),
	})

	message = append(message, msgInput...)
	ava.Debugf("req=%s", x.MustMarshal2String(message))

	return chat.ChatCompletionMessage(message)
}
