package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aiakit/ava"
)

func Evaluate(message, aiMessage, deviceId string) string {
	//智能家居评估C,B,A,S,SS,SSS，热水器，空调，灯光，扫地机，地暖，插座，电视，开关，浴霸，人体传感器，水浸传感器，烟雾报警器，燃气报警器，门锁，门，窗帘，床，洗衣机，冰箱，洗碗机，温度，湿度,等等进行综合评估
	//智能家居建议
	//设备数量，开启情况
	//场景数量介绍,大致有哪些场景
	//自动化情况介绍，大致有哪些自动化
	device := data.GetDevice()
	scene := data.GetEntityCategoryMap()[data.CategoryScript]
	auto := data.GetEntityCategoryMap()[data.CategoryAutomation]
	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role: "system",
			Content: fmt.Sprintf(`你是一个智能家居专家，现在你需要根据当前智能家居情况进行人性化的描述，需要突出重点，然后对智能家居等级进行评估，并给出建议，按照C,B,A,S,SS,SSS等级进行评估。
当前设备信息:%s
当前场景信息:%s
当前自动化信息：%s`, x.MustMarshalEscape2String(device), x.MustMarshalEscape2String(scene), x.MustMarshalEscape2String(auto)),
		},
	})
	if err != nil {
		ava.Error(err)
		return "服务器出错了"
	}

	return result
}

// 控制命令类型
const (
	ControlPause = "暂停"
	ControlStop  = "停止"
)

// 添加全局变量存储起始值，使用map存储每个设备的状态
var (
	deviceStates = make(map[string]*DeviceState)
	stateMutex   = sync.RWMutex{}
)

// DeviceState 设备状态结构体
type DeviceState struct {
	StartStep            int
	SceneStartIndex      int
	AutomationStartIndex int
	ControlChan          chan string //用于控制展示流程的通道
	PlayingStats         int32
}

// 获取设备状态，如果不存在则创建新的
func getOrCreateDeviceState(deviceId string) *DeviceState {
	stateMutex.RLock()
	state, exists := deviceStates[deviceId]
	stateMutex.RUnlock()

	if !exists {
		stateMutex.Lock()
		// 双重检查
		if state, exists = deviceStates[deviceId]; !exists {
			state = &DeviceState{
				ControlChan: make(chan string, 1), // 创建带缓冲的通道，避免阻塞
			}
			deviceStates[deviceId] = state
		}
		stateMutex.Unlock()
		return state
	}

	return state
}

// 重置设备状态
func resetDeviceState(deviceId string) {
	stateMutex.Lock()
	state, exists := deviceStates[deviceId]
	if exists {
		// 关闭通道
		close(state.ControlChan)
		delete(deviceStates, deviceId)
	}
	stateMutex.Unlock()
}

func loadPlayingStats(deviceId string) int32 {
	stateMutex.RLock()
	state, exists := deviceStates[deviceId]
	stateMutex.RUnlock()
	if exists {
		return state.PlayingStats
	}
	return 0
}

// 展示，支持暂停，继续，停止，循环
// 0.介绍灯光
// 1.简单播报情况，设备类型，设备数量，场景，自动化
// 2.逐个介绍场景和自动化
func Display(message, aiMessage, deviceId string) string {
	deviceState := getOrCreateDeviceState(deviceId)

	// 发送控制命令到通道 - 修改为包含匹配而不是精确匹配
	if strings.Contains(message, "暂停") || strings.Contains(message, "等一下") || strings.Contains(message, "等等") {
		deviceState.ControlChan <- ControlPause
		atomic.SwapInt32(&deviceState.PlayingStats, 0)
		return "好的，正在暂停演示。"
	}

	if strings.Contains(message, "退出") || strings.Contains(message, "可以了") || strings.Contains(message, "退下") {
		deviceState.ControlChan <- ControlStop
		resetDeviceState(deviceId)
		atomic.SwapInt32(&deviceState.PlayingStats, 0)
		return "好的，已退出演示模式。"
	}

	if atomic.LoadInt32(&deviceState.PlayingStats) == 1 {
		deviceState.ControlChan <- ControlStop
	}

	atomic.SwapInt32(&deviceState.PlayingStats, 1)

	go play(deviceState.StartStep, deviceState.SceneStartIndex, deviceState.AutomationStartIndex, deviceId)
	return "正在载入展示数据，展示数据加载中，即将启动演示介绍模式。"
}

func play(startStep, sceneStartIndex, automationStartIndex int, deviceId string) {
	// 按步骤执行展示逻辑
	executeSteps(startStep, sceneStartIndex, automationStartIndex, deviceId)
}

