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
	localAreaName := getAreaName(deviceId)
	if localAreaName == "" {
		return "你得给我分配一个区域，我才好判断你的位置"
	}

	f := func(message, aiMessage, deviceId string) string {
		var gShortScripts = make(map[string]*shortScene)
		var gShortScenes = make(map[string]*shortScene)

		entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		if !ok {
			return ""
		}

		for _, e := range entities {
			if strings.Contains(e.OriginalName, "视图") {
				continue
			}

			if runFilterEntityByArea(localAreaName, e) {
				continue
			}

			if strings.Contains(message, e.OriginalName) ||
				x.Similarity(message, e.OriginalName) > 0.8 ||
				x.ContainsAllChars(message, e.OriginalName) {

				gShortScenes[e.UniqueID] = &shortScene{
					id:    e.EntityID,
					Alias: e.OriginalName,
				}
			}
			gShortScripts[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		if len(gShortScenes) > 0 {
			gShortScripts = nil
			gShortScripts = gShortScenes
		}

		var sendData = make([]string, 0, 2)
		for _, v := range gShortScripts {
			sendData = append(sendData, v.Alias)
		}

		//发送所有场景简短数据给ai
		result, err := chatCompletionInternal([]*chat.ChatMessage{{
			Role:    "user",
			Content: fmt.Sprintf(`场景数据%v，当前我的位置是%s。根据我的意图严格地从场景信息中选择我要的场景名称返回给我。返回格式: ["名称1","名称2"]`, x.MustMarshal2String(sendData), localAreaName),
		}, {Role: "user", Content: message}})
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string
		var alias string

		for _, v := range gShortScripts {
			if strings.Contains(result, v.Alias) {
				alias = v.Alias
				id = v.id
				break
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

		return alias + "运行了"
	}

	return CoreDelay(message, aiMessage, deviceId, f)
}
