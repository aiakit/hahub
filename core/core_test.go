package core

import (
	"hahub/data"
	"hahub/internal/chat"
	"testing"
)

func init() {
	data.WaitForInit()
	CoreChaos()
}

func TestChat(t *testing.T) {
	//result, err := chatCompletion([]*chat.ChatMessage{
	//	{Role: "user", Content: fmt.Sprintf(preparePrompts, x.MustMarshalEscape2String(logicDataMap))},
	//	{Role: "user", Content: "修改回家场景"},
	//})
	//
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Log(result)
}

func TestSendMessage(t *testing.T) {
	speakerProcessSend(&conversationor{
		Conversation: []*chat.ChatMessage{{Role: "user", Content: "我有多少个场景"}, {Role: "assistant", Content: "我还不太明白", Name: "jinx"}},
		entityId:     "text.xiaomi_lx06_ae32_play_text",
		deviceId:     "323aa55fea880a7e5bdb075f6a8ef925",
	})
	select {}
}
