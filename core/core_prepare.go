package core

import (
	"errors"
	"fmt"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
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

var logicDataTwo = make([]*ObjectLogic, 0)
var logicDataOne = make([]*ObjectLogic, 0)
var logicDataALL = make([]*ObjectLogic, 0)
var logicData = make([]*ObjectLogic, 0)

func init() {

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发查询智能场景信息",
		FunctionName: "查询智能家居场景",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询场景数量"},
			{Name: "query_detail", Description: "查询场景详情"},
		},
		f: QueryScene,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "查询智能家居自动化，通过触发条件自动控制某些设备的流程",
		FunctionName: "查询智能家居自动化",
		SubFunction: []subFunction{
			{Name: "query_number", Description: "查询自动化数量"},
			{Name: "query_detail", Description: "查询自动化详情"},
		},
		f: QueryAutomation,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "对当前智能家居系统完善情况进行等级评估",
		FunctionName: "评估智能系统",
		f:            Evaluate,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "对我家的智能家居演示和讲解、介绍，用户可以通过输入'演示模式'来启动演示，支持输入暂停、继续、停止等指令。",
		FunctionName: "演示我家的智能家居系统",
		f:            Display,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "通过用户口令或者手动指令触发运行智能场景",
		FunctionName: "执行场景",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			//{Name: "trigger_scene", Description: "触发条件是场景"},
			//{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "直接运行场景，不需要触发条件"},
		},
		f: RunScene,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "智能家居自动化，通过触发条件自动控制某些设备的流程",
		FunctionName: "执行自动化",
		SubFunction: []subFunction{
			{Name: "trigger_time", Description: "触发条件是时间点"},
			//{Name: "trigger_scene", Description: "触发条件是场景"},
			//{Name: "trigger_automation", Description: "触发条件是自动化"},
			{Name: "run", Description: "我的意图中没有触发条件，直接运行"},
			{Name: "turn_on_automation", Description: "某个自动化由禁用变为启用"},
			{Name: "turn_off_automation", Description: "某个自动化由启用变为禁用"},
		},
		f: RunAutomation,
	})

	//platform不等于xiaomi_home的设备需要AI操作,例如热水器等
	logicDataTwo = append(logicDataTwo, &ObjectLogic{
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

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
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

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "执行系统初始化",
		FunctionName: "系统初始化",
		f:            InitALL,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "家庭对讲功能，允许用户在家中进行语音通话和传话。例如，用户可以通过此功能通知家人晚餐准备好了。带有‘喊’、‘叫’、‘通知’等字眼都属于对讲功能。",
		FunctionName: "对讲功能",
		SubFunction: []subFunction{
			{Name: "send_message_to_someone", Description: "向特定位置或人员发送消息，适用于点对点对讲。"},
			{Name: "send_message_to_all", Description: "向所有人员发送消息，适用于群体消息通知。"},
		},
		f: SendMessagePlay,
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "判断某个区域是否有人",
		FunctionName: "区域是否有人",
		f:            isAnyoneHere,
	})

	//logicDataTwo = append(logicDataTwo, &ObjectLogic{
	//	Description:  "用于记录和管理个人日程和重要事情备忘内容的功能",
	//	FunctionName: "日程安排",
	//	f:            RunNote,
	//	SubFunction: []subFunction{
	//		{Name: "add_note", Description: "添加记事内容"},
	//		{Name: "query_note", Description: "查询记事内容"},
	//	},
	//})
	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "sos紧急请求功能",
		FunctionName: "sos紧急求助",
		f:            RunSOS,
	})

	//logicDataTwo = append(logicDataTwo, &ObjectLogic{
	//	Description:  "用于记录和管理家庭留言内容的功能",
	//	FunctionName: "家庭留言功能添加和查询",
	//	f:            RunMessage,
	//	SubFunction: []subFunction{
	//		{Name: "add_message", Description: "添加留言内容"},
	//		{Name: "query_message", Description: "查询留言内容"},
	//	},
	//})

	logicDataOne = append(logicDataOne, &ObjectLogic{
		Description:  "精准选择目标函数",
		FunctionName: "功能选择",
	})

	logicDataTwo = append(logicDataTwo, &ObjectLogic{
		Description:  "学习、技术、情感、心理、生活等领域。",
		FunctionName: "咨询",
		f:            Conversation,
	})

	var ojb = &ObjectLogic{
		Description:  "我家里的智能家居功能。包括：",
		FunctionName: "智能家居",
	}

	for _, v := range logicDataTwo {
		ojb.Description += v.FunctionName + "、"
	}

	logicData = append(logicData, ojb)

	logicData = append(logicData, &ObjectLogic{
		Description:  "专业知识领域",
		FunctionName: "专业知识咨询",
	})
}

var preparePrompts = `你是我的私人助理，根据前面的对话内容，选择精准的功能函数返回给我，除了返回的数据格式，禁止有其他内容。
功能选项：%s
返回数据：
返回JSON格式： {"function_name":""}`

// 预调用提示
var preparePromptsOne = `你是我的私人助理，根据前面的对话内容，选择精准的功能函数返回给我，除了返回的数据格式，禁止有其他内容。
功能选项：%s
返回数据：
返回JSON格式： {"function_name":"","sub_function":""}`

var preparePromptsTwo = `你是一个智能家居管家，根据前面的对话内容，选择精准的功能函数返回给我。除了返回的数据格式，禁止有其他内容。
功能选项：%s
返回数据格式：{"function_name":"","sub_function":"query_offline_number"}`

func prepareCall(messageInput *chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "system", Content: fmt.Sprintf(preparePrompts, x.MustMarshal2String(logicData))})

	messageList = append(messageList, messageInput)

	result, err := chatCompletionInternal(messageList)
	if err != nil {
		ava.Error(err)
		return "智能体选择失败", err
	}

	if strings.Contains(result, "智能家居") {
		return prepareCallTwo(messageInput, deviceId)
	}

	if strings.Contains(result, "专业知识") {
		return prepareCallOne(messageInput, deviceId)
	}

	return "没有找到智能体", errors.New("no agent")
}

func prepareCallOne(messageInput *chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "system", Content: fmt.Sprintf(preparePromptsOne, x.MustMarshal2String(logicDataOne))})

	messageList = append(messageList, messageInput)

	return chatCompletionHistory(messageList, deviceId)
}
func prepareCallTwo(messageInput *chat.ChatMessage, deviceId string) (string, error) {
	var messageList = make([]*chat.ChatMessage, 0, 6)
	messageList = append(messageList, &chat.ChatMessage{Role: "system", Content: fmt.Sprintf(preparePromptsTwo, x.MustMarshal2String(logicDataTwo))})

	messageList = append(messageList, messageInput)

	return chatCompletionHistory(messageList, deviceId)
}
