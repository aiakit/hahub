package core

import (
	"context"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aiakit/ava"
)

// Value:    "[" + message + ",false]", //这是发起指令的穿参数
func PlayTextAction(deviceID, message string) {
	entityId, ok := gSpeakerProcess.speakerEntityPlayText[deviceID]
	if !ok {
		return
	}

	// 如果消息长度超过200字符，则拆分为多个片段
	if len(message) > 800 {
		// 按200字符拆分消息
		for i := 0; i < len(message); i += 800 {
			end := i + 800
			if end > len(message) {
				end = len(message)
			}

			err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
				EntityId: entityId,
				Message:  message[i:end],
			}, nil)
			if err != nil {
				ava.Error(err)
			}
			// 暂停，等待播放完成
			time.Sleep(GetPlaybackDuration(message[i:end]))
		}
	} else {
		err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
			EntityId: entityId,
			Message:  message,
		}, nil)
		if err != nil {
			ava.Error(err)
		}
		// 暂停，等待播放完成
		time.Sleep(GetPlaybackDuration(message))
	}
}

func GetPlaybackDuration(message string) time.Duration {
	// 设置中文字符和非中文字符的播报时间
	var (
		chineseCharDuration    = 130 * time.Millisecond
		nonChineseCharDuration = 200 * time.Millisecond
		minPlaybackDuration    = 2 * time.Second
	)

	// 计算总播报时间
	var totalDuration time.Duration
	var isDot bool

	for _, char := range message {
		if !isChineseChar(char) {
			isDot = true
			break
		}
	}
	if !isDot {
		chineseCharDuration = 180
	}

	for _, char := range message {
		if isChineseChar(char) {
			totalDuration += chineseCharDuration
		} else {
			totalDuration += nonChineseCharDuration
		}
	}

	// 确保最短播报时间为1秒
	if totalDuration < minPlaybackDuration {
		totalDuration = minPlaybackDuration
	}

	return totalDuration
}

// 判断字符是否为中文字符
func isChineseChar(char rune) bool {
	// 使用正则表达式检测中文字符范围
	matched, _ := regexp.MatchString(`[\u4e00-\u9fa5]`, string(char))
	return matched
}

type Conversationor struct {
	Conversation *chat.ChatMessage `json:"conversation"`
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
	playTextMessage             chan *Conversationor
	timeout                     time.Duration
	speakerEntityPlayText       map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityDirective      map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityMediaPlayer    map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityWakeUp         map[string]string //xiaomi_iot_device_id:xiaomi_home
	speakerEntityPlayTextEntity map[string]*simpleEntity
	lastUpdateTime              map[string]time.Time
	// 添加会话状态管理
	// 添加轮询控制
	pollCancelFuncs map[string]context.CancelFunc

	aiSwitchLock sync.RWMutex
	aiSwitch     map[string]int32
}

func aiIsLock(id string) bool {
	gSpeakerProcess.aiSwitchLock.RLock()
	flag := gSpeakerProcess.aiSwitch[id]
	gSpeakerProcess.aiSwitchLock.RUnlock()
	return atomic.LoadInt32(&flag) != 0
}

func aiLock(id string) {
	gSpeakerProcess.aiSwitchLock.Lock()
	flag := gSpeakerProcess.aiSwitch[id]
	atomic.SwapInt32(&flag, 1)
	gSpeakerProcess.aiSwitchLock.Unlock()
}

func aiUnlock(id string) {
	gSpeakerProcess.aiSwitchLock.Lock()
	flag := gSpeakerProcess.aiSwitch[id]
	atomic.SwapInt32(&flag, 0)
	gSpeakerProcess.aiSwitchLock.Unlock()
}

var gSpeakerProcess *speakerProcess

