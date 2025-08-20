package intelligent

import "fmt"

//通知模块
//1.app内持久化通知
//2.音箱设备通知
//3.手机app弹窗通知
//4.向云端发送通知

//小米音箱
//1.播放文本
//2.执行文本指令

// PlayText 播放文本
func PlayText(entityId string, text string) *ActionService {
	// 实现执行文本指令的逻辑
	return &ActionService{
		Action: "notify.send_message",
		Data:   map[string]interface{}{"message": text},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: entityId},
	}
}

// ExecuteTextCommand 执行文本指令
func ExecuteTextCommand(DeviceId string, command string, silent bool) *ActionNotify {
	// 实现执行文本指令的逻辑
	return &ActionNotify{
		Action: "notify.send_message",
		Data: struct {
			Message string `json:"message,omitempty"`
			Title   string `json:"title,omitempty"`
		}{Message: fmt.Sprintf("[%s,%v]", command, silent)},
		Target: struct {
			DeviceID string `json:"device_id,omitempty"`
		}{DeviceID: DeviceId},
	}
}
