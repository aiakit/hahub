package chat

import (
	"fmt"
	"testing"
)

func TestChat(t *testing.T) {
	r, err := NewOpenAIProvider(DefaultQianwenModel).ChatCompletion([]*ChatMessage{
		{
			Role:    "user",
			Content: "你好吗",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(r)
}
