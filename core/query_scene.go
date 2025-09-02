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

type shortScene struct {
	id    string
	Alias string `json:"alias"`
}

type script struct {
	Alias       string                   `json:"alias"`       //自动化名称
	Description string                   `json:"description"` //自动化描述
	Sequence    []map[string]interface{} `json:"sequence"`    //执行动作
}

func QueryScene(message, aiMessage, deviceId string) string {

	var gShortScenes = make(map[string]*shortScene)
	var golaScene = make(map[string]*shortScene)

	entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
	if !ok {
		return ""
	}

	for _, e := range entities {
		//判断是否有指定的场景
		if strings.Contains(message, e.OriginalName) ||
			x.Similarity(message, e.OriginalName) > 0.8 ||
			x.ContainsAllChars(message, e.OriginalName) ||
			x.ContainsAllChars(message, e.OriginalName) {
			golaScene[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}
		gShortScenes[e.UniqueID] = &shortScene{
			id:    e.EntityID,
			Alias: e.OriginalName,
		}
	}

	if len(golaScene) > 0 {
		gShortScenes = nil
		gShortScenes = golaScene
	}

	var sendData = make([]string, 0, 2)
	for _, v := range gShortScenes {
		sendData = append(sendData, v.Alias)
	}

	var content string

	if strings.Contains(aiMessage, "query_number") {
		content = fmt.Sprintf(`这是我的全部场景信息%v，我计算好了总数是%d个，根据我的意图用简洁、人性化的语言回答我。`, x.MustMarshalEscape2String(sendData), len(sendData))
	}

	if strings.Contains(aiMessage, "query_detail") {
		content = fmt.Sprintf(`这是我的全部场景信息%v，根据我的意图用简洁、人性化的语言回答我。
返回数据格式：["名称1","名称2"]`, x.MustMarshalEscape2String(sendData))
	}

	if content == "" {
		content = fmt.Sprintf(`这是我的全部场景信息%v，根据我的意图用简洁、人性化的语言回答我：
功能1:需要获取某个场景，返回["名称1","名称2"]
功能2:根据场景称统计自动化数量：例如："共有5个场景"`, x.MustMarshalEscape2String(sendData))
	}

	//发送所有场景简短数据给ai
	result, err := chatCompletionInternal([]*chat.ChatMessage{{
		Role:    "user",
		Content: content,
	}, {Role: "user", Content: message}})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一会儿再试试"
	}
	var id string

	for _, v := range gShortScenes {
		if strings.Contains(result, v.Alias) {
			id = v.id
			break
		}
	}

	if id == "" {
		return result
	}

	//拿到场景名称和id
	e, ok := data.GetEntityByEntityId()[id]
	if !ok {
		return "没有找到这个场景"
	}

	var scene = &script{}
	//获取场景信息
	err = intelligent.GetScript(e.UniqueID, scene)
	if err != nil {
		ava.Error(err)
		return "没有找到这个场景"
	}

	//找出场景下的设备名称
	for index, seq := range scene.Sequence {
		for k, v := range seq {
			if k == "entity_id" {
				if _, ok := v.(string); !ok {
					continue
				}
				ee, ok := data.GetEntityByEntityId()[v.(string)]
				if !ok {
					continue
				}
				scene.Sequence[index]["device_name"] = ee.DeviceName
			}
			if k == "device_id" {
				did := ""
				if did, ok = v.(string); !ok {
					continue
				}
				ee := data.GetDevice(did)
				if ee == nil {
					continue
				}
				scene.Sequence[index]["device_name"] = ee.Name
			}

			if k == "target" {

				v1, ok := v.(map[string]interface{})
				if !ok {
					continue
				}

				for k2, v2 := range v1 {
					if k2 == "entity_id" || k2 == "device_id" {
						v2Id := ""
						if v2Id, ok = v2.(string); !ok {
							continue
						}
						ee := data.GetDevice(v2Id)
						if ee == nil {
							continue
						}
						scene.Sequence[index]["device_name"] = ee.Name
					}
				}
			}
		}
	}

	//发送给ai
	msg, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf(`请根据场景信息用简洁、人性化的语言且不超过50字回答我，当前场景信息配置如下：%s`, x.MustMarshalEscape2String(scene)),
		},
		{
			Role:    "user",
			Content: message,
		},
	})
	if err != nil {
		ava.Error(err)
		return "场景信息处理失败了"
	}

	return msg
}
