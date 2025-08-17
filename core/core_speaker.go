package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"context"

	"github.com/aiakit/ava"
)

// Value:    "[" + message + ",false]", //这是发起指令的穿参数
func PlayTextAction(deviceID, message string) {
	entityId, ok := gSpeakerProcess.speakerEntityPlayText[deviceID]
	if !ok {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
		EntityId: entityId,
		Value:    message,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
	//暂停，等待播放完成
	time.Sleep(GetPlaybackDuration(message))
}

func PlayTextActionDirect(entityId, message string) {
	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
		EntityId: entityId,
		Value:    message,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

func PlayTextActionWithMemory(deviceID, message string) {
	entityId, ok := gSpeakerProcess.speakerEntityPlayText[deviceID]
	if !ok {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
		EntityId: entityId,
		Value:    message,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
	//暂停，等待播放完成
	time.Sleep(GetPlaybackDuration(message))

	AddXiaoaiMessage(deviceID, message)
}

func GetPlaybackDuration(message string) time.Duration {
	// 每个字符需要0.3秒播报
	charDuration := 130 * time.Millisecond

	// 计算总播报时间
	totalDuration := time.Duration(len(message)) * charDuration

	// 确保最短播报时间为1秒
	if totalDuration < 1*time.Second {
		totalDuration = 1 * time.Second
	}

	return totalDuration
}

type conversationor struct {
	Conversation []*chat.ChatMessage `json:"conversation"`
	entityId     string
	deviceId     string
}

type simpleEntity struct {
	Id       string `json:"id"`
	Name     string `json:"name"`      //音箱名称
	AreaName string `json:"area_name"` //所在区域名称
}

type speakerProcess struct {
	lock                        sync.Mutex
	deviceLocks                 map[string]*sync.Mutex // 每个设备独立的锁
	playTextMessage             chan *conversationor
	timeout                     time.Duration
	speakerEntityPlayText       map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityDirective      map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityMediaPlayer    map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityWakeUp         map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityPlayTextEntity map[string]*simpleEntity
	lastUpdateTime              map[string]time.Time
	// 添加会话状态管理
	activeSessions map[string]bool
	// 添加轮询控制
	pollCancelFuncs map[string]context.CancelFunc
	pollContexts    map[string]context.Context

	isReceivedLock     sync.RWMutex
	isReceivedPlayText map[string]int32
}

func getIsReceivedPlayText(entityId string) bool {
	gSpeakerProcess.isReceivedLock.RLock()
	flag := gSpeakerProcess.isReceivedPlayText[entityId]
	gSpeakerProcess.isReceivedLock.RUnlock()
	return atomic.LoadInt32(&flag) == 0
}

func setIsReceivedPlayText(entityId string, f int32) {
	gSpeakerProcess.isReceivedLock.Lock()
	flag := gSpeakerProcess.isReceivedPlayText[entityId]
	atomic.SwapInt32(&flag, f)
	gSpeakerProcess.isReceivedLock.Unlock()
}

var gSpeakerProcess *speakerProcess

func chaosSpeaker() {
	data.RegisterDataHandler(SpeakerAsk2ConversationHandler)
	data.RegisterDataHandler(SpeakerAsk2PlayTextHandler)

	gSpeakerProcess = &speakerProcess{
		deviceLocks:                 make(map[string]*sync.Mutex), // 初始化设备锁map
		playTextMessage:             make(chan *conversationor, 5),
		timeout:                     time.Second * 5,
		speakerEntityPlayText:       make(map[string]string),
		speakerEntityPlayTextEntity: make(map[string]*simpleEntity),
		speakerEntityDirective:      make(map[string]string),
		speakerEntityMediaPlayer:    make(map[string]string),
		speakerEntityWakeUp:         make(map[string]string),
		lastUpdateTime:              make(map[string]time.Time),
		// 初始化会话状态
		activeSessions:     make(map[string]bool),
		pollCancelFuncs:    make(map[string]context.CancelFunc),
		pollContexts:       make(map[string]context.Context),
		isReceivedPlayText: make(map[string]int32),
	}

	entitieXiaomiHome, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if !ok {
		return
	}
	entitieXiaomIot, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiMiotSpeaker]
	if !ok {
		return
	}

	for _, e := range entitieXiaomIot {

		for _, e1 := range entitieXiaomiHome {
			if e.Name == e1.Name && strings.Contains(e1.EntityID, "_play_text") {
				gSpeakerProcess.speakerEntityPlayText[e.DeviceID] = e1.EntityID
				gSpeakerProcess.speakerEntityPlayTextEntity[e.DeviceID] = &simpleEntity{
					Id:       e.DeviceID,
					Name:     e1.DeviceName,
					AreaName: data.SpiltAreaName(e1.AreaName),
				}
			}

			if e.Name == e1.Name && strings.Contains(e1.EntityID, "_execute_text_directive") {
				gSpeakerProcess.speakerEntityDirective[e.DeviceID] = e1.EntityID
			}

			if e.Name == e1.Name && strings.HasPrefix(e1.EntityID, "media_player.") {
				gSpeakerProcess.speakerEntityMediaPlayer[e.DeviceID] = e1.EntityID
			}

			if e.Name == e1.Name && strings.Contains(e1.EntityID, "_wake_up") {
				gSpeakerProcess.speakerEntityWakeUp[e.DeviceID] = e1.EntityID
			}
		}
	}

	go gSpeakerProcess.runSpeakerPlayText()
}

func speakerProcessSend(message *conversationor) {
	gSpeakerProcess.playTextMessage <- message
}

// 获取指定设备的锁，如果不存在则创建
func (s *speakerProcess) getDeviceLock(deviceId string) *sync.Mutex {
	// 保护对deviceLocks的并发访问
	s.lock.Lock()
	defer s.lock.Unlock()

	if mutex, exists := s.deviceLocks[deviceId]; exists {
		return mutex
	}

	// 创建新的mutex并存储
	mutex := &sync.Mutex{}
	s.deviceLocks[deviceId] = mutex
	return mutex
}

func (s *speakerProcess) runSpeakerPlayText() {
	var ticker = time.NewTicker(time.Second * 10)
	for {
		select {
		case message := <-s.playTextMessage:
			//todo: 增加基础指令拦截

			fmt.Println("---------2-2-2----", x.MustMarshalEscape2String(message.Conversation))

			if v := getOrCreateDeviceState(message.deviceId); v != nil && loadPlayingStats(message.deviceId) == 1 {
				continue
			}

			// 使用设备独立的锁
			deviceLock := s.getDeviceLock(message.deviceId)
			deviceLock.Lock()
			// 标记会话开始
			if !s.activeSessions[message.deviceId] {
				pausePlay(message.deviceId)
				s.activeSessions[message.deviceId] = true
				// 启动轮询
				fmt.Println("-------1-----", "会话开始，启动轮训，暂停播放")
				s.startPolling(message.deviceId)
			}
			s.lastUpdateTime[message.deviceId] = time.Now()
			s.sendToRemote(message)
			deviceLock.Unlock()

			var isHaveHuman bool
			// 添加消息到历史记录 (使用memory.go中的函数)
			for _, msg := range message.Conversation {
				switch msg.Role {
				case "user":
					AddUserMessage(message.deviceId, msg.Content)
					isHaveHuman = true
				case "assistant":
					if msg.Name == "jinx" {
						fmt.Println("----------1111", msg.Content)
						AddXiaoaiMessage(message.deviceId, msg.Content)
					} else {
						AddAIMessage(message.deviceId, msg.Content)
					}
				case "system":
					AddSystemMessage(message.deviceId, msg.Content)
				}
			}

			if !isHaveHuman {
				continue
			}

			//todo: 增加交互优化，如果5秒内没有收到消息，可以主动询问是否需要其他帮助，或者直接终止对话
		case <-ticker.C:
			for deviceId := range s.lastUpdateTime {
				s.getDeviceLock(deviceId).Lock()

				elapsed := time.Now().Sub(s.lastUpdateTime[deviceId])
				if elapsed > 10*time.Second && elapsed <= 20*time.Second {
					// 10秒内没有新消息，主动询问
					if s.activeSessions[deviceId] {
						// 唤醒设备并询问是否还需要帮助
						// 暂停轮询
						if cancel, exists := gSpeakerProcess.pollCancelFuncs[deviceId]; exists && cancel != nil {
							cancel()
							gSpeakerProcess.pollCancelFuncs[deviceId] = nil
							fmt.Println("--------3---", "暂停轮训")
						}

						PlayTextAction(deviceId, "主人，还需要其他帮助吗？")
						gSpeakerProcess.startPolling(deviceId)
						wakeup(deviceId)
						// 更新时间以避免重复询问
						s.lastUpdateTime[deviceId] = time.Now().Add(10 * time.Second)
					}
				} else if elapsed > 20*time.Second {
					// 20秒内没有新消息，结束会话
					if s.activeSessions[deviceId] {
						// 停止轮询
						if cancel, exists := s.pollCancelFuncs[deviceId]; exists && cancel != nil {
							cancel()
						}
						delete(s.activeSessions, deviceId)
						delete(s.pollCancelFuncs, deviceId)
						delete(s.pollContexts, deviceId)
					}
					delete(s.lastUpdateTime, deviceId)
				}
				s.getDeviceLock(deviceId).Unlock()
			}
		}
	}
}

// 修改:sendToRemote现在发送整个历史记录
func (s *speakerProcess) sendToRemote(conversations *conversationor) {

	//1.获取函数调用
	//2.发起调用,在处理函数中询问ai获取调用数据
	//3.发送通知

	var message string

	defer func() {
		if message == "" {
			for _, conversation := range conversations.Conversation {
				if conversation.Role == "assistant" {
					message = conversation.Content
					break
				}
			}
		}

		if message != "" {
			AddAIMessage(conversations.deviceId, message)
			fmt.Println("--------4---", "播放")

			// 暂停轮询
			if cancel, exists := gSpeakerProcess.pollCancelFuncs[conversations.deviceId]; exists && cancel != nil {
				cancel()
				gSpeakerProcess.pollCancelFuncs[conversations.deviceId] = nil
				fmt.Println("--------3---", "暂停轮训")
			}
			time.Sleep(time.Second)
			PlayTextAction(conversations.deviceId, message)
			PlayTextAction(conversations.deviceId, askMessage[x.Intn(len(askMessage)-1)])
			gSpeakerProcess.startPolling(conversations.deviceId)
			fmt.Println("--------8888---", "唤醒")
			wakeup(conversations.deviceId)
		}
	}()
	// 使用memory.go中的GetHistory函数获取历史记录
	prepare, err := prepareCall(conversations.Conversation, conversations.deviceId)
	if err != nil {
		message = "主人，请稍等，网络开小差了，请重试一次..."
		return
	}

	message = Call(findFunction(prepare), conversations.deviceId, conversations.Conversation[0].Content, prepare)
}

func downPlay(deviceId string) {

	entityId, ok := gSpeakerProcess.speakerEntityMediaPlayer[deviceId]
	if !ok {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_set", data.GetToken(), &data.HttpServiceDataPlay{
		EntityId:    entityId,
		VolumeLevel: 0,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

func upPlay(deviceId string) {

	entityId, ok := gSpeakerProcess.speakerEntityMediaPlayer[deviceId]
	if !ok {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_set", data.GetToken(), &data.HttpServiceDataPlay{
		EntityId:    entityId,
		VolumeLevel: 0.5,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

func pausePlay(deviceId string) {
	entityId, ok := gSpeakerProcess.speakerEntityMediaPlayer[deviceId]
	if !ok {
		return
	}

	// 不在这里获取锁，避免死锁
	// gSpeakerProcess.lock.Lock()
	// defer gSpeakerProcess.lock.Unlock()

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_mute", data.GetToken(), &data.HttpServiceDataPlayPause{
		EntityId:      entityId,
		IsVolumeMuted: true,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

// 主动唤醒逻辑
func wakeup(deviceId string) {
	entityId, ok := gSpeakerProcess.speakerEntityWakeUp[deviceId]
	if !ok {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/button/press", data.GetToken(), &data.HttpServiceData{
		EntityId: entityId,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

var askMessage = []string{
	"主人,还有什么我能为您效劳吗?",
	"您现在有什么需要帮的吗?我随时待命。",
	"我已准备好为您服务,请告诉我需要什么。",
	"尊敬的主人,您还有什么需求吗?我竭尽全力。",
	"我会诚心诚意为您服务,您现在有什么需要吗?",
	"请告诉我您还需要什么,我会全力以赴。",
	"如有需要,尽管告诉我,我会积极响应。",
	"主人,我时刻关注您,请告诉我还需要什么。",
	"尊敬的主人,您还有什么吩咐吗?",
	"主人,我会以最高效率满足您的需求,请告诉我吧。",
	"我会专注倾听您的要求,还需要什么吗?",
	"主人,您现在有什么需要我帮的吗?",
	"有其他需求,就告诉我吧。",
	"主人，还需要什么请告诉我。",
	"尊敬的主人,我会以诚意为您服务,您还有什么需要吗?",
	"主人,我会全力满足您的需求,请告诉我吧。",
	"如有任何需要,尽管告诉我,我全心全意服务。",
	"主人,我会以热忱为您效劳,您还有什么需要吗?",
	"尊敬的主人,请告诉我您还需要什么。",
	"我在倾听您的要求,请告诉我吧。",
}

func SpeakerAsk2PlayTextHandler(event *data.StateChangedSimple, body []byte) {

	// 播放文本实体后面是play_text
	var state chatMessage
	err := x.Unmarshal(body, &state)
	if err != nil {
		ava.Error(err)
		return
	}

	if strings.Contains(state.Event.Data.EntityID, "_play_text") && strings.HasPrefix(state.Event.Data.EntityID, "text.") {
		v, ok := data.GetEntityIdMap()[state.Event.Data.EntityID]
		if !ok {
			return
		}

		if !getIsReceivedPlayText(v.DeviceID) {
			return
		}

		en, ok := data.GetEntityIdMap()[state.Event.Data.EntityID]
		if !ok {
			return
		}

		speakerProcessSend(&conversationor{
			Conversation: []*chat.ChatMessage{{Role: "assistant", Content: state.Event.Data.NewState.State}},
			entityId:     en.EntityID,
			deviceId:     en.DeviceID,
		})
	}
}

// 获取对话记录,entity_id相同
func SpeakerAsk2ConversationHandler(event *data.StateChangedSimple, body []byte) {
	// 播放文本实体后面是play_text
	var state chatMessage
	err := x.Unmarshal(body, &state)
	if err != nil {
		ava.Error(err)
		return
	}

	if strings.Contains(state.Event.Data.EntityID, "_conversation") &&
		strings.EqualFold(state.Event.Data.NewState.Attributes.EntityClass, "XiaoaiConversationSensor") {
		en, ok := data.GetEntityIdMap()[state.Event.Data.EntityID]
		if !ok {
			return
		}

		v := state.Event.Data.NewState.Attributes.Answers
		if len(v) == 0 {
			return
		}
		fmt.Println("-------999999---", string(body))
		var content = v[0].TTS.Text
		if content == "" {
			content = v[0].Llm.Text
		}

		var cs = &conversationor{
			Conversation: []*chat.ChatMessage{{
				Role:    "user",
				Content: state.Event.Data.NewState.State,
			}, {
				Role:    "assistant",
				Content: v[0].TTS.Text,
				Name:    "jinx",
			}},
			entityId: en.EntityID,
			deviceId: en.DeviceID,
		}

		speakerProcessSend(cs)
	}
}

type chatMessage struct {
	Type  string `json:"type"`
	Event struct {
		EventType string `json:"event_type"`
		Data      struct {
			EntityID string `json:"entity_id"`
			OldState struct {
				EntityID   string `json:"entity_id"`
				State      string `json:"state"`
				Attributes struct {
					EntityClass    string `json:"entity_class"`
					ParentEntityID string `json:"parent_entity_id"`
					Content        string `json:"content"`
					Answers        []struct {
						Type string `json:"type"`
						Llm  struct {
							Text string `json:"text"`
						} `json:"llm"`
					} `json:"answers"`
					History           []string  `json:"history"`
					Timestamp         time.Time `json:"timestamp"`
					Icon              string    `json:"icon"`
					FriendlyName      string    `json:"friendly_name"`
					SupportedFeatures int       `json:"supported_features"`
				} `json:"attributes"`
				LastChanged  time.Time `json:"last_changed"`
				LastReported time.Time `json:"last_reported"`
				LastUpdated  time.Time `json:"last_updated"`
				Context      struct {
					ID       string `json:"id"`
					ParentID any    `json:"parent_id"`
					UserID   any    `json:"user_id"`
				} `json:"context"`
			} `json:"old_state"`
			NewState struct {
				EntityID   string `json:"entity_id"`
				State      string `json:"state"`
				Attributes struct {
					EntityClass    string `json:"entity_class"`
					ParentEntityID string `json:"parent_entity_id"`
					Content        string `json:"content"`
					Answers        []struct {
						Type string `json:"type"`
						TTS  struct {
							Text string `json:"text"`
						} `json:"tts"`
						Llm struct {
							Text string `json:"text"`
						} `json:"llm"`
					} `json:"answers"`
					History           []string  `json:"history"`
					Timestamp         time.Time `json:"timestamp"`
					Icon              string    `json:"icon"`
					FriendlyName      string    `json:"friendly_name"`
					SupportedFeatures int       `json:"supported_features"`
				} `json:"attributes"`
				LastChanged  time.Time `json:"last_changed"`
				LastReported time.Time `json:"last_reported"`
				LastUpdated  time.Time `json:"last_updated"`
				Context      struct {
					ID       string `json:"id"`
					ParentID any    `json:"parent_id"`
					UserID   any    `json:"user_id"`
				} `json:"context"`
			} `json:"new_state"`
		} `json:"data"`
		Origin    string    `json:"origin"`
		TimeFired time.Time `json:"time_fired"`
		Context   struct {
			ID       string `json:"id"`
			ParentID any    `json:"parent_id"`
			UserID   any    `json:"user_id"`
		} `json:"context"`
	} `json:"event"`
	ID int `json:"id"`
}

func (s *speakerProcess) startPolling(deviceId string) {
	// 如果已有轮询在运行，先取消它
	if cancel, exists := s.pollCancelFuncs[deviceId]; exists && cancel != nil {
		cancel()
	}

	// 创建新的上下文
	ctx, cancel := context.WithCancel(context.Background())
	s.pollContexts[deviceId] = ctx
	s.pollCancelFuncs[deviceId] = cancel

	// 启动轮询goroutine
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 在轮询中调用pausePlay时不需要获取锁
				// 因为这可能与其他需要锁的操作形成死锁
				pausePlay(deviceId)
			}
		}
	}()
}
