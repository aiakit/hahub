package data

import (
	"fmt"
	"hahub/x"
	urlpkg "net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aiakit/ava"
	"github.com/gorilla/websocket"

	jsoniter "github.com/json-iterator/go"
)

var defaultToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIwMTZkZmM2ZDEwMTg0ZTRjYjJkMDBkMDUzMTYwNmFmZSIsImlhdCI6MTc1MTM0ODQyNCwiZXhwIjoyMDY2NzA4NDI0fQ.2W03gIpG2mJaYUPuT0OGST8zFN1paJ40ltFE9WG52Yg"

//var defaultToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxN2RjYzk2NjJlMGU0MDQ4ODJjMjE4MmRhZWFlYzE1NiIsImlhdCI6MTc1NDk2NTg1NCwiZXhwIjoyMDcwMzI1ODU0fQ.TLBye1wzFQTxwb46fI14_PdvUltgYTZDY6_zmsrQ1LE"

// 添加初始化完成信号
var initDone = make(chan struct{})

// 添加初始化状态跟踪
var (
	initMutex sync.Mutex
	initState = struct {
		areasLoaded    bool
		devicesLoaded  bool
		entitiesLoaded bool
		servicesLoaded bool
		statesLoaded   bool
		wsConnected    bool
	}{}
)

func GetToken() string {
	return defaultToken
}

func GetHassUrl() string {
	return defaultURL
}

func getHostAndPath() string {
	u, err := urlpkg.Parse(GetHassUrl())
	if err != nil {
		return ""
	}
	return u.Host + u.Path
}

const defaultURL = "http://homeassistant.local:8123"

type hub struct {
	conn   *websocket.Conn
	lock   *sync.RWMutex
	idLock *sync.Mutex

	callbackMapFunc map[int]func(data []byte)
	callbackMapLock *sync.Mutex

	// 实体类型映射和实体ID映射
	entityCategoryMap map[string][]*Entity // key: 设备类型(Category)，value: []*Entity
	entityIdMap       map[string]*Entity   // key: 实体ID(EntityID)，value: *Entity

	// 新增：区域ID映射
	entityAreaMap map[string][]*Entity // key: 区域ID(AreaID)，value: []*Entity

	deviceState   map[string][]*Entity //设备名称：所有实体
	deviceIdState map[string][]*Entity //设备id：所有实体
	deviceMap     map[string]*device   //设备名称：所有实体

	areas    []string
	areaName map[string]string

	xinguang    map[string]string //key:deviceId,value:id
	notifyPhone []string

	speakersXiaomiHome []string //device_id

	service map[string]interface{}
}

func GetDevice() map[string]*device {
	gHub.lock.RLock()
	data := gHub.deviceMap
	gHub.lock.RUnlock()

	return data
}

func GetService() map[string]interface{} {
	gHub.lock.RLock()
	data := gHub.service
	gHub.lock.RUnlock()

	return data
}

func GetDeviceFirstState() map[string][]*Entity {
	gHub.lock.RLock()
	data := gHub.deviceState
	gHub.lock.RUnlock()
	return data
}

func GetEntitiesById() map[string][]*Entity {
	gHub.lock.RLock()
	data := gHub.deviceIdState
	gHub.lock.RUnlock()
	return data
}

func GetSpeakersXiaomiHome() []string {
	gHub.lock.RLock()
	data := gHub.speakersXiaomiHome
	gHub.lock.RUnlock()
	return data
}

func GetNotifyPhone() []string {
	gHub.lock.RLock()
	data := gHub.notifyPhone
	gHub.lock.RUnlock()
	return data
}

func GetXinGuang(deviceId string) string {
	gHub.lock.RLock()
	data := gHub.xinguang[deviceId]
	gHub.lock.RUnlock()
	return data
}

func GetAreaName(areaId string) string {
	gHub.lock.RLock()
	data := gHub.areaName[areaId]
	gHub.lock.RUnlock()
	return data
}

func GetAreas() []string {
	gHub.lock.RLock()
	data := gHub.areas
	gHub.lock.RUnlock()
	return data
}

func GetEntityAreaMap() map[string][]*Entity {
	gHub.lock.RLock()
	data := gHub.entityAreaMap
	gHub.lock.RUnlock()
	return data
}

func GetEntityIdMap() map[string]*Entity {
	gHub.lock.RLock()
	data := gHub.entityIdMap
	gHub.lock.RUnlock()
	return data
}

func GetEntityCategoryMap() map[string][]*Entity {
	gHub.lock.RLock()
	data := gHub.entityCategoryMap
	gHub.lock.RUnlock()
	return data
}

