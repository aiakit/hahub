package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

var gFunctionRouter *FunctionRouter

func CoreChaos() {
	gFunctionRouter = NewFunctionRouter()

	for _, v := range logicDataTwo {
		logicDataALL = append(logicDataALL, v)
		gFunctionRouter.Register(v.FunctionName, v.f)
	}
	for _, v := range logicDataOne {
		logicDataALL = append(logicDataALL, v)
		gFunctionRouter.Register(v.FunctionName, v.f)
	}

	gFunctionRouter.Register("开发中", IsInDevelopment)

	data.RegisterDataHandler(registerHomingWelcome)
	//data.RegisterDataHandler(registerToggleLight)

	chaosSpeaker()
}

// 定义函数处理类型
type FunctionHandler func(message, aiMessage, deviceId string) string

// 函数映射表结构
type FunctionRouter struct {
	handlers map[string]FunctionHandler
}

// 创建新的函数路由器
func NewFunctionRouter() *FunctionRouter {
	return &FunctionRouter{
		handlers: make(map[string]FunctionHandler),
	}
}

// 注册函数处理器
func (fr *FunctionRouter) Register(functionName string, handler FunctionHandler) {
	fr.handlers[functionName] = handler
}

// 根据函数名调用对应的函数
func Call(functionName, deviceId, message, aiMessage string) string {
	if handler, exists := gFunctionRouter.handlers[functionName]; exists {
		var now = time.Now()
		result := handler(message, aiMessage, deviceId)
		ava.Debugf("latency=%.2f |funcion_name=%s |message=%s |ai_message=%s |FROM=%s", time.Since(now).Seconds(), functionName, message, aiMessage, result)
		return result
	}
	// 如果没有找到对应的处理器，返回空字符串或错误信息
	return "未知指令"
}

func findFunction(message string) string {
	for _, v := range logicDataALL {
		if strings.Contains(message, v.FunctionName) {
			return v.FunctionName
		}
	}

	return "开发中"
}

func chatCompletionInternal(msgInput []*chat.ChatMessage) (string, error) {
	var message = make([]*chat.ChatMessage, 0, 5)

	message = append(message, msgInput...)

	return chat.ChatCompletionMessage(message)
}

func chatCompletionHistory(msgInput []*chat.ChatMessage, deviceId string) (string, error) {
	history := GetHistory(deviceId)
	var message = make([]*chat.ChatMessage, 0, 5)

	if len(history) > 0 {
		message = append(message, &chat.ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf(`历史对话记录: %s`, x.MustMarshal2String(history)),
		})
	}

	message = append(message, msgInput...)

	result, err := chat.ChatCompletionMessage(message)
	if err != nil {
		return "发生未知错误，请重试", err
	}

	for _, v := range msgInput {
		if v.Role == "user" {
			AddUserMessage(deviceId, v.Content)
		}
	}

	AddAIMessage(deviceId, result)

	return result, nil
}

