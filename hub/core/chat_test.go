package speaker

import (
	"fmt"
	"hahub/hub/internal/chat"
	"hahub/hub/internal/x"
	"testing"
)

func TestChat(t *testing.T) {
	result, err := chatCompletion([]*chat.ChatMessage{
		{Role: "user", Content: fmt.Sprintf(preparePrompts, x.MustMarshalEscape2String(logicDataMap))},
		{Role: "user", Content: "修改回家场景"},
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
