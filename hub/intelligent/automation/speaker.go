package automation

import "fmt"

//小米音箱
//1.播放文本
//2.执行文本指令

// PlayText 播放文本
func PlayText(deviceID string, text string) *ActionNotify {
	// 实现播放文本的逻辑
	return &ActionNotify{
		Action: "notify.send_message",
		Data: struct {
			Message string `json:"message,omitempty"`
			Title   string `json:"title,omitempty"`
		}{Message: text, Title: ""},
		Target: struct {
			DeviceID string `json:"device_id,omitempty"`
		}{DeviceID: deviceID},
	}
}

// ExecuteTextCommand 执行文本指令
func ExecuteTextCommand(entityId string, command string, silent bool) *ActionService {
	// 实现执行文本指令的逻辑
	return &ActionService{
		Action: "text.set_value",
		Data:   map[string]interface{}{"value": fmt.Sprintf("[%s,%v]", command, silent)},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: entityId},
	}
}
