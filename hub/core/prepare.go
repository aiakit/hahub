package core

import "hahub/proto/phome"

// 功能预处理,AI对json的组装效果不好，用文字返回代替处理,无论AI返回什么样的格式，用字符串去进行一起匹配
// 板块名称：动作：对象
// eg.场景，执行场景，回家
// eg.天气，查询天气，小爱入口忽略/云端查询
func init() {
	logicDataMap = &phome.PrepareData{}
	logicDataMap.Objects = make(map[string]*phome.ObjectLogic)

	logicDataMap.Objects["scene"] = &phome.ObjectLogic{
		Description:  "通过用户口令或者手动指令触发的对智能家居设备一系列的操作",
		Action:       []string{"执行场景", "修改场景", "查询场景"},
		FunctionName: "场景",
	}

	logicDataMap.Objects["automation"] = &phome.ObjectLogic{
		Description:  "通过一些自然条件（例如：天气、温度、湿度、时间等）或者设备条件（例如：水浸传感器、人体传感器等设备状态）或者某个事件触发（例如：当执行睡觉场景之后就关闭人来亮灯自动化）对智能家居设备一系列的操作",
		Action:       []string{"执行自动化", "修改自动化", "查询自动化"},
		FunctionName: "自动化",
	}

	logicDataMap.Objects["device"] = &phome.ObjectLogic{
		Description:  "对智能家居设备进行信息获取或者操作设备",
		Action:       []string{"操作设备", "查询设备状态", "查询设备数量"},
		FunctionName: "设备",
	}

	logicDataMap.Objects["task"] = &phome.ObjectLogic{
		Description:  "创建并执行定时、延时任务或者顺序任务",
		Action:       []string{"定时器", "顺序执行", "延时"},
		FunctionName: "任务",
	}

	logicDataMap.Objects["daily_conversation"] = &phome.ObjectLogic{
		Description:  "非智能家居所在领域的对话",
		Action:       []string{"其他"},
		FunctionName: "日常对话",
	}

	logicDataMap.Objects["weather"] = &phome.ObjectLogic{
		Description:  "查询当地或者某个地区的天气",
		Action:       []string{"查询天气"},
		FunctionName: "天气",
	}

	logicDataMap.Objects["date"] = &phome.ObjectLogic{
		Description:  "对日期时间的查",
		Action:       []string{"查询天气"},
		FunctionName: "时间日期",
	}

	logicDataMap.Objects["function_call"] = &phome.ObjectLogic{
		Description:  "执行一些函数调用",
		Action:       []string{"初始化场景", "初始化自动化", "初始化灯具", "初始化开关", "初始化灯光效果", "初始化灯控自动化"},
		FunctionName: "函数调用",
	}
}

var logicDataMap *phome.PrepareData

// 预调用提示
var preparePrompts = `我提供了一些功能选项，根据我的意图选择需要执行什么功能，并按照规定的格式返回数据，除了返回的数据格式，禁止有其他内容。
功能选项：%s
返回数据格式：{"功能模块","动作"}
返回数据例子：{"函数调用","初始化场景"}`
