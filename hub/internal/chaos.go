package internal

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aiakit/ava"
	"github.com/gorilla/websocket"

	"strings"

	jsoniter "github.com/json-iterator/go"
)

var defaultToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIwMTZkZmM2ZDEwMTg0ZTRjYjJkMDBkMDUzMTYwNmFmZSIsImlhdCI6MTc1MTM0ODQyNCwiZXhwIjoyMDY2NzA4NDI0fQ.2W03gIpG2mJaYUPuT0OGST8zFN1paJ40ltFE9WG52Yg"

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

	// 实体类型映射和实体ID映射
	entityCategoryMap map[string][]*Entity // key: 设备类型(Category)，value: []*Entity
	entityIdMap       map[string]*Entity   // key: 实体ID(EntityID)，value: *Entity

	// 新增：区域ID映射
	entityAreaMap map[string][]*Entity // key: 区域ID(AreaID)，value: []*Entity
}

func GetEntityAreaMap() map[string][]*Entity {
	return gHub.entityAreaMap
}

func GetEntityIdMap() map[string]*Entity {
	return gHub.entityIdMap
}

func GetEntityCategoryMap() map[string][]*Entity {
	return gHub.entityCategoryMap
}

var gHub *hub

func newHub() {
	gHub = &hub{
		lock:              new(sync.RWMutex),
		idLock:            new(sync.Mutex),
		entityCategoryMap: make(map[string][]*Entity),
		entityIdMap:       make(map[string]*Entity),
		entityAreaMap:     make(map[string][]*Entity),
	}
}

func (h *hub) writeJson(data interface{}) {
	if h.conn == nil {
		ava.Errorf("websocket connection is nil, cannot write")
		return
	}
	h.conn.WriteJSON(data)
}

var channelMessage = make(chan []byte, 1024)

func init() {
	newHub()
	go callback()
	go websocketHaWithInit(defaultToken, getHostAndPath())
	time.Sleep(time.Second * 3)
}

// 初始化数据，设备信息-区域信息
// 版本信息，获取版本号，替换缓存数据
func callback() {
	var areaMap = make(map[string]string, 10) // area_id -> name
	var entityShortMap = make(map[string]*Entity, 10)
	var deviceMap = make(map[string]*device, 10)
	var stateMap = make(map[string]*State, 10)
	for msg := range channelMessage {
		id := jsoniter.Get(msg, "id").ToInt64()
		tpe := jsoniter.Get(msg, "type").ToString()
		success := jsoniter.Get(msg, "success").ToBool()

		if tpe == "result" && success == true {
			func() {
				gHub.lock.Lock()
				defer gHub.lock.Unlock()
				switch id {
				case getAreaInfoId: // 获取区域数据
					var data areaList
					err := Unmarshal(msg, &data)
					if err != nil {
						ava.Errorf("Unmarshal areaList error: %v", err)
						break
					}
					for _, a := range data.Result {
						areaMap[a.AreaId] = a.Name
					}
					data.Total = len(data.Result)
					writeToFile("area.json", &data)
				case getDeviceListId: // 获取设备数据
					var data deviceList
					err := Unmarshal(msg, &data)
					if err != nil {
						ava.Errorf("Unmarshal deviceList error: %v", err)
						break
					}
					var filtered []*device
					for _, d := range data.Result {
						// area_id 为空则忽略
						if d.AreaID == "" {
							continue
						}
						if name, ok := areaMap[d.AreaID]; ok {
							d.AreaName = name
						}
						filtered = append(filtered, d)
						deviceMap[d.ID] = d
					}
					data.Result = filtered
					data.Total = len(filtered)
					ava.Debugf("total device=%d", len(filtered))
					writeToFile("device.json", data)
				case getEntityListId: // 获取实体数据
					var data EntityList
					//var dataTest EntityListTest
					err := Unmarshal(msg, &data)
					if err != nil {
						ava.Errorf("Unmarshal EntityList error: %v", err)
						break
					}
					//err = Unmarshal(msg, &dataTest)
					//if err != nil {
					//	ava.Errorf("Unmarshal EntityList error: %v", err)
					//	break
					//}
					var filtered []*Entity
					for _, e := range data.Result {
						if strings.Contains(e.OriginalName, "厂家设置") || strings.Contains(e.OriginalName, "厂商") || strings.Contains(e.OriginalName, "恢复出厂设置") {
							continue
						}
						if e.Platform == "hacs" || e.Platform == "hassio" || e.Platform == "sun" || e.Platform == "backup" || e.Platform == "person" ||
							e.Platform == "shopping_list" || e.Platform == "google_translate" || e.Platform == "met" {
							continue
						}

						filtered = append(filtered, e)
					}
					data.Result = filtered
					data.Total = len(filtered)
					ava.Debugf("total Entity=%d", len(filtered))
					writeToFile("entity.json", &data)
					//writeToFile("entity_test.json", &dataTest)

					// 写入短实体
					shortEntities := FilterEntities(filtered, deviceMap)
					shortData := EntityList{ID: data.ID, Type: data.Type, Success: data.Success, Result: shortEntities}
					shortData.Total = len(shortEntities)
					for _, d := range shortEntities {
						entityShortMap[d.EntityID] = d
					}
					writeToFile("entity_short.json", &shortData)

					// 填充entityCategoryMap、entityIdMap和entityAreaMap
					for _, e := range shortEntities {
						if e.Category != "" {
							gHub.entityCategoryMap[e.Category] = append(gHub.entityCategoryMap[e.Category], e)
						}
						if e.EntityID != "" {
							gHub.entityIdMap[e.EntityID] = e
						}
						if e.AreaID != "" {
							gHub.entityAreaMap[e.AreaID] = append(gHub.entityAreaMap[e.AreaID], e)
						}
					}
				case getServicesId: // 获取服务数据
					var data serviceList
					err := Unmarshal(msg, &data)
					if err != nil {
						ava.Errorf("Unmarshal serviceList error: %v", err)
						break
					}
					data.Total = len(data.Result)
					ava.Debugf("total services=%d", len(data.Result))
					writeToFile("services.json", &data)
				case getStatesId: // 获取实体详细信息
					var data stateList
					err := Unmarshal(msg, &data)
					if err != nil {
						ava.Errorf("Unmarshal stateList error: %v", err)
						break
					}

					var filter = make([]*State, 0, 1024)
					for _, v := range data.Result {
						tmp := MustMarshal(v)
						id := Json.Get(tmp, "entity_id").ToString()
						if _, ok := entityShortMap[id]; ok {
							filter = append(filter, v)
							stateMap[id] = v
						}
					}
					data.Result = nil
					data.Result = filter
					data.Total = len(data.Result)
					ava.Debugf("total states=%d", len(data.Result))
					writeToFile("states.json", &data)
				default:
					ava.Debugf("no id function |data=%v", BytesToString(msg))
				}
			}()
		}
		if tpe == "event" {
			handleData(msg)
		}
	}
}

func writeToFile(filename string, data interface{}) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(MustMarshal(data))
}

func handleData(data []byte) {

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
		err = Unmarshal(message, &result)
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
		err = Unmarshal(stateMessage, &stateResult)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		if !stateResult.Success {
			ava.Errorf("host=%s |token=%s |stateResult=%v", url, token, stateResult)
			return nil, fmt.Errorf("State subscription failed")
		}
		gHub.conn = nil
		gHub.conn = conn
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
	ava.Debugf("callServiceHttpWs |home=%s |data=%s |latncy=%f", defaultURL, MustMarshal2String(data), time.Since(now).Seconds())
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
