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
	device := data.GetDevice(deviceId)
	if device == nil {
		return "没有找到位置设备"
	}

	f := func(message, aiMessage, deviceId string) string {
		var gShortAutomations = make(map[string]*shortScene)
		var golaAutomation = make(map[string]*shortScene)

		entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
		if !ok {
			return ""
		}

		for _, e := range entities {
			if strings.Contains(message, e.OriginalName) ||
				x.Similarity(message, e.OriginalName) > 0.8 ||
				x.ContainsAllChars(message, e.OriginalName) {

				golaAutomation[e.UniqueID] = &shortScene{
					id:    e.EntityID,
					Alias: e.OriginalName,
				}
			}

			gShortAutomations[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		if len(golaAutomation) > 0 {
			gShortAutomations = nil
			gShortAutomations = golaAutomation
		}

		var sendData = make([]string, 0, 2)
		for _, v := range gShortAutomations {
			sendData = append(sendData, v.Alias)
		}

		//发送所有自动化简短数据给ai
		result, err := chatCompletionInternal([]*chat.ChatMessage{{
			Role: "user",
			Content: fmt.Sprintf(`这是我的全部自动化信息%v，我所在位置是：%s，根据我的意图严格从我的自动化信息数据中选择名称返回给我。
返回格式：["名称1","名称2"]`, x.MustMarshalEscape2String(sendData), data.GetAreaName(device.AreaID)),
		}, {Role: "user", Content: message}})
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string
		var alias string

		for _, v := range gShortAutomations {
			if strings.Contains(result, v.Alias) {
				alias = v.Alias
				id = v.id
				break
			}
		}

		if id == "" {
			return "没有找到你要的自动化信息"
		}

		if strings.Contains(aiMessage, "turn_on_automation") {
			intelligent.TurnOnAutomation(ava.Background(), id)
			return alias + "已开启"
		}

		if strings.Contains(aiMessage, "turn_off_automation") {
			intelligent.TurnOffAutomation(ava.Background(), id)
			return alias + "已禁用"
		}

		err = intelligent.RunAutomation(id)
		if err != nil {
			ava.Error(err)
			return "运行自动化失败了,错误信息是：" + err.Error()
		}

		return alias + "运行了"
	}

	return CoreDelay(message, aiMessage, deviceId, f)
}
