package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

var gFunctionRouter *FunctionRouter

func CoreChaos() {
	gFunctionRouter = NewFunctionRouter()

	for _, v := range logicDataTwo {
		logicDataALL = append(logicDataALL, v)
		gFunctionRouter.Register(v.FunctionName, v.f)
	}
	for _, v := range logicDataOne {
		logicDataALL = append(logicDataALL, v)
		gFunctionRouter.Register(v.FunctionName, v.f)
	}

	gFunctionRouter.Register("开发中", IsInDevelopment)

	data.RegisterDataHandler(registerHomingWelcome)
	//data.RegisterDataHandler(registerToggleLight)

	chaosSpeaker()
}

// 定义函数处理类型
type FunctionHandler func(message, aiMessage, deviceId string) string

// 函数映射表结构
type FunctionRouter struct {
	handlers map[string]FunctionHandler
}

// 创建新的函数路由器
func NewFunctionRouter() *FunctionRouter {
	return &FunctionRouter{
		handlers: make(map[string]FunctionHandler),
	}
}

// 注册函数处理器
func (fr *FunctionRouter) Register(functionName string, handler FunctionHandler) {
	fr.handlers[functionName] = handler
}

// 根据函数名调用对应的函数
func Call(functionName, deviceId, message, aiMessage string) string {
	if handler, exists := gFunctionRouter.handlers[functionName]; exists {
		var now = time.Now()
		result := handler(message, aiMessage, deviceId)
		ava.Debugf("latency=%.2f |funcion_name=%s |message=%s |ai_message=%s |FROM=%s", time.Since(now).Seconds(), functionName, message, aiMessage, result)
		return result
	}
	// 如果没有找到对应的处理器，返回空字符串或错误信息
	return "未知指令"
}

func findFunction(message string) string {
	for _, v := range logicDataALL {
		if strings.Contains(message, v.FunctionName) {
			return v.FunctionName
		}
	}

	return "开发中"
}

var systemPrompts = `你是一个幽默的智能家居主管。你的中文名:小爱同学,英文名:jax，我给你配了一个助理，英文名字叫：jinx,中文名字：金克丝。你需要准分析并判断jinx是否解决了宿主的问题，如果没有需要你亲自处理。以下是我们最近的对话记录%s。`
var systemPromptsNone = `你是一个幽默的智能家居主管。你的中文名:小爱同学,英文名:jax，我给你配了一个助理英文名字叫：jinx,中文名字：金克丝。你需要准确分析并判断jinx是否解决了宿主的问题，如果没有需要你亲自处理。`

func chatCompletionInternal(msgInput []*chat.ChatMessage) (string, error) {
	var message = make([]*chat.ChatMessage, 0, 5)

	message = append(message, msgInput...)

	return chat.ChatCompletionMessage(message)
}

func chatCompletionHistory(msgInput []*chat.ChatMessage, deviceId string) (string, error) {
	history := GetHistory(deviceId)
	var message = make([]*chat.ChatMessage, 0, 5)

	var content = systemPromptsNone
	if len(history) > 0 {
		content = fmt.Sprintf(systemPrompts, x.MustMarshal2String(history))
	}

	message = append(message, &chat.ChatMessage{
		Role:    "system",
		Content: content,
	})

	message = append(message, msgInput...)

	result, err := chat.ChatCompletionMessage(message)
	if err != nil {
		return "发生未知错误，请重试", err
	}

	AddAIMessage(deviceId, result)

	return result, nil
}

func registerHomingWelcome(simple *data.StateChangedSimple, body []byte) {

	if !strings.Contains(simple.Event.Data.NewState.State, "on") {
		return
	}

	if (strings.HasPrefix(simple.Event.Data.NewState.EntityID, "automation.") || strings.HasPrefix(simple.Event.Data.NewState.EntityID, "script.")) &&
		strings.Contains(simple.Event.Data.NewState.Attributes.FriendlyName, "回家") {
		result, err := chat.ChatCompletionMessage([]*chat.ChatMessage{{
			Role:    "user",
			Content: "你是一个智能家居系统，我是你的宿主，我现在回家了，你得想一句俏皮话欢迎我。",
		}})

		if err != nil {
			ava.Error(err)
			return
		}

		entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
		if !ok {
			return
		}

		var id string
		for _, e := range entities {
			if strings.Contains(e.EntityID, "media_player.") && strings.Contains(e.AreaName, "客厅") {
				id = e.EntityID
				break
			}
		}

		// 添加检查，防止id为空
		if id == "" {
			ava.Warn("No suitable media player found for welcome message")
			return
		}

		aiLock(id)

		err = x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
			EntityId: id,
			Value:    result,
		}, nil)
		if err != nil {
			ava.Error(err)
		}

		x.TimingwheelAfter(GetPlaybackDuration(result), func() {
			aiUnlock(id)
		})
	}
}

// 语音指令关灯之后，就不要再自动开灯,直到语音指令开灯
func registerToggleLight(simple *data.StateChangedSimple, body []byte) {
	var state chatMessage
	err := x.Unmarshal(body, &state)
	if err != nil {
		ava.Error(err)
		return
	}

	if strings.Contains(state.Event.Data.EntityID, "_conversation") &&
		strings.EqualFold(state.Event.Data.NewState.Attributes.EntityClass, "XiaoaiConversationSensor") {
		//找到所有根据存在传感器有人亮灯的自动化
		if (strings.Contains(simple.Event.Data.NewState.State, "关") && strings.Contains(simple.Event.Data.NewState.State, "灯")) ||
			strings.Contains(simple.Event.Data.NewState.State, "晚安") || strings.Contains(simple.Event.Data.NewState.State, "睡觉") || strings.Contains(simple.Event.Data.NewState.State, "午觉") {

			entity, ok := data.GetEntityByEntityId()[simple.Event.Data.EntityID]
			if !ok {
				return
			}
			areaName := data.SpiltAreaName(entity.AreaName)

			//查询所有自动化
			as, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
			if !ok {
				return
			}

			for _, a := range as {
				if strings.Contains(a.OriginalName, areaName) && strings.Contains(a.OriginalName, "有人亮灯") {
					//关闭自动化
					err = intelligent.TurnOffAutomation(ava.Background(), a.EntityID)
					if err != nil {
						ava.Error(err)
						return
					}
				}
			}
		}

		if (strings.Contains(simple.Event.Data.NewState.State, "开") && strings.Contains(simple.Event.Data.NewState.State, "灯")) ||
			strings.Contains(simple.Event.Data.NewState.State, "起床") || strings.Contains(simple.Event.Data.NewState.State, "早安") {
			{
				entity, ok := data.GetEntityByEntityId()[simple.Event.Data.EntityID]
				if !ok {
					return
				}
				areaName := data.SpiltAreaName(entity.AreaName)

				//查询所有自动化
				as, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
				if !ok {
					return
				}

				for _, a := range as {
					if strings.Contains(a.OriginalName, areaName) && strings.Contains(a.OriginalName, "有人亮灯") {
						err = intelligent.TurnOnAutomation(ava.Background(), a.EntityID)
						if err != nil {
							ava.Error(err)
							return
						}
					}
				}
			}
		}
	}
}
