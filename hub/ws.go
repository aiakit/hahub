package hub

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/aiakit/ava"
	"github.com/gorilla/websocket"
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

func (h *hub) getConn() *websocket.Conn {
	return h.conn
}

func (h *hub) addConn(conn *websocket.Conn) {
	h.conn = conn
}

func (h *hub) writeJson(data interface{}) {
	h.conn.WriteJSON(data)
}

// todo 根据最终情况要修改
func init() {
	newHub()
	go websocketHa(defaultToken, defaultURL)
}

var channelMessage = make(chan []byte, 1024)

// curl -H "Upgrade: websocket" -H "Connection: Upgrade"  ws://192.168.1.71:8123/api/websocket
func websocketHa(accessToken, host string) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	reconnect := func() (*websocket.Conn, error) {

		u := url.URL{Scheme: "ws", Host: host, Path: "/api/websocket"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", u.String(), accessToken, err)
			return nil, err
		}

		//过滤掉要求
		conn.ReadMessage()

		// 鉴权
		var authReq = struct {
			Type        string `json:"type"`
			AccessToken string `json:"access_token"`
		}{Type: "auth", AccessToken: accessToken}

		err = conn.WriteJSON(&authReq)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		type Result struct {
			Type    string `json:"type"`
			Success bool   `json:"success"`
		}
		var result Result

		err = Unmarshal(message, &result)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		if result.Type != "auth_ok" {
			ava.Errorf("host=%s |token=%s", host, accessToken)
			return nil, fmt.Errorf("authentication failed")
		}

		ava.Debugf("handshake success |host=%s |token=%s", host, accessToken)

		// 监听状态变化
		var state = struct {
			Id        int    `json:"id"`
			Type      string `json:"type"`
			EventType string `json:"event_type"`
		}{Id: time.Now().Nanosecond(), Type: "subscribe_events", EventType: "state_changed"}

		err = conn.WriteJSON(&state)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		_, stateMessage, err := conn.ReadMessage()
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		ava.Debugf("state_changed |message=%s", string(stateMessage))

		var stateResult Result

		err = Unmarshal(stateMessage, &stateResult)
		if err != nil {
			ava.Errorf("host=%s |token=%s |err=%v", host, accessToken, err)
			return nil, err
		}

		if !stateResult.Success {
			ava.Errorf("host=%s |token=%s |stateResult=%v", host, accessToken, stateResult)
			return nil, fmt.Errorf("state subscription failed")
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

				c := ava.Background()
				c.Debug("----------设备状态变更---------", BytesToString(message))
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
