package core

import (
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
)

// 功能预处理,AI对json的组装效果不好，用文字返回代替处理,无论AI返回什么样的格式，用字符串去进行一起匹配
// 板块名称：动作：对象
// eg.场景，执行场景，回家
// eg.天气，查询天气，小爱入口忽略/云端查询
const (
	query_scene                         = "query_scene"
	run_scene                           = "run_scene"
	run_automation                      = "run_automation"
	query_automation                    = "query_automation"
	run_device                          = "run_device"
	query_device_state                  = "query_device_state"
	query_device_number                 = "query_device_number"
	delayed_task                        = "delayed_task"
	scheduled_task                      = "scheduled_task"
	daily_conversation                  = "daily_conversation"
	function_call_init_scene            = "function_call_init_scene"
	function_call_init_automation       = "function_call_init_automation"
	function_call_init_light            = "function_call_init_light"
	function_call_init_switch           = "function_call_init_switch"
	function_call_init_lowest_light     = "function_call_init_lowest_light"
	function_call_init_lighting_control = "function_call_init_lighting_control"
	is_handled                          = "is_handled"
	is_in_development                   = "is_in_development"
	is_done                             = "is_done"
)

func init() {

	logicDataMap[query_scene] = &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发查询智能场景信息",
		FunctionName: "查询场景",
	}

	logicDataMap[run_scene] = &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发运行智能场景",
		FunctionName: "执行场景",
	}

	logicDataMap[run_automation] = &ObjectLogic{
		Description:  "运行通过一些自然条件（例如：天气、温度、湿度、时间等）或者设备条件（例如：水浸传感器、人体传感器等设备状态）或者某个事件触发（例如：当执行睡觉场景之后就关闭人来亮灯自动化）对智能家居设备一系列的操作",
		FunctionName: "执行自动化",
	}

	logicDataMap[query_automation] = &ObjectLogic{
		Description:  "查询通过一些自然条件（例如：天气、温度、湿度、时间等）或者设备条件（例如：水浸传感器、人体传感器等设备状态）或者某个事件触发（例如：当执行睡觉场景之后就关闭人来亮灯自动化）对智能家居设备一系列的操作",
		FunctionName: "查询自动化",
	}

	//platform不等于xiaomi_home的设备需要AI操作,例如热水器等
	logicDataMap[run_device] = &ObjectLogic{
		Description:  "对智能家居设备进行控制",
		FunctionName: "操作设备",
	}

	logicDataMap[query_device_state] = &ObjectLogic{
		Description:  "查询智能家居设备的运行状态",
		FunctionName: "查询设备状态",
	}

	logicDataMap[query_device_number] = &ObjectLogic{
		Description:  "查询智能家居设备数量信息",
		FunctionName: "查询设备数量",
	}

	logicDataMap[delayed_task] = &ObjectLogic{
		Description:  "创建一个延时的任务",
		FunctionName: "延时任务",
	}

	logicDataMap[scheduled_task] = &ObjectLogic{
		Description:  "创建一个定时任务",
		FunctionName: "任务",
	}

	logicDataMap[daily_conversation] = &ObjectLogic{
		Description:  "非智能家居所在领域的对话",
		FunctionName: "日常对话",
	}

	logicDataMap[function_call_init_scene] = &ObjectLogic{
		Description:  "执行初始化场景函数调用",
		FunctionName: "初始化场景函数调用",
	}

	logicDataMap[function_call_init_automation] = &ObjectLogic{
		Description:  "执行初始化自动化函数调用",
		FunctionName: "初始化自动化函数调用",
	}
	logicDataMap[function_call_init_light] = &ObjectLogic{
		Description:  "执行初始化灯具参数函数调用",
		FunctionName: "初始化灯具参数函数调用",
	}
	logicDataMap[function_call_init_switch] = &ObjectLogic{
		Description:  "执行初始化开关参数函数调用",
		FunctionName: "初始化开关参数函数调用",
	}
	logicDataMap[function_call_init_lowest_light] = &ObjectLogic{
		Description:  "执行初始化最低亮度函数调用",
		FunctionName: "初始化最低亮度函数调用",
	}
	logicDataMap[function_call_init_lighting_control] = &ObjectLogic{
		Description:  "执行初始化开关对灯设备的控制设置",
		FunctionName: "初始化灯控函数调用",
	}

	logicDataMap[is_handled] = &ObjectLogic{
		Description:  "在最近一次对话中，assistant的回复是已经处理了的情况，返回这个对象，防止重复处理",
		FunctionName: "已经处理过一次",
	}

	logicDataMap[is_in_development] = &ObjectLogic{
		Description:  "在我们的对话中，如果没有找到对应功能，就返回这个对象",
		FunctionName: "功能开发中",
	}

	logicDataMap[is_done] = &ObjectLogic{
		Description:  "根据对话上下文判断是否需要终止对话",
		FunctionName: "终止对话",
	}
}

var logicDataMap = make(map[string]*ObjectLogic)

// 预调用提示
var preparePrompts = `根据对话内容，以及我提供的一些功能选项，判断我的意图选择需要执行什么功能，并按照规定的格式返回数据，除了返回的数据格式，禁止有其他内容。
功能选项：%s
返回数据格式：{"功能模块":"功能名称"}
返回数据例子：{"query_scene":"查询场景"}`

// todo: 加入当前对话位置名称，方便操作对应位置的设备
func prepareCall(messageInput []*chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "user", Content: fmt.Sprintf(preparePrompts, x.MustMarshalEscape2String(logicDataMap))})

	if len(messageInput) > 0 {
		messageList = append(messageList, messageInput...)
	}

	return chatCompletionHistory(messageList, deviceId)
}
