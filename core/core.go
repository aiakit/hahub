package core

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
func (fr *FunctionRouter) Call(functionName string) string {
	if handler, exists := fr.handlers[functionName]; exists {
		return handler(functionName)
	}
	// 如果没有找到对应的处理器，返回空字符串或错误信息
	return ""
}

// 默认的函数处理器示例
func DefaultHandler(functionName string) string {
	return "Handling function: " + functionName
}