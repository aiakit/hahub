package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

var gFunctionRouter *FunctionRouter

func init() {
	gFunctionRouter = NewFunctionRouter()

	ChaosSpeaker()
}

// 定义函数处理类型
type FunctionHandler func(functionName string) string

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
func (fr *FunctionRouter) Register(name string, handler FunctionHandler) {
	fr.handlers[name] = handler
}

// 根据函数名调用对应的函数
func Call(functionName string) string {
	if handler, exists := gFunctionRouter.handlers[functionName]; exists {
		return handler(functionName)
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

var systemPrompts = `你是一个智能家居助理音箱，名字叫做'小爱同学'，以下是我们最近的对话记录%s。`

func chatCompletion(deviceId string, msgInput []*chat.ChatMessage) (string, error) {
	history := GetHistory(deviceId)
	var message = make([]*chat.ChatMessage, 0, 5)

	if len(history) == 0 {
		history = []*chat.ChatMessage{}
	}

	message = append(message, &chat.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf(systemPrompts, x.MustMarshal2String(history)),
	})

	message = append(message, msgInput...)
	ava.Debugf("req=%s", x.MustMarshal2String(message))

	return chat.ChatCompletionMessage(message)
}
