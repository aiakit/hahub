package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/internal/x"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

//小米音箱
//1.播放文本
//2.执行文本指令

// PlayText 播放文本
func PlayText(deviceID string, text string) *intelligent.ActionNotify {
	// 实现播放文本的逻辑
	return &intelligent.ActionNotify{
		Action: "notify.send_message",
		Data: struct {
			Message string `json:"message,omitempty"`
			Title   string `json:"title,omitempty"`
		}{Message: text, Title: ""},
		Target: struct {
			DeviceID string `json:"device_id,omitempty"`
		}{DeviceID: deviceID},
	}
}

func PlayTextAction(deviceID, entityId, message string) {
	//调整音量
	upPlay(deviceID)
	fmt.Println("----0-----", "调高音量,播报文本", time.Now(), entityId)
	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/text/set_value", data.GetToken(), &data.HttpServiceData{
		EntityId: entityId,
		Value:    message,
	}, nil)
	if err != nil {
		ava.Error(err)
	}

	time.Sleep(GetPlaybackDuration(message))
	//唤醒之前把音量调最小

	downPlay(deviceID)
	fmt.Println("----1--", "调低音量,唤醒", time.Now(), entityId)

	wakeup(deviceID)
}

func GetPlaybackDuration(message string) time.Duration {
	// 每个字符需要0.3秒播报
	charDuration := 100 * time.Millisecond

	// 计算总播报时间
	totalDuration := time.Duration(len(message)) * charDuration

	// 确保最短播报时间为1秒
	if totalDuration < 1*time.Second {
		totalDuration = 1 * time.Second
	}

	return totalDuration
}

//func PlayControlAction(deviceID, entityId, message string) {
//	err := core.Post(ava.Background(), core.GetHassUrl()+"/api/services/text/set_value", core.GetToken(), &core.HttpServiceData{
//		EntityId: entityId,
//		Value:    "[" + message + ",false]",
//	}, nil)
//	if err != nil {
//		ava.Error(err)
//	}
//
//	core.TimingwheelAfter(GetPlaybackDuration(message), func() {
//		//进入唤醒状态
//		wakeup(deviceID)
//
//		//暂停
//		core.TimingwheelAfter(GetPlaybackDuration(message), func() {
//			pausePlay(entityId)
//		})
//	})
//}

// ExecuteTextCommand 执行文本指令
func ExecuteTextCommand(entityId string, command string, silent bool) *intelligent.ActionService {
	// 实现执行文本指令的逻辑
	return &intelligent.ActionService{
		Action: "text.set_value",
		Data:   map[string]interface{}{"value": fmt.Sprintf("[%s,%v]", command, silent)},
		Target: &struct {
			EntityId string `json:"entity_id"`
		}{EntityId: entityId},
	}
}

type conversationor struct {
	Conversation []*chat.ChatMessage `json:"conversation"`
	entityId     string
	deviceId     string
}

type speakerProcess struct {
	lock            sync.Mutex
	playTextMessage chan *conversationor
	timeout         time.Duration
	callbacks       map[string]func(*http.Response, error)
	history         []*conversationor
	speakerEntity   map[string][]*data.Entity
	lastUpdateTime  map[string]time.Time
}

var gSpeakerProcess *speakerProcess

func ChaosSpeaker() {
	data.RegisterDataHandler(SpeakerAsk2manAction4HomingHandler)

	gSpeakerProcess = &speakerProcess{
		playTextMessage: make(chan *conversationor, 5),
		timeout:         time.Second * 5,
		callbacks:       make(map[string]func(*http.Response, error)),
		history:         make([]*conversationor, 0, 5), // 初始化容量为5的切片
		speakerEntity:   make(map[string][]*data.Entity),
		lastUpdateTime:  make(map[string]time.Time),
	}

	entitieXiaomiHome, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if !ok {
		return
	}
	entitieXiaomIot, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiMiotSpeaker]
	if !ok {
		return
	}

	for _, e := range entitieXiaomiHome {
		gSpeakerProcess.speakerEntity[e.DeviceID] = append(gSpeakerProcess.speakerEntity[e.DeviceID], e)
	}

	for _, e := range entitieXiaomIot {
		gSpeakerProcess.speakerEntity[e.DeviceID] = append(gSpeakerProcess.speakerEntity[e.DeviceID], e)
	}

	// 注册不同的业务逻辑处理
	gSpeakerProcess.RegisterCallback("play_text", func(resp *http.Response, err error) {
		if err != nil {
			fmt.Println("Error sending to remote:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Conversations sent to remote successfully")
		} else {
			fmt.Println("Remote server returned error:", resp.Status)
			// 可以添加重试或其他错误处理逻辑
		}
	})

	// 注册其他业务逻辑的回调函数

	go gSpeakerProcess.runSpeakerPlayText()
}

func (s *speakerProcess) RegisterCallback(businessID string, callback func(*http.Response, error)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.callbacks[businessID] = callback
}

func speakerProcessSend(message *conversationor) {
	gSpeakerProcess.playTextMessage <- message
}

