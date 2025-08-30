package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

// eg.当我离开家后，打开扫地机器人
// eg.5分钟后，打开扫地机器人
// 延时执行的动作：设备，场景，自动化
func CoreDelay(message, aiMessage, deviceId string, f FunctionHandler) string {
	//1.解析指令
	if strings.Contains(aiMessage, "trigger_time") {
		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "user",
				Content: fmt.Sprintf(`当前时间是%s,根据我的意图，返回延时执行动作的时间点，计算好是多少秒，并且按照格式进行返回，用人性化的语言描述content字段。
返回JSON格式：{"delay":60,"at":"22:50:50","content":"好的，我将在22:50:50执行"}`, time.Now()),
			},
			{
				Role:    "user",
				Content: message,
			},
		})

		if err != nil {
			ava.Error(err)
			return "人功智能有点混乱了"
		}

		var delay = struct {
			Delay   int    `json:"delay"`
			At      string `json:"at"`
			Content string `json:"content"`
		}{}

		err = x.Unmarshal([]byte(result), &delay)
		if err != nil {
			ava.Error(err)
			return "服务器出错了"
		}

		x.TimingwheelAfter(time.Second*time.Duration(delay.Delay), func() {
			f(message, aiMessage, deviceId)
		})

		return delay.Content
	}

	if strings.Contains(aiMessage, "trigger_scene") {
		//找到场景id
		var gShortScenes = make(map[string]*shortScene)
		var gShortScripts = make(map[string]*shortScene)

		entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		if !ok {
			return ""
		}

		for _, e := range entities {
			//判断是否有指定的场景
			if strings.Contains(message, e.OriginalName) ||
				x.Similarity(message, e.OriginalName) > 0.8 ||
				x.ContainsAllChars(message, e.OriginalName) {

				gShortScenes[e.UniqueID] = &shortScene{
					id:    e.EntityID,
					Alias: e.OriginalName,
				}
				continue
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
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role:    "user",
			Content: fmt.Sprintf(`这是我的全部场景信息%v，根据我的意图找到作为触发条件的场景名称返回给我，名称必须严格从我给的场景信息中获取。返回格式："名称1"`, sendData),
		}, {Role: "user", Content: message}}, deviceId)
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
			return "没有找到这个场景"
		}

		//启动协程
		go registerRunScene(message, aiMessage, deviceId, id, alias, f)

		return "已设置" + alias + "执行时触发"
	}

	if strings.Contains(aiMessage, "trigger_automation") {
		var gShortAutomations = make(map[string]*shortScene)
		var golaAutomation = make(map[string]*shortScene)
		entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]

		if !ok {
			return ""
		}

		for _, e := range entities {
			if strings.Contains(message, e.OriginalName) || x.Similarity(message, e.OriginalName) > 0.8 {
				golaAutomation[e.UniqueID] = &shortScene{
					id:    e.EntityID,
					Alias: e.OriginalName,
				}
				continue
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
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role:    "user",
			Content: fmt.Sprintf(`这是我的全部自动化信息%v，根据我的意图将作为触发条件的名称告诉我，名称必须严格从我给的场景信息中获取。返回格式："名称1"`, sendData),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string
		var alias string
		for _, v := range gShortAutomations {
			if strings.Contains(result, v.Alias) {
				id = v.id
				alias = v.Alias
				break
			}
		}

		if id == "" {
			return "没有找到这个自动化"
		}

		//启动协程
		go registerRunScene(message, aiMessage, deviceId, id, alias, f)

		return "已设置" + alias + "执行时触发"
	}

	return f(message, aiMessage, deviceId)
}

func registerRunScene(message, AiMessage, deviceId, id, trigger string, ff FunctionHandler) {

	var done = make(chan struct{})
	fId := data.RegisterDataHandler(func(simple *data.StateChangedSimple, bytes []byte) {
		ava.Debugf("registerRunScene |data=%s", x.MustMarshal2String(simple))
		var state chatMessage
		err := x.Unmarshal(bytes, &state)
		if err != nil {
			ava.Error(err)
			return
		}

		if strings.EqualFold(state.Event.Data.EntityID, id) {
			runMessage := ff(message, AiMessage, deviceId)
			ava.Debugf("trigger=%s |action=%s", trigger, runMessage)
			done <- struct{}{}
		}
	})

	go func() {
		closeId := fId
		select {
		case <-done:
			data.UnregisterDataHandler(closeId)
		case <-time.After(time.Hour * 24 * 7): //最多保留7天
			data.UnregisterDataHandler(closeId)
		}
	}()
}
