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

// 记事本功能
// 播放所有记事
// 定时播放某个事情

// key总结归纳事件:value:事件详情
var note = make([]string, 0, 10)
var noteLock sync.RWMutex

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
				Content: fmt.Sprintf(`请根据用户的查询请求，从以下记事本内容中提取相关信息并回答用户: %s。如果记事本为空，请告知我暂无记事。如果记事本中有人物称谓，记得也返回给我。`, x.MustMarshalEscape2String(sendData)),
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
				Role:    "system",
				Content: fmt.Sprintf(`你的一个职责是记事本，根据我给你的内容中提取记事的内容。如果我告诉了你我是谁你要返回到内容中，这样我才能知道是谁记录的此次事情。当前时间：%v`, time.Now()),
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

		noteLock.Lock()
		// 添加特殊字符前缀并保存到数组
		note = append(note, result)
		noteLock.Unlock()

		return replyMessage[len(replyMessage)-1]
	}

	return ""
}
