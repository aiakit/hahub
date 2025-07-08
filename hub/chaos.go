package hub

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"os/signal"
	"time"

	"github.com/aiakit/ava"
	"github.com/gorilla/websocket"

	"strings"

	jsoniter "github.com/json-iterator/go"
)

var (
	defaultAreaInfo *areaList
)

var defaultToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIwMTZkZmM2ZDEwMTg0ZTRjYjJkMDBkMDUzMTYwNmFmZSIsImlhdCI6MTc1MTM0ODQyNCwiZXhwIjoyMDY2NzA4NDI0fQ.2W03gIpG2mJaYUPuT0OGST8zFN1paJ40ltFE9WG52Yg"

const defaultURL = "homeassistant.local:8123"

type hub struct {
	conn *websocket.Conn
}

var gHub *hub

func newHub() {
	gHub = &hub{}
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
	go websocketHaWithInit(defaultToken, defaultURL)
}

// 初始化数据，设备信息-区域信息
// 版本信息，获取版本号，替换缓存数据
func callback() {
	var areaMap = make(map[string]string) // area_id -> name

	for msg := range channelMessage {
		id := jsoniter.Get(msg, "id").ToInt64()
		tpe := jsoniter.Get(msg, "type").ToString()
		success := jsoniter.Get(msg, "success").ToBool()

		if tpe == "result" && success == true {
			switch id {
			case getAreaInfoId: // 获取区域数据
				var data areaList
				if err := Unmarshal(msg, &data); err == nil {
					for _, a := range data.Result {
						areaMap[a.AreaId] = a.Name
					}
					writeToFile("area.json", msg)
				}
			case getDeviceListId: // 获取设备数据
				var data deviceList
				if err := Unmarshal(msg, &data); err == nil {
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
					}
					data.Result = filtered
					ava.Debugf("total device=%d", len(filtered))
					writeToFile("device.json", MustMarshal(data))
				}
			case getEntityListId: // 获取实体数据
				var data entityList
				if err := Unmarshal(msg, &data); err == nil {
					var filtered []*entity
					for _, e := range data.Result {
						if strings.Contains(e.OriginalName, "厂家设置") || strings.Contains(e.OriginalName, "厂商") || strings.Contains(e.OriginalName, "恢复出厂设置") {
							continue
						}
						filtered = append(filtered, e)
					}
					data.Result = filtered
					ava.Debugf("total entity=%d", len(filtered))
					writeToFile("entity.json", MustMarshal(data))
				}
			default:
				ava.Debugf("no id function |data=%v", BytesToString(msg))
			}
		}
		if tpe == "event" {
			handleData(msg)
		}
	}
}

func writeToFile(filename string, data []byte) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(data)
}

func handleData(data []byte) {

}

// websocketHaWithInit wraps websocketHa, on first connect success, calls data fetchers
func websocketHaWithInit(token, url string) {
	firstConnect := true
	websocketHaWithCallback(token, url, func() {
		if firstConnect {
			callAreaList()
			callDeviceList()
			callEntityList()
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
		ava.Debugf("state_changed |message=%s", string(stateMessage))
		var stateResult Result
		err = Unmarshal(stateMessage, &stateResult)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", url, token, err)
			return nil, err
		}
		if !stateResult.Success {
			ava.Errorf("host=%s |token=%s |stateResult=%v", url, token, stateResult)
			return nil, fmt.Errorf("state subscription failed")
		}
		gHub.conn = nil
		gHub.conn = conn
		if onConnect != nil {
			onConnect()
		}
		return conn, nil
	}

	var backoffTime = time.Second * 10 // 初始重连时间
	for {
		conn, err := reconnect()
		if err != nil {
			ava.Errorf("initial connection failed, retrying in %v |err=%v |home=%s", backoffTime, err, defaultURL)
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
					ava.Errorf("home=%s |err=%v", defaultURL, err)
					return
				}
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
	ava.Debugf("callServiceHttpWs |home=%s |data=%s", defaultURL, MustMarshal2String(data))

	if data == nil {
		data = struct{}{}
	}

	gHub.writeJson(&data)
}
