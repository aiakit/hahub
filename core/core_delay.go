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
	if strings.Contains(message, "trigger_time") {
		result, err := chatCompletion([]*chat.ChatMessage{
			{
				Role: "user",
				Content: fmt.Sprintf(`当前时间是%s,根据我的意图，返回延时执行动作的时间点，计算好是多少秒，并且按照格式进行返回，用人性化的语言描述content字段。
返回JSON格式：{"delay":60,"at":"22:50:50","content":"好的，我将在22:50:50操作x设备"}`, time.Now()),
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

	if strings.Contains(message, "trigger_scene") {
		//找到场景id
		var gShortScenes = make(map[string]*shortScene)

		entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		if !ok {
			return ""
		}

		for _, e := range entities {
			gShortScenes[e.UniqueID] = &shortScene{
				Id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		//发送所有场景简短数据给ai
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role: "user",
			Content: fmt.Sprintf(`这是我的全部场景信息%s，总数是%d个，根据对话内容将信息返回给我：
1.查询某个场景："id":""
2.查询场景数量："你总共有x个场景"`, x.MustMarshalEscape2String(gShortScenes), len(gShortScenes)),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string
		var alias string

		for _, v := range gShortScenes {
			if strings.Contains(result, v.Id) {
				alias = v.Alias
				id = v.Id
			}
		}

		if id == "" {
			return result
		}

		//启动协程
		registerRunScene(message, aiMessage, deviceId, id, f)

		return "已设置" + alias + "执行时触发"
	}

	if strings.Contains(message, "trigger_automation") {
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
		var alias string

		for _, v := range gShortAutomations {
			if strings.Contains(result, v.Id) {
				id = v.Id
				alias = v.Alias
			}
		}

		if id == "" {
			return result
		}

		//启动协程
		registerRunScene(message, aiMessage, deviceId, id, f)

		return "已设置" + alias + "执行时触发"
	}

	return f(message, aiMessage, deviceId)
}

func registerRunScene(message, AiMessage, deviceId, id string, ff FunctionHandler) {

	var done = make(chan struct{})
	fId := data.RegisterDataHandler(func(simple *data.StateChangedSimple, bytes []byte) {
		var state chatMessage
		err := x.Unmarshal(body, &state)
		if err != nil {
			ava.Error(err)
			return
		}

		if strings.EqualFold(state.Event.Data.EntityID, id) {
			ff(message, AiMessage, deviceId)
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
