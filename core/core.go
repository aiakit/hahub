package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
)

var gFunctionRouter *FunctionRouter

func CoreChaos() {
	gFunctionRouter = NewFunctionRouter()

	gFunctionRouter.Register(functionCallInitAll, InitALL)
	gFunctionRouter.Register(isHandled, IsDone)
	gFunctionRouter.Register(functionCallInitAll, InitScene)
	gFunctionRouter.Register(functionCallInitAll, InitAutomation)
	gFunctionRouter.Register(queryScene, QueryScene)
	gFunctionRouter.Register(queryAutomation, QueryAutomation)
	gFunctionRouter.Register(queryDevice, QueryDevice)
	gFunctionRouter.Register(sendMessage2Speaker, SendMessagePlay)
	gFunctionRouter.Register(isInDevelopment, IsInDevelopment)
	gFunctionRouter.Register(runAutomation, RunAutomation)
	gFunctionRouter.Register(runScene, RunScene)
	gFunctionRouter.Register(controlDevice, RunDevice)
	gFunctionRouter.Register(timingTask, RunTming)
	gFunctionRouter.Register(dailyConversation, Conversation)

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
		return handler(message, aiMessage, deviceId)
	}
	// 如果没有找到对应的处理器，返回空字符串或错误信息
	return "未知指令"
}

func findFunction(message string) string {
	for k := range logicDataMap {
		if strings.Contains(message, k) {
			return k
		}
	}

	return "is_in_development"
}

var systemPrompts = `你是一个智能家居助理，你的中文名:小爱同学,英文名:jax，和你共同工作的另一个AI助理的英文名字叫：jinx,中文名字：金克丝。你的所有回答必须简洁，以下是我们最近的对话记录%s。`
var systemPromptsNone = `你是一个智能家居助理音箱，你的中文名:小爱同学,英文名:jax，和你共同工作的另一个AI助理的英文名字叫：jinx,中文名字：金克丝。你的所有回答必须简洁。`

func chatCompletion(msgInput []*chat.ChatMessage) (string, error) {
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

	return chat.ChatCompletionMessage(message)
}
