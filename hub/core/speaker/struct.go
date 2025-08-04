package speaker

type ObjectLogic struct {
	Description string    `json:"description,omitempty"`
	Action      []string  `json:"action,omitempty"`
	Object      []*Object `json:"object,omitempty"`
}

type Object struct {
	Name       string `json:"name,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
	EntityId   string `json:"entity_id,omitempty"`
	DeviceId   string `json:"device_id,omitempty"`
}
