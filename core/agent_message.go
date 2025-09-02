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

// 留言功能
// 1.留言，2.播放留言

// 留言内容存储
var messages = make([]string, 0, 10)
var messageMutex sync.RWMutex

func RunMessage(message, aiMessage, deviceId string) string {
	// 查询留言
	if strings.Contains(aiMessage, "query_message") {
		// 如果留言为空
		messageMutex.RLock()
		defer messageMutex.RUnlock()

		msgLen := len(messages)
		if msgLen == 0 {
			return "暂无留言。"
		}

		var sendData = make([]string, 0)
		if len(messages) > 3 {
			sendData = messages[len(messages)-3:]
		} else {
			sendData = messages
		}

		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role:    "system",
				Content: fmt.Sprintf(`请根据用户的查询请求，从以下留言内容中提取相关信息并回答用户: %s。如果留言为空，请告知我暂无留言。如果有留言，请按时间顺序列出留言内容。`, x.MustMarshalEscape2String(sendData)),
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

	// 添加留言
	if strings.Contains(aiMessage, "add_message") {
		// 从AI消息中提取留言内容
		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role:    "system",
				Content: fmt.Sprintf(`你的一个职责是留言功能，根据我给你的内容中提取留言的内容。如果我知道了留言人信息，请包含在回复内容中。当前时间：%v`, time.Now().Format(time.DateTime)),
			},
			{
				Role:    "user",
				Content: message,
			},
		})
		if err != nil {
			ava.Error(err)
			return "无法解析留言内容"
		}

		// 添加留言到内存
		messageMutex.Lock()
		messages = append(messages, result)
		messageMutex.Unlock()

		// 保存到本地文件

		return "留言已保存成功"
	}

	return ""
}
