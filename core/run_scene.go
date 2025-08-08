package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

func RunScene(message, aiMessage, deviceId string) string {
	f := func(message, aiMessage, deviceId string) string {
		var gShortAutomations = make(map[string]*shortScene)
		entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		if !ok {
			return ""
		}

		for _, e := range entities {
			gShortAutomations[e.UniqueID] = &shortScene{
				Id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		//发送所有场景简短数据给ai
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role:    "user",
			Content: fmt.Sprintf(`这是我的全部场景信息%s，总数是%d个，根据对话内容将信息返回给我："id":""`, x.MustMarshalEscape2String(gShortAutomations), len(gShortAutomations)),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string

		for _, v := range gShortAutomations {
			if strings.Contains(result, v.Id) {
				id = v.Id
			}
		}

		if id == "" {
			return "没有找到你要的场景信息"
		}

		err = intelligent.RunSript(id)
		if err != nil {
			ava.Error(err)
			return "运行场景失败了,错误信息是：" + err.Error()
		}

		return "场景已经帮你运行了"
	}

	return CoreDelay(message, aiMessage, deviceId, f)
}
