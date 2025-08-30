package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"

	"github.com/aiakit/ava"
)

// 当用户在客厅音箱喊话，找到需要播报的文本的音箱地址，播报消息给其他用户,相当于对讲机的功能
// 播放完之后，唤醒目标音响，监听是否有内容，如果有，返回播报内容，记录用户位置。
// 需要记录某个人的名字对应的音箱地址
var speakerMap = map[string]string{} //name:area
var speakerMapMutex = new(sync.RWMutex)

// 定向发送,如果没有位置，就再问一次要向什么房间发送
// 需要一个拦截器，不走前面的流程
func SendMessagePlay(message, aiMessage, deviceId string) string {
	var speaker = make([]*simpleEntity, 0)
	for dId, e := range gSpeakerProcess.speakerEntityPlayTextEntity {
		speaker = append(speaker, &simpleEntity{
			AreaName: e.AreaName,
			Name:     e.Name,
			Id:       dId,
		})
	}

	speakerMapMutex.RLock()
	speakerData := speakerMap
	speakerMapMutex.RUnlock()

	var content string

	if strings.Contains(aiMessage, "send_message_to_someone") {
		content = fmt.Sprintf(`音箱设备信息数据：%s,人员历史位置数据%s，根据我意图找出目标所在位置的设备，并严格按照格式返回给我。
1.优先从我的意图中找到目标对象位置信息，如果没有找到，再从对象历史位置中进行查找，如果都没有找到，need_position为true。
2.content是你用俏皮的语气把我的意图告诉目标对象。
3.设备id是通过人员位置信息得到对应的音箱设备id，我需要通过设备id进行消息播放。
4.name字段是你通过上下文得到的目标人员名称。
返回JSON格式：{"id":"设备id","content":"","need_position":false,"name":"目标人员名称"}`, x.MustMarshalEscape2String(speaker), x.MustMarshalEscape2String(speakerData))
	} else if strings.Contains(aiMessage, "send_message_to_all") {
		content = fmt.Sprintf(`音箱设备数据：%s，根据我意图找出目标所在位置的设备，并按照格式返回给我。
1.content是你用俏皮的语气把我的意图告诉目标对象。
返回JSON格式：{"content":""}`, x.MustMarshalEscape2String(speaker))
	} else {
		return "请告诉我你要发送消息给谁。"
	}

	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role:    "user",
			Content: content,
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

	var playBody struct {
		Id           string `json:"id,omitempty"`
		Content      string `json:"content,omitempty"`
		NeedPosition bool   `json:"need_position,omitempty"`
		Name         string `json:"name,omitempty"`
	}

	err = x.Unmarshal([]byte(x.FindJSON(result)), &playBody)
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一下"
	}

	if playBody.Id != "" {
		if v, ok := gSpeakerProcess.speakerEntityPlayTextEntity[playBody.Id]; ok {
			speakerMapMutex.Lock()
			speakerMap[playBody.Name] = v.AreaName
			speakerMapMutex.Unlock()
		}
	}

	if playBody.NeedPosition {
		//生成拦截器
		interceptorLock.Lock()
		// 使用deviceId变量创建闭包，确保获取到正确的deviceId值
		interceptorCall[deviceId] = func(messageInput []*chat.ChatMessage, deviceID string) string {
			messageInput = append(messageInput, &chat.ChatMessage{
				Role:    "user",
				Content: content,
			})

			var sendMessageInput = make([]*chat.ChatMessage, 0)
			for _, v := range messageInput {
				if v.Role == "assistant" && v.Name == "jinx" {
					continue
				}
				sendMessageInput = append(sendMessageInput, v)
			}

			result, err := chatCompletionHistory(sendMessageInput, deviceID)
			if err != nil {
				// 即使出错也需要清理拦截器
				interceptorLock.Lock()
				delete(interceptorCall, deviceID)
				interceptorLock.Unlock()
				return "出了点小问题呢"
			}

			var playBody struct {
				Id           string `json:"id,omitempty"`
				Content      string `json:"content,omitempty"`
				NeedPosition bool   `json:"need_position,omitempty"`
				Name         string `json:"name,omitempty"`
			}

			err = x.Unmarshal([]byte(x.FindJSON(result)), &playBody)
			if err != nil {
				ava.Error(err)
				return "服务器开小差了，请等一下"
			}

			if playBody.Id != "" {
				if v, ok := gSpeakerProcess.speakerEntityPlayTextEntity[playBody.Id]; ok {
					speakerMapMutex.Lock()
					speakerMap[playBody.Name] = v.AreaName
					speakerMapMutex.Unlock()
				}
			}

			if playBody.Id == "" {
				ava.Debugf("没有找到播放的音箱")
				return ""
			}

			//暂停ai其他对话功能
			aiLock(gSpeakerProcess.speakerEntityPlayText[playBody.Id])
			//发送广播
			PlayTextAction(playBody.Id, playBody.Content)
			//唤醒，监听等待
			aiUnlock(gSpeakerProcess.speakerEntityPlayText[playBody.Id])

			return getRandMessage()
		}

		interceptorLock.Unlock()
		return "你需要告诉我" + playBody.Name + "在哪个房间呢？"
	}

	if playBody.Content == "" {
		return "我有点迷糊了"
	}

	go func() {
		if strings.Contains(aiMessage, "send_message_to_someone") {
			if playBody.Id == "" {
				ava.Debugf("没有找到播放的音箱")
				return
			}

			//暂停ai其他对话功能
			aiLock(gSpeakerProcess.speakerEntityPlayText[playBody.Id])
			//发送广播
			PlayTextAction(playBody.Id, playBody.Content)
			//唤醒，监听等待
			aiUnlock(gSpeakerProcess.speakerEntityPlayText[playBody.Id])
			return
		}

		//默认广播
		for xiaomiIotDeviceID := range gSpeakerProcess.speakerEntityPlayText {
			stopId := xiaomiIotDeviceID
			playId := xiaomiIotDeviceID

			//消息不发送给自己
			if playId == deviceId {
				continue
			}

			go func() {
				//暂停ai其他对话功能
				aiLock(stopId)
				//发送广播
				PlayTextAction(playId, playBody.Content)
				//唤醒，监听等待
				//开启接收
				aiUnlock(stopId)
			}()

		}
	}()

	return getRandMessage()
}

func getRandMessage() string {
	return replyMessage[x.Intn(len(replyMessage)-1)]
}

var replyMessage = []string{
	"好的",
	"没问题",
	"ok啦",
	"欧了",
	"ok",
	"明白",
	"好的呢",
}
