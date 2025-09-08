package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"

	"github.com/aiakit/ava"
)

func Evaluate(message, aiMessage, deviceId string) string {
	//智能家居评估C,B,A,S,SS,SSS，热水器，空调，灯光，扫地机，地暖，插座，电视，开关，浴霸，人体传感器，水浸传感器，烟雾报警器，燃气报警器，门锁，门，窗帘，床，洗衣机，冰箱，洗碗机，温度，湿度,等等进行综合评估
	//智能家居建议
	//设备数量，开启情况
	//场景数量介绍,大致有哪些场景
	//自动化情况介绍，大致有哪些自动化
	device := data.GetDevices()
	var d = make([]*shortDevice, 0)
	for _, e := range device {
		d = append(d, &shortDevice{
			Name: e.Name,
			id:   e.ID,
		})
	}
	scene := data.GetEntityCategoryMap()[data.CategoryScript]
	var s = make([]*shortScene, 0)
	for _, e := range scene {
		s = append(s, &shortScene{
			Alias: e.OriginalName,
			id:    e.EntityID,
		})
	}

	auto := data.GetEntityCategoryMap()[data.CategoryAutomation]
	var a = make([]*shortScene, 0)
	for _, e := range auto {
		a = append(a, &shortScene{
			Alias: e.OriginalName,
			id:    e.EntityID,
		})
	}

	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role: "system",
			Content: fmt.Sprintf(`你是一个智能家居专家，现在你需要根据当前智能家居情况进行人性化的描述，300字左右，需要突出重点，然后按照C,B,A,S,SS,SSS等级对当前智能家居系统进行评估。
当前设备信息:%s。
当前场景信息:%s。
当前是否使用AI助手：是。
当前自动化信息：%s`, x.MustMarshal2String(d), x.MustMarshal2String(s), x.MustMarshal2String(a)),
		},
	})
	if err != nil {
		ava.Error(err)
		return "服务器出错了"
	}

	return result
}

// 展示，支持暂停，继续，停止，循环
// 0.介绍灯光
// 1.简单播报情况，设备类型，设备数量，场景，自动化
// 2.逐个介绍场景和自动化
func Display(message, aiMessage, deviceId string) string {
	executeSteps(deviceId)
	return "好的宿主，正在加载数据，即将启动演示模式。"
}

type simple struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	State    string `json:"state,omitempty"`
	AreaName string `json:"area_name,omitempty"`
}

// executeSteps 从指定步骤开始执行展示逻辑
// step: 起始步骤 (0=介绍灯光, 1=播报总体情况, 2=逐个介绍场景和自动化)
// sceneIndex: 场景介绍起始索引
// automationIndex: 自动化介绍起始索引
var descriptionIsRunning bool
var isRunningLock sync.RWMutex

