package core

type ObjectLogic struct {
	Description  string        `json:"description"`
	Object       []*Object     `json:"object"`
	FunctionName string        `json:"function_name,omitempty"`
	SubFunction  []subFunction `json:"sub_function"`
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
