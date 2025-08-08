package core

type ObjectLogic struct {
	Description  string            `json:"description"`
	Object       []*Object         `json:"object"`
	FunctionName string            `json:"function_name"`
	SubFunction  []subFunction     `json:"sub_function"`
	localKey     map[string]string //todo: 后期优化再做，用来跳过第一步向ai获取预处理动作
}

type subFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Object struct {
	Name       string `json:"name"`
	DeviceType string `json:"device_type"`
	EntityId   string `json:"entity_id"`
	DeviceId   string `json:"device_id"`
}
