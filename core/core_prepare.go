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
// queryScene          = "query_scene"
// runScene            = "run_scene"
// runAutomation       = "run_automation"
// queryAutomation     = "query_automation"
// controlDevice       = "control_device"
// queryDevice         = "query_device"
// dailyConversation   = "daily_conversation"
// functionCallInitAll = "function_call_init_all"
// isHandled           = "is_handled"
// sendMessage2Speaker = "send_message_to_speaker"
// evaluate            = "evaluate"
// display             = "display"
)

var logicData = make([]*ObjectLogic, 0)

func init() {

	logicData = append(logicData, &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发查询智能场景信息",
		FunctionName: "查询智能家居场景",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询场景数量"},
			{Name: "query_detail", Description: "查询场景详情"},
		},
		f: QueryScene,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "查询智能家居自动化，通过触发条件自动控制某些设备的流程",
		FunctionName: "查询智能家居自动化",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询自动化数量"},
			{Name: "query_detail", Description: "查询自动化详情"},
		},
		f: QueryAutomation,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "对当前智能家居系统完善情况进行等级评估",
		FunctionName: "评估智能系统",
		f:            Evaluate,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "对我家的智能家居演示和讲解、介绍，用户可以通过输入'演示模式'来启动演示，支持输入暂停、继续、停止等指令。",
		FunctionName: "演示我家的智能家居系统",
		f:            Display,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发运行智能场景",
		FunctionName: "执行场景",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			{Name: "trigger_scene", Description: "触发条件是场景"},
			{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "我的意图中没有触发条件，直接运行"},
		},
		f: RunScene,
	})

	logicData = append(logicData, &ObjectLogic{
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
		f: RunAutomation,
	})

	//platform不等于xiaomi_home的设备需要AI操作,例如热水器等
	logicData = append(logicData, &ObjectLogic{
		Description:  "对智能家居设备进行控制，直接控制设备，不含任何其他定时、延时等条件。支持延时控制或者其他条件出发控制。",
		FunctionName: "控制设备",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			{Name: "trigger_scene", Description: "触发条件是场景"},
			{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "我的意图中没有触发条件，直接运行"},
		},
		f: RunDevice,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "查询智能家居设备的相关信息,包括设备离线，在线，状态，数量",
		FunctionName: "查询设备相关信息",
		SubFunction: []subFunction{
			{Name: "query_offline_number", Description: "查询离线设备总数量"},
			{Name: "query_offline_state", Description: "查询离线设备状态"},
			{Name: "query_online_number", Description: "查询在线设备总数量"},
			{Name: "query_online_state", Description: "查询在线设备状态"},
			{Name: "query_all_number", Description: "查询设备总数量"},
		},
		f: QueryDevice,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "学习、技术、情感、心理、生活等领域。",
		FunctionName: "咨询",
		f:            Conversation,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "执行系统初始化",
		FunctionName: "系统初始化",
		f:            InitALL,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "在最近一次对话中，判断jinx的回答是否已经处理好了问题，返回这个对象，你不用去重复处理我的请求",
		FunctionName: "已经处理",
		f:            IsDone,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "家庭对讲功能，允许用户在家中进行语音通话和传话。例如，用户可以通过此功能通知家人晚餐准备好了。带有‘喊’、‘叫’、‘通知’等字眼都属于对讲功能。",
		FunctionName: "对讲功能",
		SubFunction: []subFunction{
			{Name: "send_message_to_someone", Description: "向特定位置或人员发送消息，适用于点对点对讲。"},
			{Name: "send_message_to_all", Description: "向所有人员发送消息，适用于群体消息通知。"},
		},
		f: SendMessagePlay,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "判断某个区域是否有人",
		FunctionName: "区域是否有人",
		f:            isAnyoneHere,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "用于记录和管理个人记事内容的功能",
		FunctionName: "记事本添加和查询",
		f:            RunNote,
		SubFunction: []subFunction{
			{Name: "add_note", Description: "添加记事内容"},
			{Name: "query_note", Description: "查询记事内容"},
		},
	})
	logicData = append(logicData, &ObjectLogic{
		Description:  "sos紧急请求功能",
		FunctionName: "sos紧急求助",
		f:            RunSOS,
	})

	logicData = append(logicData, &ObjectLogic{
		Description:  "用于记录和管理家庭留言内容的功能",
		FunctionName: "家庭留言功能添加和查询",
		f:            RunNote,
		SubFunction: []subFunction{
			{Name: "add_message", Description: "添加留言内容"},
			{Name: "query_message", Description: "查询留言内容"},
		},
	})
}

// 预调用提示
var preparePrompts = `根据对话内容，以及我提供的一些功能选项，判断我的意图选择需要执行什么功能，并按照规定的格式返回数据，除了返回的数据格式，禁止有其他内容。如果我没有告诉你楼层信息，默认是一楼。
功能选项：%s
返回数据格式：{"function_name":"查询设备","sub_function":{"query_offline_number"}}`

// todo: 加入当前对话位置名称，方便操作对应位置的设备
func prepareCall(messageInput []*chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "user", Content: fmt.Sprintf(preparePrompts, x.MustMarshalEscape2String(logicData))})

	if len(messageInput) > 0 {
		messageList = append(messageList, messageInput...)
	}

	return chatCompletionHistory(messageList, deviceId)
}
