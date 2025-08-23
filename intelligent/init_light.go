package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

var lightGradientTime = &Script{
	Alias:       "初始化灯光模式设置-慎用",
	Description: "初始化灯光模式设置，包括渐变时间、默认状态等配置",
	Sequence:    []interface{}{},
}

var slowestSetting = &Script{
	Alias:       "初始化最低亮度设置-慎用",
	Description: "设置灯光最低亮度参数",
	Sequence:    []interface{}{},
}

func init() {
	data.RegisterEntityCallback(registerLightGradientTime)
	data.RegisterEntityCallback(lowestBrightness)
	data.RegisterDataHandler(registerVirtualEvent)
}

// 初始化灯光
// 创建灯光初始化
func InitLight(c *ava.Context) {
	if len(slowestSetting.Sequence) > 0 {
		AddScript2Queue(c, slowestSetting)
	}

	if len(lightGradientTime.Sequence) > 0 {
		AddScript2Queue(c, lightGradientTime)
	}
}

// 最低亮度设置
func lowestBrightness(entity *data.Entity) {
	// 为每个实体添加动作到场景序列中
	if strings.Contains(entity.OriginalName, "最低亮度(x10)") && strings.HasPrefix(entity.EntityID, "number.") {
		slowestSetting.Sequence = append(slowestSetting.Sequence, ActionCommon{
			Type:     "set_value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "number",
			Value:    8,
		})
		return
	}

	if strings.Contains(entity.OriginalName, "最低亮度") && strings.HasPrefix(entity.EntityID, "number.") {
		slowestSetting.Sequence = append(slowestSetting.Sequence, ActionCommon{
			Type:     "set_value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "number",
			Value:    1,
		})
	}
}

// 灯光时间设置
func registerLightGradientTime(entity *data.Entity) {

	//判断区域，带流动
	//柜子灯带
	if (strings.Contains(entity.OriginalName, "开灯渐变时长(单位ms)") || strings.Contains(entity.OriginalName, "灯光调光时长(单位ms)") || strings.Contains(entity.OriginalName, "关灯渐变时长(单位ms)")) && strings.Contains(entity.EntityID, "number.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "set_value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "number",
			Value:    5000,
		})
	}

	if (strings.Contains(entity.OriginalName, "通电默认状态") || strings.Contains(entity.OriginalName, "分段开关") || strings.Contains(entity.OriginalName, "恢复断电前灯光")) && strings.Contains(entity.EntityID, "switch.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "turn_off",
			EntityID: entity.EntityID,
			DeviceID: entity.DeviceID,
			Domain:   "switch",
		})
	}

	if strings.Contains(entity.OriginalName, "指数调光") && strings.HasPrefix(entity.EntityID, "switch.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "turn_on",
			EntityID: entity.EntityID,
			DeviceID: entity.DeviceID,
			Domain:   "switch",
		})
	}

	// 提取为独立函数处理默认状态渐变时间设置
	actionCommon := handleDefaultGradientTimeSettings(entity, 1)
	if actionCommon != nil {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, actionCommon)
	}

	if strings.Contains(entity.OriginalName, "默认上电状态") && strings.Contains(entity.EntityID, "select.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "select_option",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "select",
			Option:   "上电关闭",
		})
	}

	if strings.Contains(entity.OriginalName, "默认状态 灯光变化") && strings.HasPrefix(entity.EntityID, "select.") {

		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "select_option",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "select",
			Option:   "Gradient",
		})
	}

	if strings.Contains(entity.OriginalName, "默认状态 默认灯光") && strings.HasPrefix(entity.EntityID, "select.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "select_option",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "select",
			Option:   "OFF",
		})
	}

	//馨光灯带，关灯断电
	if strings.Contains(entity.OriginalName, "关灯断电") && strings.HasPrefix(entity.EntityID, "select.") {
		lightGradientTime.Sequence = append(lightGradientTime.Sequence, ActionCommon{
			Type:     "select_option",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "select",
			Option:   "断电",
		})
	}
}

// 处理默认状态渐变时间设置的独立函数
func handleDefaultGradientTimeSettings(entity *data.Entity, option int) *ActionCommon {
	if strings.Contains(entity.OriginalName, "默认状态 渐变时间设置，字节[0]开灯渐变时间，字节[1]关灯渐变时间，字节[2]模式渐变时间") && !strings.Contains(entity.OriginalName, "字节3（配置渐变、默认灯光、配置灯光、灯光变化、配置变化）") {
		var value = "4278853"
		switch option {
		case 1:
			value = "4278853" //5,10,1
		case 2:
			value = "4278787" //300ms,10,1
		}

		return &ActionCommon{
			Type:     "set_value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "number",
			Value:    value,
		}
	}

	if strings.Contains(entity.OriginalName, "字节3（配置渐变、默认灯光、配置灯光、灯光变化、配置变化）") {
		var value = "21056069"
		switch option {
		case 1:
			value = "21056069" //5,10,1
		case 2:
			value = "21056003" //300ms,10,1
		}

		return &ActionCommon{
			Type:     "set_value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "number",
			Value:    value,
		}
	}

	return nil
}
