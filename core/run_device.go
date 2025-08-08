package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

type commandData struct {
	Id     string                 `json:"id"`
	Filed  map[string]interface{} `json:"filed"`
	Action string                 `json:"action"`
}

func RunDevice(message, aiMessage, deviceId string) string {

	f := func(message, aiMessage, deviceId string) string {
		var devices = data.GetDeviceFirstState()

		//获取所有实体状态,再匹配实体
		states, err := data.GetStates()
		if err != nil {
			ava.Error(err)
			return "没有找到设备数据"
		}

		var entities = make([]shortDevice, 0, 20)
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

			entities = append(entities, shortDevice{
				Name:  deviceName,
				Id:    state.EntityID,
				State: state.State,
			})

			entitiesMap[state.EntityID] = state
		}

		result, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`这是我的全部设备信息%s，选择对应的设备信息按照格式返回给我，
返回格式: [{"name":"设备名称","id":"设备id"}]`, x.MustMarshalEscape2String(entities))},
			{Role: "user", Content: message},
		}, deviceId)
		if err != nil {
			ava.Error(err)
			return "没有找到对应的设备信息"
		}

		var entity = make([]data.StateAll, 0, 2)
		var isServiceExist = make(map[string]bool)
		var command = make(map[string]interface{})

		for _, v := range entities {
			if strings.Contains(result, v.Id) {
				entity = append(entity, entitiesMap[v.Id])

				var sp = strings.Split(v.Id, ".")
				if len(sp) == 0 {
					continue
				}
				cc := sp[0]
				if _, ok := isServiceExist[cc]; ok {
					continue
				}

				v1, ok := data.GetService()[cc]
				if !ok {
					continue
				}
				command[cc] = v1
			}
		}

		result2, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`这是我的设备信息%s，设备对应的指令信息%s，根据我的意图选择并组装操作指令fields按照格式进行返回。
返回格式例子：[{"id":"","fields":{"rgb_color":"[255, 100, 100]}","action":"light/turn_on"}]；id,fileds,action是必要字段不能遗漏。`, x.MustMarshalEscape2String(entity), x.MustMarshalEscape2String(command))},
			{Role: "user", Content: message},
		}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请重新来一次" + err.Error()
		}

		var comm []commandData

		js := x.FindJSON(result2)

		err = x.Unmarshal([]byte(js), &comm)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请重来一次" + err.Error()
		}

		//执行设备操作
		var isExistError bool
		for _, v := range comm {
			v.Filed["entity_id"] = v.Id
			err = x.Post(ava.Background(), fmt.Sprintf("%s/api/services/%s", data.GetHassUrl(), v.Action), data.GetToken(), v.Filed, nil)
			if err != nil {
				ava.Error(err)
				isExistError = true
			}
		}

		if isExistError {
			return "控制设备失败了，请检查设备"
		}

		return "已经执行了"
	}
	return CoreDelay(message, aiMessage, deviceId, f)
}
