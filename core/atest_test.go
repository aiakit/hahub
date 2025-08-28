package core

import (
	_ "hahub/intelligent"
	"hahub/internal/chat"
	"testing"
)

func init() {
	CoreChaos()
}
func TestChat(t *testing.T) {
	//result, err := chatCompletionInternal([]*chat.ChatMessage{
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
	//aaa("我有多少个场景","我还不太明白")
	aaa("小孩房温馨场景", "我不太清楚")
	select {}
}

func aaa(message1, message2 string) {
	SpeakerProcessSend(&Conversationor{
		Conversation: []*chat.ChatMessage{{Role: "user", Content: message1}, {Role: "assistant", Content: message2, Name: "jinx"}},
		entityId:     "text.xiaomi_lx06_ae32_play_text",
		deviceId:     "323aa55fea880a7e5bdb075f6a8ef925",
	})
}
