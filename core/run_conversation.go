package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

// 当用户在客厅音箱喊话，找到需要播报的文本的音箱地址，播报消息给其他用户,相当于对讲机的功能
// 需要记录某个人的名字对应的音箱地址
func Conversation(message, aiMessage, deviceId string) string {
	result, err := chatCompletion([]*chat.ChatMessage{
		{
			Role: "user",
			Content: fmt.Sprintf(`音箱设备数据：%s,根据我意图找出目标音箱，你需要将我的话进行整理表达放到content字段中，并按照格式返回给我。
返回JSON格式：{"id":"设备id","content":"妈妈叫你吃饭了","is_broadcast":false}`, x.MustMarshalEscape2String(gSpeakerProcess.speakerEntityPlayTextEntity)),
		},
		{
			Role:    "user",
			Content: message,
		},
	})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	var play struct {
		Id          string `json:"id,omitempty"`
		Content     string `json:"content,omitempty"`
		IsBroadcast bool   `json:"is_broadcast,omitempty"`
	}

	err = x.Unmarshal([]byte(result), &play)
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	if strings.Contains(aiMessage, "send_message_to_someone") {
		if play.Id == "" || play.Content == "" {
			return "没有找到音箱设备"
		}

		//暂停接收
		setIsReceivedPlayText(gSpeakerProcess.speakerEntityPlayText[play.Id], 1)
		//发送广播
		PlayTextActionDirect(play.Id, play.Content)
		time.Sleep(GetPlaybackDuration(play.Content))
		//开启接收
		setIsReceivedPlayText(gSpeakerProcess.speakerEntityPlayText[play.Id], 0)
	}

	if strings.Contains(aiMessage, "send_message_to_multiple") {
		if play.Content == "" {
			return "要说的内容丢失了"
		}

		for xiaomiIotDeviceID, xiaomiHomeEntityId := range gSpeakerProcess.speakerEntityPlayText {
			stopId := xiaomiIotDeviceID
			playId := xiaomiHomeEntityId
			go func() {
				//暂停接收
				setIsReceivedPlayText(stopId, 1)
				//发送广播
				PlayTextActionDirect(playId, play.Content)
				time.Sleep(GetPlaybackDuration(play.Content))
				//开启接收
				setIsReceivedPlayText(stopId, 0)
			}()

		}
	}

	return ""
}