func chaosSpeaker() {
	data.RegisterDataHandler(SpeakerAsk2ConversationHandler)
	data.RegisterDataHandler(SpeakerAsk2PlayTextHandler)

	gSpeakerProcess = &speakerProcess{
		deviceLocks:                 make(map[string]*sync.Mutex), // 初始化设备锁map
		playTextMessage:             make(chan *Conversationor, 5),
		timeout:                     time.Second * 5,
		speakerEntityPlayText:       make(map[string]string),
		speakerEntityPlayTextEntity: make(map[string]*simpleEntity),
		speakerEntityDirective:      make(map[string]string),
		speakerEntityMediaPlayer:    make(map[string]string),
		speakerEntityWakeUp:         make(map[string]string),
		lastUpdateTime:              make(map[string]time.Time),
		pollCancelFuncs:             make(map[string]context.CancelFunc),
		aiSwitch:                    make(map[string]int32),
	}

	entitieXiaomiHome, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if !ok {
		return
	}
	entitieXiaomIot, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiMiotSpeaker]
	if !ok {
		return
	}

	for _, e1 := range entitieXiaomiHome {
		for _, e := range entitieXiaomIot {
			if e1.DeviceName == e.DeviceName && strings.Contains(e1.EntityID, "_play_text_") && strings.HasPrefix(e1.EntityID, "notify.") {
				gSpeakerProcess.speakerEntityPlayText[e.DeviceID] = e1.EntityID
				gSpeakerProcess.speakerEntityPlayTextEntity[e.DeviceID] = &simpleEntity{
					Id:       e1.EntityID,
					Name:     e1.DeviceName,
					AreaName: data.SpiltAreaName(e1.AreaName),
				}
				break
			}

			if e1.DeviceName == e.DeviceName && strings.Contains(e1.EntityID, "_execute_text_directive") {
				gSpeakerProcess.speakerEntityDirective[e.DeviceID] = e1.EntityID
				break
			}

			if e1.DeviceName == e.DeviceName && strings.HasPrefix(e1.EntityID, "media_player.") {
				gSpeakerProcess.speakerEntityMediaPlayer[e.DeviceID] = e1.EntityID
				break
			}

			if e1.DeviceName == e.DeviceName && strings.Contains(e1.EntityID, "_wake_up") {
				gSpeakerProcess.speakerEntityWakeUp[e.DeviceID] = e1.EntityID
				break
			}
		}
	}

	go gSpeakerProcess.runSpeakerPlayText()
}

func getAreaName(deviceId string) string {
	if v, ok := gSpeakerProcess.speakerEntityPlayTextEntity[deviceId]; ok && v.AreaName != "" {
		return data.SpiltAreaName(v.AreaName)
	}

	return ""
}

