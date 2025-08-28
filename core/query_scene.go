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
	Id    string `json:"id"`
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
		if strings.Contains(message, e.OriginalName) || x.Similarity(message, e.Name) > 0.8 {
			golaScene[e.UniqueID] = &shortScene{
				Id:    e.EntityID,
				Alias: e.OriginalName,
			}
			continue
		}
		gShortScenes[e.UniqueID] = &shortScene{
			Id:    e.EntityID,
			Alias: e.OriginalName,
		}
	}

	if len(golaScene) > 0 {
		gShortScenes = nil
		gShortScenes = golaScene
	}

	var sendData = make([]*shortScene, 0, 2)
	for _, v := range gShortScenes {
		sendData = append(sendData, v)
	}

	//发送所有场景简短数据给ai
	result, err := chatCompletionInternal([]*chat.ChatMessage{{
		Role: "user",
		Content: fmt.Sprintf(`这是我的全部场景信息%s，总数是%d个，根据对话内容将信息返回给我，内容必须简洁，不超过30个字：
1.查询某个场景："id":""
2.查询场景数量："共有x个场景"`, x.MustMarshalEscape2String(sendData), len(sendData)),
	}, {Role: "user", Content: message}})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一会儿再试试"
	}
	var id string

	for _, v := range gShortScenes {
		if strings.Contains(result, v.Id) {
			id = v.Id
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
				if _, ok := v.(string); !ok {
					continue
				}
				ee, ok := data.GetDevice()[v.(string)]
				if !ok {
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
						if _, ok := v2.(string); !ok {
							continue
						}
						ee, ok := data.GetDevice()[v2.(string)]
						if !ok {
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
			Content: fmt.Sprintf(`请根据场景信息用总结、归纳性的人性化的语言回答我，当前场景信息配置如下：%s`, x.MustMarshalEscape2String(scene)),
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
