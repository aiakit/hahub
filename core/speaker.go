package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

// Value:    "[" + message + ",false]", //这是发起指令的穿参数
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

type conversationor struct {
	Conversation []*chat.ChatMessage `json:"conversation"`
	entityId     string
	deviceId     string
}

type speakerProcess struct {
	lock            sync.Mutex
	playTextMessage chan *conversationor
	timeout         time.Duration
	speakerEntity   map[string][]*data.Entity
	lastUpdateTime  map[string]time.Time
}

var gSpeakerProcess *speakerProcess

func ChaosSpeaker() {
	data.RegisterDataHandler(SpeakerAsk2manAction4HomingHandler)

	gSpeakerProcess = &speakerProcess{
		playTextMessage: make(chan *conversationor, 5),
		timeout:         time.Second * 5,
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

	go gSpeakerProcess.runSpeakerPlayText()
}

func speakerProcessSend(message *conversationor) {
	gSpeakerProcess.playTextMessage <- message
}

func (s *speakerProcess) runSpeakerPlayText() {
	var ticker = time.NewTicker(time.Second * 10)

	for {
		select {
		case message := <-s.playTextMessage:
			//todo: 如何判断处理同一轮对话当中
			//todo: 增加基础指令拦截

			pausePlay(message.deviceId)

			// 添加消息到历史记录 (使用memory.go中的函数)
			for _, msg := range message.Conversation {
				switch msg.Role {
				case "user":
					AddUserMessage(message.deviceId, msg.Content)
				case "assistant":
					AddAIMessage(message.deviceId, msg.Content)
				case "system":
					AddSystemMessage(message.deviceId, msg.Content)
				}
			}

			s.lock.Lock()
			s.lastUpdateTime[message.deviceId] = time.Now()
			s.sendToRemote(message)
			s.lock.Unlock()

			//todo: 增加交互优化，如果5秒内没有收到消息，可以主动询问是否需要其他帮助，或者直接终止对话
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

// 修改:sendToRemote现在发送整个历史记录
func (s *speakerProcess) sendToRemote(conversations *conversationor) {

	//1.获取函数调用
	//2.发起调用,在处理函数中询问ai获取调用数据
	//3.发送通知

	var message string

	defer func() {
		PlayTextAction(conversations.deviceId, conversations.entityId, message)
	}()
	// 使用memory.go中的GetHistory函数获取历史记录
	history := GetHistory(conversations.deviceId)
	prepare, err := prepareCall(history)
	if err != nil {
		message = "主人，请稍等，网络开小差了，请重试一次..."
		return
	}

	message = Call(findFunction(prepare), conversations.deviceId)
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
