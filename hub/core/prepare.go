package core

// 功能预处理
// 板块名称：动作：对象
// eg.场景，执行场景，回家
// eg.天气，查询天气，小爱入口忽略/云端查询
func init() {
	logicDataMap = make(map[string]*ObjectLogic)

	logicDataMap["场景"] = &ObjectLogic{
		Description: "通过用户口令或者手动指令触发的对智能家居设备一系列的操作",
		Action:      []string{"执行场景", "修改场景", "查询场景"},
	}

	logicDataMap["自动化"] = &ObjectLogic{
		Description: "通过一些自然条件（例如：天气、温度、湿度、时间等）或者设备条件（例如：水浸传感器、人体传感器等设备状态）或者某个事件触发（例如：当执行睡觉场景之后就关闭人来亮灯自动化）对智能家居设备一系列的操作",
		Action:      []string{"执行自动化", "修改自动化", "查询自动化"},
	}

	logicDataMap["设备"] = &ObjectLogic{
		Description: "对智能家居设备进行信息获取或者操作设备",
		Action:      []string{"操作设备", "修改设备信息", "查询设备状态", "查询设备数量"},
	}

	logicDataMap["任务"] = &ObjectLogic{
		Description: "创建并执行定时、延时任务或者顺序任务",
		Action:      []string{"定时器", "顺序执行", "延时"},
	}

	logicDataMap["日常对话"] = &ObjectLogic{
		Description: "并非针对智能家居所在领域的对话,每种类型相当于一个智能agent,默认是其他类型",
		Action:      []string{"其他"},
	}

	logicDataMap["天气"] = &ObjectLogic{
		Description: "查询当地或者某个地区的天气",
		Action:      []string{"查询天气"},
	}

	logicDataMap["时间日期"] = &ObjectLogic{
		Description: "对日期时间的查",
		Action:      []string{"查询天气"},
	}

	logicDataMap["函数调用"] = &ObjectLogic{
		Description: "执行一些函数调用",
		Action:      []string{"初始化"},
	}
}

var logicDataMap = make(map[string]*ObjectLogic)

// 预调用提示
var preparePrompts = ``
