package core

type ObjectLogic struct {
	Description  string        `json:"description"`
	FunctionName string        `json:"function_name,omitempty"`
	SubFunction  []subFunction `json:"sub_function"`
	f            func(message, aiMessage, deviceId string) string
}

type subFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