// 等待3秒运行
// 当前温度湿度
// 当前家里有人，空调，热水器是否打开，生成欢迎词
// 播放回家场景执行逻辑
// 空调，电视，热水器或者其他设备都可以告诉我
// 打开设备
func registerHomingWelcome(simple *data.StateChangedSimple, body []byte) {

	if !strings.Contains(simple.Event.Data.NewState.State, "on") {
		return
	}

	if strings.HasPrefix(simple.Event.Data.NewState.EntityID, "automation.") &&
		strings.Contains(simple.Event.Data.NewState.Attributes.FriendlyName, "回家自动化") {
		fmt.Println("------1--", string(body))
		if simple.Event.Data.NewState.Attributes.Current == 0 {
			return
		}
		time.Sleep(time.Second * 3)

		var msg string
		//湿度
		var humidity string
		func() {
			entities := data.GetEntityCategoryMap()[data.CategoryHumiditySensor]
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					s, err := data.GetState(v.EntityID)
					if err != nil {
						continue
					}
					humidity = s.State
					break
				}
			}
		}()

		if humidity != "" {
			msg += fmt.Sprintf("当前湿度是%s, ", humidity)
		}

		//温度
		var temperature string

		func() {
			entities := data.GetEntityCategoryMap()[data.CategoryTemperatureSensor]
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					s, err := data.GetState(v.EntityID)
					if err != nil {
						continue
					}
					temperature = s.State
					break
				}
			}
		}()

		if temperature != "" {
			msg += fmt.Sprintf("，当前温度是%s, ", temperature)
		}

		//热水器状态
		var waterHeaterState string
		func() {
			entities := data.GetEntityCategoryMap()[data.CategoryWaterHeater]
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					s, err := data.GetState(v.EntityID)
					if err != nil {
						continue
					}
					if strings.ToLower(s.State) == "on" {
						waterHeaterState = "打开"
					}

					if strings.ToLower(s.State) == "off" {
						waterHeaterState = "关闭"
					}
					if waterHeaterState != "" {
						break
					}
				}
			}
		}()

		if waterHeaterState != "" {
			msg += fmt.Sprintf("，热水器状态是%s, ", waterHeaterState)
		}

		var airConditionerState string
		func() {
			entities := data.GetEntityCategoryMap()[data.CategoryAirConditioner]
			for _, v := range entities {
				if strings.Contains(v.AreaName, "客厅") {
					s, err := data.GetState(v.EntityID)
					if err != nil {
						continue
					}
					if strings.ToLower(s.State) == "on" {
						airConditionerState = "打开"
						airConditionerState += fmt.Sprintf("，空调工作状态是%v, ", s.Attributes)
					}

					if strings.ToLower(s.State) == "off" {
						airConditionerState = "关闭"
					}
					if airConditionerState != "" {
						break
					}
					break
				}
			}
		}()

		if airConditionerState != "" {
			msg += fmt.Sprintf("，空调是%s, ", airConditionerState)
		}

		//播放欢迎词
		var id string

		func() {
			result, err := chat.ChatCompletionMessage([]*chat.ChatMessage{{
				Role:    "user",
				Content: "你是一个智能家居系统，我是你的宿主，我现在回家了，用30个字的左右的人性化、俏皮结合古诗词的语言欢迎我，不要有任何表情或者图案。",
			}})

			if err != nil {
				ava.Error(err)
				return
			}

			entities, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
			if !ok {
				return
			}

			for _, e := range entities {
				if strings.HasPrefix(e.EntityID, "text.") && strings.Contains(e.EntityID, "play_text_") && strings.Contains(e.AreaName, "客厅") {
					id = e.EntityID
					break
				}
			}

			// 添加检查，防止id为空
			if id == "" {
				ava.Warn("No suitable media player found for welcome message")
				return
			}

			err = x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
				EntityId: id,
				Value:    result,
			}, nil)
			if err != nil {
				ava.Error(err)
			}
			time.Sleep(GetPlaybackDuration(result) + 1)
		}()

		func() {
			//是否打开设备
			if msg == "" {
				return
			}

			result, err := chat.ChatCompletionMessage([]*chat.ChatMessage{{
				Role:    "user",
				Content: fmt.Sprintf("你是一个智能家居系统，我是你的宿主，我现在回家了，根据环境和状态信息(%s)用人性化的语言给出建议。例如：当前温度xxx，是否需要打开空调。", msg),
			}})

			if err != nil {
				ava.Error(err)
				return
			}
			// 添加检查，防止id为空
			if id == "" {
				ava.Warn("No suitable media player found for welcome message")
				return
			}

			err = x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
				EntityId: id,
				Value:    result,
			}, nil)
			if err != nil {
				ava.Error(err)
			}

			time.Sleep(GetPlaybackDuration(result) + 1)
		}()

		func() {
			//播放回家场景详情
			var auto interface{}
			err := intelligent.GetAutomation("automation.hui_jia_zi_dong_hua", &auto)
			if err == nil {
				result, err := chat.ChatCompletionMessage([]*chat.ChatMessage{{
					Role:    "user",
					Content: fmt.Sprintf("你是一个智能家居系统，用人性化的语言描述智能系统回家自动化信息%v,这个系统是属于你创建的，所以尽量用‘我’作为第一人称去描述，描述结束之后记得有一个温馨的结尾。", auto),
				}})

				if err != nil {
					ava.Error(err)
					return
				}
				// 添加检查，防止id为空
				if id == "" {
					ava.Warn("No suitable media player found for welcome message")
					return
				}

				err = x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
					EntityId: id,
					Value:    result,
				}, nil)
				if err != nil {
					ava.Error(err)
				}

				time.Sleep(GetPlaybackDuration(result) + 1)
			}
		}()
	}
}

// 语音指令关灯之后，就不要再自动开灯,直到语音指令开灯
func registerToggleLight(simple *data.StateChangedSimple, body []byte) {
	var state chatMessage
	err := x.Unmarshal(body, &state)
	if err != nil {
		ava.Error(err)
		return
	}

	if strings.Contains(state.Event.Data.EntityID, "_conversation") &&
		strings.EqualFold(state.Event.Data.NewState.Attributes.EntityClass, "XiaoaiConversationSensor") {
		//找到所有根据存在传感器有人亮灯的自动化
		if (strings.Contains(simple.Event.Data.NewState.State, "关") && strings.Contains(simple.Event.Data.NewState.State, "灯")) ||
			strings.Contains(simple.Event.Data.NewState.State, "晚安") || strings.Contains(simple.Event.Data.NewState.State, "睡觉") || strings.Contains(simple.Event.Data.NewState.State, "午觉") {

			entity, ok := data.GetEntityByEntityId()[simple.Event.Data.EntityID]
			if !ok {
				return
			}
			areaName := data.SpiltAreaName(entity.AreaName)

			//查询所有自动化
			as, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
			if !ok {
				return
			}

			for _, a := range as {
				if strings.Contains(a.OriginalName, areaName) && strings.Contains(a.OriginalName, "有人亮灯") {
					//关闭自动化
					err = intelligent.TurnOffAutomation(ava.Background(), a.EntityID)
					if err != nil {
						ava.Error(err)
						return
					}
				}
			}
		}

		if (strings.Contains(simple.Event.Data.NewState.State, "开") && strings.Contains(simple.Event.Data.NewState.State, "灯")) ||
			strings.Contains(simple.Event.Data.NewState.State, "起床") || strings.Contains(simple.Event.Data.NewState.State, "早安") {
			{
				entity, ok := data.GetEntityByEntityId()[simple.Event.Data.EntityID]
				if !ok {
					return
				}
				areaName := data.SpiltAreaName(entity.AreaName)

				//查询所有自动化
				as, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
				if !ok {
					return
				}

				for _, a := range as {
					if strings.Contains(a.OriginalName, areaName) && strings.Contains(a.OriginalName, "有人亮灯") {
						err = intelligent.TurnOnAutomation(ava.Background(), a.EntityID)
						if err != nil {
							ava.Error(err)
							return
						}
					}
				}
			}
		}
	}
}
