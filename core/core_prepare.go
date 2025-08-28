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
	queryScene          = "query_scene"
	runScene            = "run_scene"
	runAutomation       = "run_automation"
	queryAutomation     = "query_automation"
	controlDevice       = "control_device"
	queryDevice         = "query_device"
	timingTask          = "timing_task"
	dailyConversation   = "daily_conversation"
	functionCallInitAll = "function_call_init_all"
	isHandled           = "is_handled"
	sendMessage2Speaker = "send_message_to_speaker"
	evaluate            = "evaluate"
	display             = "display"
)

func init() {

	logicDataMap[queryScene] = &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发查询智能场景信息",
		FunctionName: "查询智能家居场景",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询场景数量"},
			{Name: "query_detail", Description: "查询场景详情"},
		},
	}

	logicDataMap[queryAutomation] = &ObjectLogic{
		Description:  "查询智能家居自动化，通过触发条件自动控制某些设备的流程",
		FunctionName: "查询智能家居自动化",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询自动化数量"},
			{Name: "query_detail", Description: "查询自动化详情"},
		},
	}

	logicDataMap[evaluate] = &ObjectLogic{
		Description:  "对当前智能家居系统完善情况进行等级评估",
		FunctionName: "智能家居完善度评估",
	}

	logicDataMap[display] = &ObjectLogic{
		Description:  "执行智能家居演示模式，在运行过程中，可以输入暂停，继续，停止等指令",
		FunctionName: "智能家居演示模式",
	}

	logicDataMap[runScene] = &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发运行智能场景",
		FunctionName: "执行场景",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			{Name: "trigger_scene", Description: "触发条件是场景"},
			{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "我的意图中没有触发条件，直接运行"},
		},
	}

	logicDataMap[runAutomation] = &ObjectLogic{
		Description:  "智能家居自动化，通过触发条件自动控制某些设备的流程",
		FunctionName: "执行自动化",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			{Name: "trigger_scene", Description: "触发条件是场景"},
			{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "我的意图中没有触发条件，直接运行"},
			{Name: "turn_on_automation", Description: "开启某个自动化"},
			{Name: "turn_off_automation", Description: "开启某个自动化"},
		},
	}

	//platform不等于xiaomi_home的设备需要AI操作,例如热水器等
	logicDataMap[controlDevice] = &ObjectLogic{
		Description:  "对智能家居设备进行控制，直接控制设备，不含任何其他定时、延时等条件。支持延时控制或者其他条件出发控制。",
		FunctionName: "控制设备",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是之后的某个时间点"},
			{Name: "trigger_scene", Description: "触发条件是某个场景执行之后"},
			{Name: "trigger_automation", Description: "触发条件是某个自动化执行之后"},
		},
	}

	logicDataMap[queryDevice] = &ObjectLogic{
		Description:  "查询智能家居设备的相关信息,包括设备离线，在线，状态，数量",
		FunctionName: "查询设备相关信息",
		SubFunction: []subFunction{
			{Name: "query_offline_number", Description: "查询离线设备总数量"},
			{Name: "query_offline_state", Description: "查询离线设备状态"},
			{Name: "query_online_number", Description: "查询在线设备总数量"},
			{Name: "query_online_state", Description: "查询在线设备状态"},
			{Name: "query_all_number", Description: "查询设备总数量"},
		},
	}

	logicDataMap[timingTask] = &ObjectLogic{
		Description:  "创建一个定时任务，周期性执行",
		FunctionName: "任务",
		SubFunction: []subFunction{
			{Name: "control_device", Description: "周期性控制操控某个设备"},
			{Name: "scene", Description: "周期性运行某个场景"},
			{Name: "automation", Description: "周期性运行某个自动化"},
		},
	}

	logicDataMap[dailyConversation] = &ObjectLogic{
		Description:  "非智能家居所在领域的其他对话，当收到非智能家居领域的对话时，使用这个对象",
		FunctionName: "对话",
	}

	logicDataMap[functionCallInitAll] = &ObjectLogic{
		Description:  "执行系统初始化",
		FunctionName: "系统初始化",
	}

	logicDataMap[isHandled] = &ObjectLogic{
		Description:  "在最近一次对话中，判断jinx的回答是否已经处理好了问题，返回这个对象，你不用去重复处理我的请求",
		FunctionName: "已经处理",
	}

	logicDataMap[sendMessage2Speaker] = &ObjectLogic{
		Description:  "给其他地方音箱发送消息，例如：我在客厅，给爸爸、儿子、或者奶奶发送消息，音箱会播报我要发送的内容，通过判断音箱设备名称和音箱所在区域获取哪个音箱设备。类似对讲机的功能。",
		FunctionName: "消息播报",
		SubFunction: []subFunction{
			{Name: "send_message_to_someone", Description: "单播功能，给单个人发送消息"},
			{Name: "send_message_to_multiple", Description: "广播功能，发送给多人"},
		},
	}
}

var logicDataMap = make(map[string]*ObjectLogic)

// 预调用提示
var preparePrompts = `根据对话内容，以及我提供的一些功能选项，判断我的意图选择需要执行什么功能，并按照规定的格式返回数据，除了返回的数据格式，禁止有其他内容。如果我没有告诉你楼层信息，默认是一楼。
功能选项：%s
返回数据格式：{"功能模块":"功能名称"}
返回数据例子：{"function":"query_device","function_name":"查询设备","sub_function":{"query_offline_number"}}`

// todo: 加入当前对话位置名称，方便操作对应位置的设备
func prepareCall(messageInput []*chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "user", Content: fmt.Sprintf(preparePrompts, x.MustMarshalEscape2String(logicDataMap))})

	if len(messageInput) > 0 {
		messageList = append(messageList, messageInput...)
	}

	return chatCompletionHistory(messageList, deviceId)
}