func executeSteps(deviceId string) {
	aiLock(deviceId)
	defer aiUnlock(deviceId)
	isRunningLock.Lock()
	descriptionIsRunning = true
	isRunningLock.Unlock()

	// 第一步：介绍灯光
	lightDevices := data.GetEntityCategoryMap()[data.CategoryLight]
	states, err := data.GetStates()
	if err != nil {
		ava.Error(err)
		return
	}

	var simples = make([]*simple, 0, 10)
	var simpleDevice = make(map[string]*simple, 0)
	for _, v := range lightDevices {
		simpleDevice[v.EntityID] = &simple{
			Name:     v.DeviceName,
			Type:     v.Category,
			AreaName: data.SpiltAreaName(v.AreaName),
		}
	}

	for _, v := range states {
		if _, ok := simpleDevice[v.EntityID]; ok {
			simpleDevice[v.EntityID].State = v.State
			simples = append(simples, simpleDevice[v.EntityID])
		}
	}

	if len(lightDevices) > 0 {
		isRunningLock.RLock()
		running := descriptionIsRunning
		isRunningLock.RUnlock()
		if !running {
			return
		}
		lightResult, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要介绍当前的灯光设备情况。
请用简洁的自然语言描述家里的灯光情况，你的描述中，标点或符号只能有逗号和句号。例如：总共有多少个区域，多少灯亮着。
灯光设备信息：%s`, x.MustMarshal2String(simples)),
			},
		})
		if err != nil {
			ava.Error(err)
			return
		} else {
			isRunningLock.RLock()
			running := descriptionIsRunning
			isRunningLock.RUnlock()
			if !running {
				return
			}
			// 这里应该调用语音播报接口
			PlayTextAction(deviceId, "宿主，你好，我将为你介绍家里的智能灯情况。")
			PlayTextAction(deviceId, lightResult)
		}
	}
	// 第二步：简单播报情况，设备类型，设备数量，场景，自动化
	// 在播报总体情况时也需要检查控制命令
	devices := data.GetDevices()

	var simpleDeviceTwo = make([]*simple, 0)
	var deviceMap = make(map[string]bool)
	for _, v := range devices {
		if strings.Contains(v.Model, "wifispeaker") && deviceMap[v.Name] {
			continue
		}

		if strings.Contains(v.Model, "wifispeaker") && !deviceMap[v.Name] {
			deviceMap[v.Name] = true
		}

		simpleDeviceTwo = append(simpleDeviceTwo, &simple{
			Name: v.Name,
			Type: v.Model,
		})
	}

	scenes := data.GetEntityCategoryMap()[data.CategoryScript]

	var simpleScript = make([]*simple, 0)
	for _, v := range scenes {
		simpleScript = append(simpleScript, &simple{
			Name: v.OriginalName,
			Type: "场景",
		})
	}

	automations := data.GetEntityCategoryMap()[data.CategoryAutomation]

	var simpleAutomation = make([]*simple, 0)
	for _, v := range automations {
		simpleAutomation = append(simpleAutomation, &simple{
			Name: v.OriginalName,
			Type: "自动化",
		})
	}

	isRunningLock.RLock()
	running := descriptionIsRunning
	isRunningLock.RUnlock()
	if !running {
		return
	}
	summaryResult, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role: "system",
			Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要简单播报当前智能家居的整体情况。你的描述中，标点或符号只能有逗号和句号。
请按以下格式进行描述：
- 设备类型和数量统计
- 场景数量
- 自动化数量
设备信息：%s
场景信息：%s
自动化信息：%s`, x.MustMarshal2String(simpleDevice), x.MustMarshal2String(simpleScript), x.MustMarshal2String(simpleAutomation)),
		},
	})
	if err != nil {
		ava.Error(err)
		return
	} else {
		isRunningLock.RLock()
		running := descriptionIsRunning
		isRunningLock.RUnlock()
		if !running {
			return
		}

		PlayTextAction(deviceId, "接下来为你介绍家里的设备种类等情况。")
		// 这里应该调用语音播报接口
		PlayTextAction(deviceId, summaryResult)
	}

	// 第三步：逐个介绍场景和自动化
	scenesThree := data.GetEntityCategoryMap()[data.CategoryScript]
	automationsThree := data.GetEntityCategoryMap()[data.CategoryAutomation]

	// 介绍场景，从指定索引开始
	for _, scene := range scenesThree {
		isRunningLock.RLock()
		running := descriptionIsRunning
		isRunningLock.RUnlock()
		if !running {
			return
		}
		sceneResult, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要详细介绍一个场景。
请用自然语言描述这个场景的名称和功能，你的描述中，标点或符号只能有逗号和句号。
场景信息：%s`, x.MustMarshal2String(scene)),
			},
		})
		if err != nil {
			ava.Error(err)
			return
		} else {
			isRunningLock.RLock()
			running := descriptionIsRunning
			isRunningLock.RUnlock()
			if !running {
				return
			}
			// 这里应该调用语音播报接口
			PlayTextAction(deviceId, sceneResult)
		}
	}

	// 介绍自动化，从指定索引开始
	for _, automation := range automationsThree {
		isRunningLock.RLock()
		running := descriptionIsRunning
		isRunningLock.RUnlock()
		if !running {
			return
		}
		autoResult, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要详细介绍一个自动化。
请用自然语言描述这个自动化的名称和触发条件及执行动作。你的描述中，标点或符号只能有逗号和句号。
自动化信息：%s`, x.MustMarshal2String(automation)),
			},
		})
		if err != nil {
			ava.Error(err)
			return
		} else {
			isRunningLock.RLock()
			running := descriptionIsRunning
			isRunningLock.RUnlock()
			if !running {
				return
			}
			// 这里应该调用语音播报接口
			PlayTextAction(deviceId, autoResult)
		}
	}
}