var gHub *hub

func newHub() {
	gHub = &hub{
		lock:               new(sync.RWMutex),
		idLock:             new(sync.Mutex),
		entityCategoryMap:  make(map[string][]*Entity),
		entityIdMap:        make(map[string]*Entity),
		entityAreaMap:      make(map[string][]*Entity),
		deviceMap:          make(map[string]*device),
		areaName:           make(map[string]string),
		areas:              make([]string, 0, 2),
		xinguang:           make(map[string]string),
		callbackMapFunc:    make(map[int]func(data []byte)),
		callbackMapLock:    new(sync.Mutex),
		deviceState:        make(map[string][]*Entity),
		deviceIdState:      make(map[string][]*Entity),
		speakersXiaomiHome: make([]string, 0),
	}
}

func sendWsRequest(req map[string]interface{}, callback func([]byte)) {
	id := GetServiceIncreaseId()
	req["id"] = id
	gHub.callbackMapLock.Lock()
	gHub.callbackMapFunc[id] = callback
	gHub.callbackMapLock.Unlock()
	gHub.writeJson(req)
}

func (h *hub) writeJson(data interface{}) {
	h.lock.RLock() // 添加读锁
	defer h.lock.RUnlock()
	if h.conn == nil {
		ava.Errorf("websocket connection is nil, cannot write")
		return
	}
	h.conn.WriteJSON(data)
}

var channelMessage = make(chan []byte, 1024)

// 添加等待初始化完成的函数
func WaitForInit() {
	<-initDone
}

// 检查初始化是否完成
func checkInitComplete() {
	initMutex.Lock()
	defer initMutex.Unlock()

	// 检查所有必要的数据是否都已加载
	if initState.areasLoaded && initState.devicesLoaded &&
		initState.entitiesLoaded && initState.servicesLoaded &&
		initState.statesLoaded && initState.wsConnected {
		// 防止重复关闭
		select {
		case <-initDone:
			return
		default:
			close(initDone)
			ava.Debugf("Initialization completed successfully")
		}
	}
}

func init() {
	newHub()

	go callback()
	go websocketHaWithInit(defaultToken, getHostAndPath())

	// 设置一个超时机制，防止无限等待
	go func() {
		time.Sleep(time.Second * 20) // 20秒超时
		select {
		case <-initDone:
			return // 已经完成
		default:
			ava.Warnf("Initialization timeout, proceeding anyway")
			close(initDone)
		}
	}()
}

// 初始化数据，设备信息-区域信息
// 版本信息，获取版本号，替换缓存数据
var areaMap = make(map[string]string, 10) // area_id -> name
var entityShortMap = make(map[string]*Entity, 10)
var stateMap = make(map[string]*State, 10)

func callback() {
	for msg := range channelMessage {
		id := jsoniter.Get(msg, "id").ToInt64()
		tpe := jsoniter.Get(msg, "type").ToString()
		success := jsoniter.Get(msg, "success").ToBool()
		if tpe == "result" && !success {
			ava.Errorf("some error occurred |data=%s", string(msg))
			continue
		}

		if tpe == "result" && success {
			gHub.callbackMapLock.Lock()
			cb, ok := gHub.callbackMapFunc[int(id)]
			if ok {
				delete(gHub.callbackMapFunc, int(id))
			}
			gHub.callbackMapLock.Unlock()
			if ok {
				cb(msg)
				continue
			}
			ava.Debugf("No callback registered for id=%d", id)
		}

		if tpe == "event" {
			ava.Debugf("--------%s", string(msg))
			var eventData StateChangedSimple
			err := x.Unmarshal(msg, &eventData)
			if err != nil {
				ava.Error(err)
				continue
			}

			handleData(&eventData, msg)
		}
	}
}

func writeToFile(filename string, data interface{}) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(x.MustMarshal(data))
}

func writeBytesToFile(filename string, data []byte) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(data)
}

// 添加数据处理函数映射表
var (
	dataHandlers = make(map[int]func(*StateChangedSimple, []byte)) // key: handler ID, value: data handling function
	handlerID    = 1                                               // 自增的 handler ID
)

type EventHandler func(*StateChangedSimple, []byte)

// 注册数据处理函数
func RegisterDataHandler(handler func(*StateChangedSimple, []byte)) int {
	gHub.lock.Lock()
	defer gHub.lock.Unlock()

	// 分配一个新的 handler ID 并保存到映射表中
	currentID := handlerID
	dataHandlers[currentID] = handler
	handlerID++

	return currentID
}