func (s *speakerProcess) runSpeakerPlayText() {
	var ticker = time.NewTicker(time.Second * 5)

	for {
		select {
		case message := <-s.playTextMessage:

			pausePlay(message.deviceId)

			// 添加消息到历史记录
			s.addToHistory(message)

			s.lock.Lock()
			// 发送历史记录
			// 修改:在调用PlayTextAction之前发送时间到channel
			s.lastUpdateTime[message.deviceId] = time.Now()
			s.sendToRemote(s.history)
			s.lock.Unlock()
		case <-ticker.C:
			for deviceId := range s.lastUpdateTime {
				if time.Now().Sub(s.lastUpdateTime[deviceId]) > 5*time.Minute {
					upPlay(deviceId)
					s.lock.Lock()
					delete(s.lastUpdateTime, deviceId)
					s.lock.Unlock()
				}
			}
		}
	}
}

// 修改:添加addToHistory方法来维护历史记录
func (s *speakerProcess) addToHistory(conv *conversationor) {
	s.lock.Lock()
	// 如果历史记录已满(5条)，移除最旧的记录
	if len(s.history) >= 5 {
		// 移除第一个元素(最旧的记录)
		s.history = s.history[1:]
	}

	// 添加新记录
	s.history = append(s.history, conv)
	s.lock.Unlock()
}

func GetHistory() []*conversationor {
	if gSpeakerProcess == nil {
		return nil
	}
	gSpeakerProcess.lock.Lock()

	history := gSpeakerProcess.history
	gSpeakerProcess.lock.Unlock()
	return history
}

// 修改:sendToRemote现在发送整个历史记录
func (s *speakerProcess) sendToRemote(conversations []*conversationor) {

	//模拟请求服务返回的让音箱播报的内容情景
	time.Sleep(time.Second * 3)

	PlayTextAction("cd5be8092557a0dcd20162114ad99de3", "text.xiaomi_cn_865393253_lx06_play_text_a_5_1", "主人，还有什么需要帮助的吗？")
	//// 根据业务 ID 执行不同的回调函数
	//businessID := s.getBusinessID(conversations)
	//if callback, ok := s.callbacks[businessID]; ok {
	//	callback(resp, err)
	//} else {
	//	fmt.Println("No callback found for business ID:", businessID)
	//}
}

// 修改:getBusinessID现在处理对话历史数组
func (s *speakerProcess) getBusinessID(conversations []*chat.ChatMessage) string {
	// 根据对话内容获取业务 ID
	for _, conv := range conversations {
		if conv.Role == "user" {
			return strings.Split(conv.Content, ",")[0]
		}
	}
	return ""
}

func downPlay(deviceId string) {

	var playEntityId string
	for _, e := range gSpeakerProcess.speakerEntity[deviceId] {
		if e.Category == data.CategoryXiaomiHomeSpeaker && e.DeviceID == deviceId && strings.HasPrefix(e.EntityID, "media_player.") {
			playEntityId = e.EntityID
		}
	}
	if playEntityId == "" {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_set", data.GetToken(), &data.HttpServiceDataPlay{
		EntityId:    playEntityId,
		VolumeLevel: 0,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

func upPlay(deviceId string) {

	var playEntityId string
	for _, e := range gSpeakerProcess.speakerEntity[deviceId] {
		if e.Category == data.CategoryXiaomiHomeSpeaker && e.DeviceID == deviceId && strings.HasPrefix(e.EntityID, "media_player.") {
			playEntityId = e.EntityID
		}
	}
	if playEntityId == "" {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_set", data.GetToken(), &data.HttpServiceDataPlay{
		EntityId:    playEntityId,
		VolumeLevel: 0.5,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

func pausePlay(deviceId string) {

	var playEntityId string
	for _, e := range gSpeakerProcess.speakerEntity[deviceId] {
		if e.Category == data.CategoryXiaomiHomeSpeaker && e.DeviceID == deviceId && strings.HasPrefix(e.EntityID, "media_player.") {
			playEntityId = e.EntityID
		}
	}
	if playEntityId == "" {
		return
	}

	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/media_player/volume_mute", data.GetToken(), &data.HttpServiceDataPlayPause{
		EntityId:      playEntityId,
		IsVolumeMuted: true,
	}, nil)
	if err != nil {
		ava.Error(err)
	}
}

// 主动唤醒逻辑
func wakeup(deviceId string) {
	var wakeupEntityId string
	for _, e := range gSpeakerProcess.speakerEntity[deviceId] {
		if e.Category == data.CategoryXiaomiHomeSpeaker && e.DeviceID == deviceId && strings.Contains(e.OriginalName, "唤醒") {
			wakeupEntityId = e.EntityID
		}
	}
	if wakeupEntityId == "" {
		return
	}
	err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/button/press", data.GetToken(), &data.HttpServiceData{
		EntityId: wakeupEntityId,
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

// 获取对话记录,entity_id相同
func SpeakerAsk2manAction4HomingHandler(event *data.StateChangedSimple, body []byte) {
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

		var cs = &conversationor{
			Conversation: []*chat.ChatMessage{{
				Role:    "user",
				Content: state.Event.Data.NewState.State,
			}, {
				Role:    "assistant",
				Content: v[0].Llm.Text,
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
