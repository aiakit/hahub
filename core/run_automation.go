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

func RunAutomation(message, aiMessage, deviceId string) string {
	device, ok := data.GetDevice()[deviceId]
	if !ok {
		return "没有找到位置设备"
	}

	f := func(message, aiMessage, deviceId string) string {
		var gShortAutomations = make(map[string]*shortScene)
		entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
		if !ok {
			return ""
		}

		for _, e := range entities {
			gShortAutomations[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		//发送所有自动化简短数据给ai
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role:    "user",
			Content: fmt.Sprintf(`这是我的全部自动化信息%s，我所在位置是：%s，根据对话内容选择最合适的场景返回给我。返回格式："id":""`, x.MustMarshalEscape2String(gShortAutomations), data.GetAreaName(device.AreaID)),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string

		for _, v := range gShortAutomations {
			if strings.Contains(result, v.id) {
				id = v.id
			}
		}

		if id == "" {
			return "没有找到你要的自动化信息"
		}

		if strings.Contains(aiMessage, "turn_on_automation") {
			intelligent.TurnOnAutomation(ava.Background(), id)
			return "自动化已开启"
		}

		if strings.Contains(aiMessage, "turn_off_automation") {
			intelligent.TurnOffAutomation(ava.Background(), id)
			return "自动化已关闭"
		}

		err = intelligent.RunAutomation(id)
		if err != nil {
			ava.Error(err)
			return "运行自动化失败了,错误信息是：" + err.Error()
		}

		return "自动化已经帮你运行了"
	}

	return CoreDelay(message, aiMessage, deviceId, f)
}