func SpeakerProcessSend(message *Conversationor) {
	var msg = message.Conversation.Content

	if strings.Contains(msg, "救") {
		intelligent.RunSript("script.sos")
	}

	if aiIsLock(message.deviceId) {
		if strings.Contains(msg, "开启智能管家") {
			aiUnlock(message.deviceId)
		}
		isRunningLock.Lock()
		if descriptionIsRunning {
			descriptionIsRunning = false
			aiUnlock(message.deviceId)
		}
		isRunningLock.Unlock()
	}

	if !aiIsLock(message.deviceId) {
		gSpeakerProcess.playTextMessage <- message
		if strings.Contains(msg, "关闭智能管家") {
			aiLock(message.deviceId)
		}
	}
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

// 过滤小爱已经成功处理的关键词
var filterMessage = []string{
	"好的", "发送指令", "已", "收到", "正在为", "搞定", "没问题", "正在", "你有好几个设备",
	"关咯", "空气质量",
}

func (s *speakerProcess) runSpeakerPlayText() {
	for {
		select {
		case message := <-s.playTextMessage:
			go s.sendToRemote(message)
		}
	}
}

// 拦截器，避免向ai发送的数据过多，影响响应时间
var interceptorLock sync.RWMutex
var interceptorCall = make(map[string]func(messageInput *chat.ChatMessage, deviceId string) string, 2)

// 修改:sendToRemote现在发送整个历史记录
func (s *speakerProcess) sendToRemote(conversations *Conversationor) {
	deviceLock := s.getDeviceLock(conversations.deviceId)
	deviceLock.Lock()
	defer deviceLock.Unlock()

	switch conversations.Conversation.Role {
	case "user":
		sendMessage2Panel("input_text.my_input_text_1", conversations.Conversation.Content)
	case "assistant":
		sendMessage2Panel("input_text.my_input_text_2", conversations.Conversation.Content)
	}

	//1.获取函数调用
	//2.发起调用,在处理函数中询问ai获取调用数据
	//3.发送通知

	s.lastUpdateTime[conversations.deviceId] = time.Now()

	var message string
	var ifAsk = true

	defer func() {
		if len(message) > 1000 {
			message = "宿主，你要的内容太长了..."
		}

		if message != "" {
			sendMessage2Panel("input_text.my_input_text_2", message)
			// 暂停轮询
			if cancel, exists := gSpeakerProcess.pollCancelFuncs[conversations.deviceId]; exists && cancel != nil {
				cancel()
				gSpeakerProcess.pollCancelFuncs[conversations.deviceId] = nil
			}

			PlayTextAction(conversations.deviceId, message)

			if ifAsk {
				PlayTextAction(conversations.deviceId, askMessage[x.Intn(len(askMessage)-1)])
			}

			gSpeakerProcess.startPolling(conversations.deviceId)
			wakeup(conversations.deviceId)
		}
	}()

	interceptorLock.RLock()
	// 检查并执行拦截器逻辑
	if len(interceptorCall) > 0 {
		// 有拦截器，优先执行拦截器
		if interceptor, exists := interceptorCall[conversations.deviceId]; exists {
			message = interceptor(conversations.Conversation, conversations.deviceId)
			interceptorLock.RUnlock()
			// 使用完后删除拦截器
			interceptorLock.Lock()
			delete(interceptorCall, conversations.deviceId)
			interceptorLock.Unlock()
			return
		}
	}
	interceptorLock.RUnlock()

	//todo 拦截器,走拦截器逻辑
	prepare, err := prepareCall(conversations.Conversation, conversations.deviceId)
	if err != nil {
		ava.Error(err)
		message = "宿主，请稍等，网络开小差了，请重试一次..."
		return
	}

	if prepare == "" {
		return
	}

	message = Call(findFunction(prepare), conversations.deviceId, conversations.Conversation.Content, prepare)

	interceptorLock.RLock()
	if interceptorCall[conversations.deviceId] != nil {
		ifAsk = false
	}
	interceptorLock.RUnlock()
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

	err := x.PostWithOutLog(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_mute", data.GetToken(), &data.HttpServiceDataPlayPause{
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
	"还有什么我能为您效劳的吗?",
	"还有什么需要帮助吗?我随时待命。",
	"尊敬的宿主,您还有什么需要吗?",
	"您还有什么需要吗?",
	"请告诉我您还需要我为你做什么。",
	"如有需要,尽管告诉我。",
	"宿主,请告诉我还需要什么帮助。",
	"宿主,您还有什么吩咐吗?",
	"宿主,您还有什么需要我帮的吗?",
	"有其他需要,就告诉我。",
	"宿主，还需要什么帮助。",
	"尊敬的宿主,您还有什么需要吗?",
	"如有任何需要,尽管告诉我。",
	"宿主,您还有什么需要吗?",
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
		en, ok := data.GetEntityByEntityId()[state.Event.Data.EntityID]
		if !ok {
			return
		}

		var deviceId string

		//查找设备id
		if v, ok := data.GetEntityByEntityId()[en.EntityID]; ok {
			deviceId = v.DeviceID
		} else {
			return
		}

		if state.Event.Data.NewState.State == "" {
			return
		}

		cs := &Conversationor{
			Conversation: &chat.ChatMessage{Role: "assistant", Content: state.Event.Data.NewState.State},
			deviceId:     deviceId,
		}
		ava.Debugf("speaker ask2 text: %s |data=%s", x.MustMarshal2String(cs), string(body))

		SpeakerProcessSend(cs)
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
		en, ok := data.GetEntityByEntityId()[state.Event.Data.EntityID]
		if !ok {
			return
		}

		v := state.Event.Data.NewState.Attributes.Answers
		if len(v) == 0 {
			return
		}
		var content = v[0].TTS.Text
		if content == "" {
			content = v[0].Llm.Text
		}
		userMsg := state.Event.Data.NewState.State
		if userMsg == "" {
			return
		}

		ava.Debugf("SpeakerAsk2ConversationHandler |小爱=%s |用户=%s", content, userMsg)

		for _, f := range filterMessage {
			if strings.Contains(content, f) && !strings.Contains(userMsg, "场景") && !strings.Contains(userMsg, "自动化") {
				return
			}
		}

		if strings.Contains(state.Event.Data.NewState.State, "扫地机器人") &&
			(strings.Contains(state.Event.Data.NewState.State, "开始") || strings.Contains(state.Event.Data.NewState.State, "启动")) {
			s := data.GetEntityCategoryMap()[data.CategoryInputBoolean]
			if len(s) > 0 {
				for _, e := range s {
					if strings.Contains(e.OriginalName, "扫地机器人") {
						x.PostWithOutLog(ava.Background(), data.GetHassUrl()+"/api/services/input_boolean/turn_on", data.GetToken(), &data.HttpServiceData{
							EntityId: "",
						}, nil)
					}
				}
			}
		}

		if strings.Contains(state.Event.Data.NewState.State, "扫地机器人") &&
			(strings.Contains(state.Event.Data.NewState.State, "停止") || strings.Contains(state.Event.Data.NewState.State, "回")) {
			s := data.GetEntityCategoryMap()[data.CategoryInputBoolean]
			if len(s) > 0 {
				for _, e := range s {
					if strings.Contains(e.OriginalName, "扫地机器人") {
						x.PostWithOutLog(ava.Background(), data.GetHassUrl()+"/api/services/input_boolean/turn_off", data.GetToken(), &data.HttpServiceData{
							EntityId: "",
						}, nil)
					}
				}
			}
		}

		var deviceId string

		//查找设备id
		if v, ok := data.GetEntityByEntityId()[en.EntityID]; ok {
			deviceId = v.DeviceID
		} else {
			return
		}

		SpeakerProcessSend(&Conversationor{
			Conversation: &chat.ChatMessage{
				Role:    "user",
				Content: userMsg,
			},
			deviceId: deviceId,
		})
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
	s.pollCancelFuncs[deviceId] = cancel

	// 启动轮询goroutine
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		ticker1 := time.NewTicker(time.Second * 20)

		defer ticker.Stop()
		defer ticker1.Stop()
		defer func() {
			sendMessage2Panel("input_text.my_input_text_1", " ")
			sendMessage2Panel("input_text.my_input_text_2", " ")
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 在轮询中调用pausePlay时不需要获取锁
				// 因为这可能与其他需要锁的操作形成死锁
				pausePlay(deviceId)
			case <-ticker1.C:
				// 检查最后一次更新时间是否超过20秒
				s.getDeviceLock(deviceId).Lock()
				lastUpdate, exists := s.lastUpdateTime[deviceId]
				s.getDeviceLock(deviceId).Unlock()

				if exists && time.Since(lastUpdate) > time.Second*20 {
					// 超过20秒，暂停轮询并执行收尾工作
					if cancel, exists := s.pollCancelFuncs[deviceId]; exists && cancel != nil {
						cancel()
						s.pollCancelFuncs[deviceId] = nil
					}
					return
				}
			}
		}
	}()
}

func sendMessage2Panel(entityId string, message string) {

	// 计算要发送的消息数量
	var chunkSize = 25
	runes := []rune(message) // 转换为 rune 切片，以正确处理中文字符

	for len(runes) > 0 {
		// 如果剩余的字符长度小于或等于 chunkSize，直接发送剩余的消息
		if len(runes) <= chunkSize {
			x.Post(ava.Background(), data.GetHassUrl()+"/api/services/input_text/set_value", data.GetToken(), &data.HttpServiceData{
				EntityId: entityId,
				Value:    string(runes), // 转回字符串
			}, nil)
			break
		}

		// 截取前 chunkSize 个字符
		chunk := runes[:chunkSize]
		x.Post(ava.Background(), data.GetHassUrl()+"/api/services/input_text/set_value", data.GetToken(), &data.HttpServiceData{
			EntityId: entityId,
			Value:    string(chunk), // 转回字符串
		}, nil)

		// 更新 runes，去掉已发送的部分
		runes = runes[chunkSize:]

		if len(runes) > 0 {
			time.Sleep(time.Second * 2)
		}
	}
}
