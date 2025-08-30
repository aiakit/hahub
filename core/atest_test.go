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
	//aaa("客厅灯是什么状态", "我还不太明白")
	//aaa("离家自动化是怎样工作的，帮我详细介绍一下", "我不太清楚")
	//aaa("执行系统初始化", "我不太清楚")
	//aaa("我心情不好", "我不太清楚")
	//aaa("小孩房温馨场景执行后执行客厅温馨场景", "我不太清楚")
	//aaa("执行厨房无人关灯自动化", "我不太清楚")
	//aaa("把电竞房灯带灯组亮度调到10", "我不太清楚")
	//aaa("走廊灯亮度调到50，色温调到3000", "我不太清楚")
	//aaa("关闭客厅电视", "我不太清楚")
	//aaa("每天下午5点关闭客厅电视", "我不太清楚")
	//aaa("叫花花别玩了", "我不太清楚")
	//aaa("告诉所有人我饿了", "我不太清楚")
	aaa("介绍一下我的智能家居", "我不太清楚")

	select {}
}

func aaa(message1, message2 string) {
	SpeakerProcessSend(&Conversationor{
		Conversation: []*chat.ChatMessage{{Role: "user", Content: message1, Name: "master"}, {Role: "assistant", Content: message2, Name: "jinx"}},
		entityId:     "text.xiaomi_l16a_3165_play_text",
		deviceId:     "f6bdc4c4292eccc26412727d62148dfd",
	})
}