// 删除数据处理函数
func UnregisterDataHandler(id int) {
	gHub.lock.Lock()
	defer gHub.lock.Unlock()

	delete(dataHandlers, id)
}

// 修改 handleData 函数以支持多处理器调用
func handleData(event *StateChangedSimple, data []byte) {
	gHub.lock.RLock()
	defer gHub.lock.RUnlock()

	// 遍历所有注册的处理器并执行
	for _, handler := range dataHandlers {
		//todo 初始化过滤一些不用的事件
		handler(event, data)
	}
}

func CallService() {
	callAreaList()
	callDeviceList()
	callEntityList()
	callServices()
	callStates()
}

// websocketHaWithInit wraps websocketHa, on first connect success, calls data fetchers
func websocketHaWithInit(token, url string) {
	firstConnect := true
	websocketHaWithCallback(token, url, func() {
		if firstConnect {
			CallService()
			firstConnect = false
		}
	})
}

// websocketHaWithCallback wraps websocketHa, calls onConnect after handshake success
func websocketHaWithCallback(token, url string, onConnect func()) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	reconnect := func() (*websocket.Conn, error) {
		u := urlpkg.URL{Scheme: "ws", Host: url, Path: "/api/websocket"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", u.String(), token, err)
			return nil, err
		}
		conn.ReadMessage()
		var authReq = struct {
			Type        string `json:"type"`
			AccessToken string `json:"access_token"`
		}{Type: "auth", AccessToken: token}
		err = conn.WriteJSON(&authReq)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		type Result struct {
			Type    string `json:"type"`
			Success bool   `json:"success"`
		}
		var result Result
		err = x.Unmarshal(message, &result)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		if result.Type != "auth_ok" {
			ava.Errorf("host=%s |token=%s", url, token)
			return nil, fmt.Errorf("authentication failed")
		}
		ava.Debugf("handshake success |host=%s |token=%s", url, token)

		var state = struct {
			Id        int    `json:"id"`
			Type      string `json:"type"`
			EventType string `json:"event_type"`
		}{Id: time.Now().Nanosecond(), Type: "subscribe_events", EventType: "state_changed"}
		err = conn.WriteJSON(&state)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		_, stateMessage, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}

		var stateResult Result
		err = x.Unmarshal(stateMessage, &stateResult)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		if !stateResult.Success {
			ava.Errorf("host=%s |token=%s |stateResult=%v", url, token, stateResult)
			return nil, fmt.Errorf("State subscription failed")
		}

		gHub.lock.Lock() // 添加写锁
		gHub.conn = nil
		gHub.conn = conn // 受保护赋值
		gHub.lock.Unlock()

		// 标记 WebSocket 连接成功
		initMutex.Lock()
		initState.wsConnected = true
		initMutex.Unlock()
		checkInitComplete()

		if onConnect != nil {
			onConnect()
		}
		ava.Debug("handshake success")
		return conn, nil
	}

	var backoffTime = time.Second * 10 // 初始重连时间
	for {
		conn, err := reconnect()
		if err != nil {
			ava.Errorf("initial connection failed, retrying in %v |err=%v |home=%s", backoffTime, err, url)
			time.Sleep(backoffTime)
			backoffTime *= 2
			continue
		}
		backoffTime = time.Second * 10 // 连上之后，重置重连时间
		done := make(chan struct{})
		go func() {
			defer func() { close(done); conn.Close() }()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					ava.Errorf("home=%s |err=%v", url, err)
					return
				}

				//ava.Debugf("handshake |state_changed |message=%s", string(message))

				channelMessage <- message
			}
		}()
		select {
		case <-done:
			break
		case <-interrupt:
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func CallServiceWs(data interface{}) {
	now := time.Now()

	if data == nil {
		data = struct{}{}
	}

	gHub.writeJson(&data)
	ava.Debugf("callServiceHttpWs |home=%s |data=%s |latncy=%f", defaultURL, x.MustMarshal2String(data), time.Since(now).Seconds())
}

var idDefault = 1000000000

func GetIncreaseId() int {
	gHub.idLock.Lock()
	idDefault++
	id := idDefault
	gHub.idLock.Unlock()
	return id
}

type chatMessage struct {
	MessageType int32  `protobuf:"varint,1,opt,name=message_type,json=messageType,proto3" json:"message_type,omitempty"`
	Content     string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	CreatedAt   int64  `protobuf:"varint,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UserName    string `protobuf:"bytes,4,opt,name=user_name,json=userName,proto3" json:"user_name,omitempty"`
}

// 发送聊天记录到云端
func sendChatMessage(message *chatMessage) {
	//todo
}
