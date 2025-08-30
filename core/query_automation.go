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

func QueryAutomation(message, aiMessage, deviceId string) string {

	var gShortAutomations = make(map[string]*shortScene)
	var golaAutomation = make(map[string]*shortScene)

	entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
	if !ok {
		return ""
	}

	for _, e := range entities {
		//判断是否有指定的自动化
		//判断是否有指定的场景
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

	var content string

	if strings.Contains(aiMessage, "query_number") {
		content = fmt.Sprintf(`这是我的全部自动化信息%s，我计算好了总数是%d个，根据我的意图用简洁、人性化的语言回答我。`, x.MustMarshalEscape2String(sendData), len(sendData))
	}

	if strings.Contains(aiMessage, "query_detail") {
		content = fmt.Sprintf(`这是我的全部自动化信息%v，根据我的意图用简洁、人性化的语言回答我。
返回数据格式：["名称1","名称2"]`, x.MustMarshalEscape2String(sendData))
	}

	if content == "" {
		content = fmt.Sprintf(`这是我的全部自动化信息%v，根据我的意图用简洁、人性化的语言回答我：
功能1:需要获取某个自动化，返回["名称1","名称2"]
功能2:根据自动化名称统计自动化数量：例如："共有5个自动化"`, x.MustMarshalEscape2String(sendData))
	}

	//发送所有自动化简短数据给ai
	result, err := chatCompletionInternal([]*chat.ChatMessage{{
		Role:    "user",
		Content: content,
	}, {Role: "user", Content: message}})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请等一会儿再试试"
	}
	var id string

	for _, v := range gShortAutomations {
		if strings.Contains(result, v.Alias) {
			id = v.id
			break
		}
	}

	if id == "" {
		return result
	}

	//拿到自动化名称和id
	e, ok := data.GetEntityByEntityId()[id]
	if !ok {
		return "没有找到这个自动化"
	}

	var automation = &intelligent.Automation{}
	//获取自动化信息
	err = intelligent.GetAutomation(e.UniqueID, automation)
	if err != nil {
		ava.Error(err)
		return "没有找到这个自动化"
	}

	//找出自动化下的设备名称
	for index, seq := range automation.Triggers {
		ee, ok := data.GetEntityByEntityId()[seq.EntityID]
		if !ok {
			continue
		}
		automation.Triggers[index].Name = ee.DeviceName
	}

	for index, seq := range automation.Conditions {
		ee, ok := data.GetEntityByEntityId()[seq.EntityID]
		if !ok {
			continue
		}
		automation.Conditions[index].Name = ee.DeviceName
	}

	// 处理动作中的设备名称
	for index, seq := range automation.Actions {
		var dd, ok = seq.(map[string]interface{})
		if !ok {
			continue
		}
		for k, v := range dd {
			if k == "entity_id" {
				if _, ok := v.(string); !ok {
					continue
				}
				ee, ok := data.GetEntityByEntityId()[v.(string)]
				if !ok {
					continue
				}
				dd["device_name"] = ee.DeviceName
				automation.Actions[index] = dd
			}
			if k == "device_id" {
				if _, ok := v.(string); !ok {
					continue
				}
				ee, ok := data.GetDevice()[v.(string)]
				if !ok {
					continue
				}
				dd["device_name"] = ee.Name
				automation.Actions[index] = dd
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
						dd["device_name"] = ee.Name
						automation.Actions[index] = dd
					}
				}
			}
		}
	}

	//发送给ai
	msg, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf(`请根据自动化信息用简洁、人性化的语言且不超过50字回答我，当前自动化信息配置如下：%s`, x.MustMarshalEscape2String(automation)),
		},
		{
			Role:    "user",
			Content: message,
		},
	})
	if err != nil {
		ava.Error(err)
		return "自动化信息处理失败了"
	}

	return msg
}
