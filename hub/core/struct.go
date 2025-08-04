package speaker

type ObjectLogic struct {
	Description  string    `json:"description"`
	Action       []string  `json:"action"`
	Object       []*Object `json:"object"`
	FunctionName string    `json:"function_name"`
}

type Object struct {
	Name       string `json:"name"`
	DeviceType string `json:"device_type"`
	EntityId   string `json:"entity_id"`
	DeviceId   string `json:"device_id"`
}
