package hub

type wsCallService struct {
	Id          int64       `json:"id"`
	Type        string      `json:"type"`
	Domain      string      `json:"domain"`
	Service     string      `json:"service"`
	ServiceData interface{} `json:"service_data"`
	Target      struct {
		EntityId string `json:"entity_id"`
	} `json:"target"`

	ReturnResponse bool `json:"return_response"`
}

type stateData struct {
	Type  string `json:"type"`
	Event struct {
		EventType string `json:"event_type"`
		Data      struct {
			NewState struct {
				EntityID     string `json:"entity_id"`
				State        string `json:"state"` //语音内容
				LastChanged  string `json:"last_changed"`
				LastReported string `json:"last_reported"`
				LastUpdated  string `json:"last_updated"`
				Attributes   struct {
					Timestamp    string `json:"timestamp"`
					DeviceClass  string `json:"device_class"`  //motion运动传感器
					FriendlyName string `json:"friendly_name"` //设备名称
				} `json:"attributes"`
			} `json:"new_state"`
		} `json:"data"`
		TimeFired string `json:"time_fired"`
	} `json:"event"`
}
