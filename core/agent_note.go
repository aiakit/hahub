package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

// 记事本功能
// 播放所有记事
// 定时播放某个事情

// key总结归纳事件:value:事件详情
var note = make([]string, 0, 10)
var noteLock sync.RWMutex

// todo 多次测试得不到正确结果。
func RunNote(message, aiMessage, deviceId string) string {

	//查询记事本
	if strings.Contains(aiMessage, "query_note") {
		noteLock.RLock()
		defer noteLock.RUnlock()
		// 如果note为空，从本地文件加载
		if len(note) == 0 {
			return "当前暂无记事内容，我最多只能帮你记录最新的10条内容。"
		}

		var sendData = make([]string, 0)
		if len(note) > 10 {
			sendData = note[len(note)-10:]
		}

		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role:    "system",
				Content: fmt.Sprintf(`请根据用户的查询请求，从以下记事本内容中提取相关信息并回答用户: %s。如果记事本为空，请告知我暂无记事。如果记事本中有人物称谓，记得也返回给我。`, x.MustMarshal2String(sendData)),
			},
			{
				Role:    "user",
				Content: message,
			},
		})
		if err != nil {
			ava.Error(err)
			return "服务器出错了"
		}
		return result
	}

	//添加记事本
	if strings.Contains(aiMessage, "add_note") {

		// 从AI消息中提取记事内容
		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`你是一个日程安排管家，根据我的意图分析并返回结果。当前时间：%v。
1. content: 日程发生之后，站在你的角度通知我的内容，例如：“下午4点56分到了，记得开会。”。
2. dely: 日程是在多少秒之后，如果我告诉了你一个已经发生了的时间，为负数。
返回JSON格式：{"dely":0,"content":""}`, time.Now().Format(time.DateTime)),
			},
			{
				Role:    "user",
				Content: message,
			},
		})
		if err != nil {
			ava.Error(err)
			return "无法解析记事内容"
		}

		var resultObject struct {
			Timing  int    `json:"dely"`
			Content string `json:"content"`
		}

		err = x.Unmarshal([]byte(x.FindJSON(result)), &resultObject)
		if err != nil {
			ava.Error(err)
			return "解析内容出错"
		}

		if resultObject.Content == "" {
			return "提取内容失败了"
		}

		msg := registerNote(resultObject.Timing, resultObject.Content, deviceId)

		noteLock.Lock()
		// 添加特殊字符前缀并保存到数组
		note = append(note, result)
		noteLock.Unlock()

		return msg
	}

	return ""
}

func registerNote(timing int, content, deviceId string) string {
	entityId, ok := gSpeakerProcess.speakerEntityPlayText[deviceId]
	if !ok {
		ava.Debugf("没有找到播放的音箱设备:%s", deviceId)
		return ""
	}

	if timing > 0 {
		x.TimingwheelAfter(time.Second*time.Duration(timing), func() {

			err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
				EntityId: entityId,
				Message:  content,
			}, nil)
			if err != nil {
				ava.Error(err)
			}
		})
		return replyMessage[len(replyMessage)-1]
	}

	if timing <= 0 {
		aiLock(deviceId)
		err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
			EntityId: entityId,
			Message:  content,
		}, nil)
		if err != nil {
			ava.Error(err)
		}
		GetPlaybackDuration(content)
		aiUnlock(deviceId)
	}

	return ""
}
