package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strconv"
	"strings"

	"github.com/aiakit/ava"
)

type shortDevice struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	State    string `json:"state,omitempty"`
	AreaName string `json:"area_name,omitempty"`
}

func QueryDevice(message, aiMessage, deviceId string) string {
	var devices = data.GetDeviceFirstState()

	//获取所有实体状态,再匹配实体
	states, err := data.GetStates()
	if err != nil {
		ava.Error(err)
		return "没有找到设备数据"
	}

	//所有设备
	var entities = make([]shortDevice, 0, 20)
	//离线设备
	var offlineDevices = make([]shortDevice, 0, 5)
	//在线设备
	var onlineDevices = make([]shortDevice, 0, 5)

	var entitiesMap = make(map[string]data.StateAll, 20)

	for _, state := range states {
		name, ok := state.Attributes["friendly_name"].(string)
		if !ok {
			continue
		}

		n := strings.Split(name, " ")
		if len(n) == 0 {
			continue
		}

		deviceName := n[0]
		d, ok := devices[deviceName]
		if !ok {
			continue
		}

		var isExist bool
		for _, v := range d {
			if v.EntityID == state.EntityID {
				isExist = true
				break
			}
		}

		if !isExist {
			continue
		}

		if state.State == "unknown" || state.State == "unavailable" {
			offlineDevices = append(offlineDevices, shortDevice{
				Name: deviceName,
				Id:   state.EntityID,
			})
		} else {
			onlineDevices = append(onlineDevices, shortDevice{
				Name: deviceName,
				Id:   state.EntityID,
			})
		}
		entities = append(entities, shortDevice{
			Name: deviceName,
			Id:   state.EntityID,
		})

		entitiesMap[state.EntityID] = state
	}

	var sendData []shortDevice
	if strings.Contains(aiMessage, "query_offline_number") {
		return "你有" + strconv.Itoa(len(offlineDevices)) + "个设备离线了"
	}

	if strings.Contains(aiMessage, "query_online_number") {
		return "你有" + strconv.Itoa(len(onlineDevices)) + "个设备在线"
	}

	if strings.Contains(aiMessage, "query_total_number") {
		return "你总共有" + strconv.Itoa(len(entities)) + "设备"
	}

	if strings.Contains(aiMessage, "query_offline_state") {
		sendData = offlineDevices
	} else if strings.Contains(aiMessage, "query_online_state") {
		sendData = onlineDevices
	} else {
		sendData = entities
	}

	result, err := chatCompletionHistory([]*chat.ChatMessage{
		{Role: "user", Content: fmt.Sprintf(`这是我的全部设备信息%s，选择对应的设备信息按照格式返回给我，
返回格式: [{"name":"设备名称","id":"设备id"}]`, x.MustMarshalEscape2String(sendData))},
		{Role: "user", Content: message},
	}, deviceId)
	if err != nil {
		ava.Error(err)
		return "没有找到对应的设备信息"
	}

	var entity = make([]data.StateAll, 0, 2)

	for _, v := range entities {
		if strings.Contains(result, v.Id) {
			entity = append(entity, entitiesMap[v.Id])
		}
	}

	result2, err := chatCompletionHistory([]*chat.ChatMessage{
		{Role: "user", Content: fmt.Sprintf(`这是我的设备信息%s，根据我的意图用人性化的语言回答我。`, x.MustMarshalEscape2String(sendData))},
		{Role: "user", Content: message},
	}, deviceId)
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请重新来一次"
	}

	return result2
}
