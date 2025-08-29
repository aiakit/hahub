package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

// 当用户在客厅音箱喊话，找到需要播报的文本的音箱地址，播报消息给其他用户,相当于对讲机的功能
// 需要记录某个人的名字对应的音箱地址
var speakerMap = map[string]string{} //name:area
var speakerMapMutex = new(sync.RWMutex)

func SendMessagePlay(message, aiMessage, deviceId string) string {
	speakerMapMutex.RLock()
	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role: "user",
			Content: fmt.Sprintf(`音箱设备数据：%s,名字和位置数据：%s,根据我意图找出目标音箱，你需要将我的话进行整理表达放到content字段中，并按照格式返回给我。
返回JSON格式情况：
1.针对某个人的消息：
{"id":"设备id","content":"xx快来吃饭了","is_broadcast":false,"name":"xx"}
2.广播给多人的消息：
{"content":"xx妈妈叫你们吃饭了","is_broadcast":true}`, x.MustMarshalEscape2String(gSpeakerProcess.speakerEntityPlayTextEntity), x.MustMarshalEscape2String(speakerMap)),
		},
		{
			Role:    "user",
			Content: message,
		},
	})
	speakerMapMutex.RUnlock()

	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	var playBody struct {
		Id          string `json:"id,omitempty"`
		Content     string `json:"content,omitempty"`
		IsBroadcast bool   `json:"is_broadcast,omitempty"`
		Name        string `json:"name,omitempty"`
	}

	err = x.Unmarshal([]byte(result), &playBody)
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	if playBody.Name != "" {
		if v, ok := gSpeakerProcess.speakerEntityPlayTextEntity[playBody.Id]; ok {
			speakerMapMutex.Lock()
			speakerMap[playBody.Name] = v.AreaName
			speakerMapMutex.Unlock()
		}
	}

	if strings.Contains(aiMessage, "send_message_to_someone") {
		if playBody.Id == "" || playBody.Content == "" {
			return "没有找到音箱设备"
		}

		//暂停接收
		setIsReceivedPlayText(gSpeakerProcess.speakerEntityPlayText[playBody.Id], 1)
		//发送广播
		PlayTextActionDirect(playBody.Id, playBody.Content)
		time.Sleep(GetPlaybackDuration(playBody.Content))
		//开启接收
		setIsReceivedPlayText(gSpeakerProcess.speakerEntityPlayText[playBody.Id], 0)
	}

	//默认广播
	//if strings.Contains(aiMessage, "send_message_to_multiple") {
	if playBody.Content == "" {
		return "要说的内容丢失了"
	}

	for xiaomiIotDeviceID, xiaomiHomeEntityId := range gSpeakerProcess.speakerEntityPlayText {
		stopId := xiaomiIotDeviceID
		playId := xiaomiHomeEntityId
		go func() {
			//暂停接收
			setIsReceivedPlayText(stopId, 1)
			//发送广播
			PlayTextActionDirect(playId, playBody.Content)
			time.Sleep(GetPlaybackDuration(playBody.Content))
			//开启接收
			setIsReceivedPlayText(stopId, 0)
		}()

	}
	//}

	return ""
}
