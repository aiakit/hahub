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
	entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
	if !ok {
		return ""
	}

	for _, e := range entities {
		gShortAutomations[e.UniqueID] = &shortScene{
			Id:    e.EntityID,
			Alias: e.OriginalName,
		}
	}

	//发送所有自动化简短数据给ai
	result, err := chatCompletionHistory([]*chat.ChatMessage{{
		Role: "user",
		Content: fmt.Sprintf(`这是我的全部自动化信息%s，总数是%d个，根据对话内容将信息返回给我：
1.查询某个自动化："id":""
2.查询自动化数量："你总共有x个自动化"`, x.MustMarshalEscape2String(gShortAutomations), len(gShortAutomations)),
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
		return result
	}

	//拿到自动化名称和id
	e, ok := data.GetEntityIdMap()[id]
	if !ok {
		return "没有找到这个设备"
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
		ee, ok := data.GetEntityIdMap()[seq.EntityID]
		if !ok {
			continue
		}
		automation.Triggers[index].Name = ee.DeviceName
	}

	for index, seq := range automation.Conditions {
		ee, ok := data.GetEntityIdMap()[seq.EntityID]
		if !ok {
			continue
		}
		automation.Conditions[index].Name = ee.DeviceName
	}

	//发送给ai
	msg, err := chatCompletionHistory([]*chat.ChatMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf(`请根据自动化信息用总结、归纳性的人性化的语言回答我，当前自动化信息配置如下：%s`, x.MustMarshalEscape2String(automation)),
		},
		{
			Role:    "user",
			Content: message,
		},
	}, deviceId)
	if err != nil {
		ava.Error(err)
		return "自动化信息处理失败了"
	}

	return msg
}