// executeSteps 从指定步骤开始执行展示逻辑
// step: 起始步骤 (0=介绍灯光, 1=播报总体情况, 2=逐个介绍场景和自动化)
// sceneIndex: 场景介绍起始索引
// automationIndex: 自动化介绍起始索引
func executeSteps(startStep, sceneStartIndex, automationStartIndex int, deviceId string) {
	deviceState := getOrCreateDeviceState(deviceId)

	step := startStep
	sceneIndex := sceneStartIndex
	automationIndex := automationStartIndex

	for step < 3 {
		switch step {
		case 0:
			// 第一步：介绍灯光
			lightDevices := data.GetEntityCategoryMap()[data.CategoryLight]
			if len(lightDevices) > 0 {
				// 在介绍灯光时也需要检查控制命令
				select {
				case command := <-deviceState.ControlChan:
					switch command {
					case ControlPause:
						// 记录当前状态并退出
						deviceState.StartStep = step
						return
					case ControlStop:
						return
					}
				default:
				}

				lightResult, err := chatCompletionInternal([]*chat.ChatMessage{
					{
						Role: "system",
						Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要介绍当前的灯光设备情况。
请用自然语言描述当前灯光设备的总数和状态，例如"当前共有X个灯光设备，其中Y个开启，Z个关闭"。
灯光设备信息：%s`, x.MustMarshalEscape2String(lightDevices)),
					},
				})
				if err != nil {
					ava.Error(err)
				} else {
					// 在介绍灯光时也需要检查控制命令
					select {
					case command := <-deviceState.ControlChan:
						switch command {
						case ControlPause:
							// 记录当前状态并退出
							deviceState.StartStep = step
							return
						case ControlStop:
							return
						}
					default:
					}

					// 这里应该调用语音播报接口
					PlayTextAction(deviceId, lightResult)
				}
			}
			step++
		case 1:
			// 第二步：简单播报情况，设备类型，设备数量，场景，自动化
			// 在播报总体情况时也需要检查控制命令
			select {
			case command := <-deviceState.ControlChan:
				switch command {
				case ControlPause:
					// 记录当前状态并退出
					deviceState.StartStep = step
					return
				case ControlStop:
					return
				}
			default:
			}

			devices := data.GetDevice()
			scenes := data.GetEntityCategoryMap()[data.CategoryScript]
			automations := data.GetEntityCategoryMap()[data.CategoryAutomation]

			summaryResult, err := chatCompletionInternal([]*chat.ChatMessage{
				{
					Role: "system",
					Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要简单播报当前智能家居的整体情况。
请按以下格式进行描述：
- 设备类型和数量统计
- 场景数量
- 自动化数量
设备信息：%s
场景信息：%s
自动化信息：%s`, x.MustMarshalEscape2String(devices), x.MustMarshalEscape2String(scenes), x.MustMarshalEscape2String(automations)),
				},
			})
			if err != nil {
				ava.Error(err)
			} else {
				select {
				case command := <-deviceState.ControlChan:
					switch command {
					case ControlPause:
						// 记录当前状态并退出
						deviceState.StartStep = step
						return
					case ControlStop:
						return
					}
				default:
				}

				// 这里应该调用语音播报接口
				PlayTextAction(deviceId, summaryResult)
			}
			step++
		case 2:
			// 第三步：逐个介绍场景和自动化
			scenes := data.GetEntityCategoryMap()[data.CategoryScript]
			automations := data.GetEntityCategoryMap()[data.CategoryAutomation]

			// 介绍场景，从指定索引开始
			for i := sceneIndex; i < len(scenes); i++ {
				select {
				case command := <-deviceState.ControlChan:
					switch command {
					case ControlPause:
						// 记录当前状态并退出
						deviceState.StartStep = step
						deviceState.SceneStartIndex = i
						return
					case ControlStop:
						return
					}
				default:
				}

				scene := scenes[i]
				sceneResult, err := chatCompletionInternal([]*chat.ChatMessage{
					{
						Role: "system",
						Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要详细介绍一个场景。
请用自然语言描述这个场景的名称和功能。
场景信息：%s`, x.MustMarshalEscape2String(scene)),
					},
				})
				if err != nil {
					ava.Error(err)
				} else {
					select {
					case command := <-deviceState.ControlChan:
						switch command {
						case ControlPause:
							// 记录当前状态并退出
							deviceState.StartStep = step
							deviceState.SceneStartIndex = i
							return
						case ControlStop:
							return
						}
					default:
					}
					// 这里应该调用语音播报接口
					PlayTextAction(deviceId, sceneResult)
				}
			}

			// 介绍自动化，从指定索引开始
			for j := automationIndex; j < len(automations); j++ {
				select {
				case command := <-deviceState.ControlChan:
					switch command {
					case ControlPause:
						// 记录当前状态并退出
						deviceState.StartStep = step
						deviceState.AutomationStartIndex = j
						return
					case ControlStop:
						return
					}
				default:
				}

				automation := automations[j]
				autoResult, err := chatCompletionInternal([]*chat.ChatMessage{
					{
						Role: "system",
						Content: fmt.Sprintf(`你是一个智能家居展示专家，现在你需要详细介绍一个自动化。
请用自然语言描述这个自动化的名称和触发条件及执行动作。
自动化信息：%s`, x.MustMarshalEscape2String(automation)),
					},
				})
				if err != nil {
					ava.Error(err)
				} else {
					select {
					case command := <-deviceState.ControlChan:
						switch command {
						case ControlPause:
							// 记录当前状态并退出
							deviceState.StartStep = step
							deviceState.AutomationStartIndex = j
							return
						case ControlStop:
							return
						}
					default:
					}
					// 这里应该调用语音播报接口
					PlayTextAction(deviceId, autoResult)
				}
			}
			step++
		}
	}
}
