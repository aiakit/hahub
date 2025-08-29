package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

func RunDevice(message, aiMessage, deviceId string) string {

	f := func(message, aiMessage, deviceId string) string {
		//让ai把意图拆分
		//根据拆分执行不同设备的不同动作

		device := data.GetDevice()
		if len(device) == 0 {
			ava.Debug("没有设备")
			return "没有找到设备"
		}

		var areaName string
		if deviceId != "" {
			dd, ok := data.GetDevice()[deviceId]
			if ok {
				areaName = data.SpiltAreaName(dd.AreaName)
			}
		}

		var deviceNames = make([]string, 0, 10)
		var devicesName = make([]string, 0, 10)
		//过滤区域
		var area = data.GetAreas()
		var areaNames = make(map[string]bool, 2)
		for _, v := range area {
			nn := data.SpiltAreaName(data.GetAreaName(v))
			if strings.Contains(message, nn) {
				areaNames[nn] = true
			}
		}

		var deviceKeyMap = make(map[string]*data.Device)
		var deviceKeyMapTwo = make(map[string]*data.Device)

		//先找到设备名称
		for k, v := range device {
			//区域过滤
			if len(areaNames) > 0 {
				var isExist bool
				for k := range areaNames {
					if strings.Contains(k, data.SpiltAreaName(v.AreaName)) {
						isExist = true
						break
					}
				}

				if !isExist {
					continue
				}
			}

			//类型过滤，过滤最多的两个
			if !strings.Contains(message, "灯") {
				if strings.Contains(v.Model, "light") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "开关") {
				if strings.Contains(v.Model, "switch") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "电视") {
				if strings.Contains(v.Model, "tv") {
					delete(device, k)
					continue
				}
			}

			//删除红外控制
			if strings.Contains(message, "电视") {
				if strings.Contains(v.Model, "ir") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "空调") {
				if strings.Contains(v.Model, "airc") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "插座") {
				if strings.Contains(v.Model, "plug") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "帘") {
				if strings.Contains(v.Model, "curtain") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "音箱") {
				if strings.Contains(v.Model, "wifispeaker") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "存在传感器") {
				if strings.Contains(v.Model, "sensor_occupy") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "存在传感器") {
				if strings.Contains(v.Model, "sensor_occupy") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "人体传感器") {
				if strings.Contains(v.Model, "motion") {
					delete(device, k)
					continue
				}
			}

			if !(strings.Contains(message, "水") && strings.Contains(message, "传感器")) {
				if strings.Contains(v.Model, "flood") {
					delete(device, k)
					continue
				}
			}

			if !(strings.Contains(message, "空开") || strings.Contains(message, "阀") || strings.Contains(message, "闸")) {
				if strings.Contains(v.Model, "valve") {
					delete(device, k)
					continue
				}
			}

			if !strings.Contains(message, "床") {
				if strings.Contains(v.Model, "bed") {
					delete(device, k)
					continue
				}
			}

			//areaName1 := data.SpiltAreaName(v.AreaName)
			//v.Name = areaName1 + v.Name + v.ID
			//名称过滤
			if strings.Contains(message, v.Name) ||
				x.ContainsAllChars(message, v.Name) || x.Similarity(message, v.Name) > 0.8 {
				devicesName = append(devicesName, v.Name)
				deviceKeyMap[v.Name] = v
			}
			deviceNames = append(deviceNames, v.Name)
			deviceKeyMapTwo[v.Name] = v
		}

		if len(devicesName) > 0 {
			deviceNames = nil
			deviceNames = devicesName
		}

		if len(deviceKeyMap) > 0 {
			deviceKeyMapTwo = nil
			deviceKeyMapTwo = deviceKeyMap
		}

		if len(deviceNames) == 0 {
			ava.Debug("没有设备")
			return "没有找到设备"
		}

		var content string
		if areaName != "" {
			content = fmt.Sprintf(`这是我的全部设备名称信息%v，我所在的位置%s，根据我的意图严格返回完整的设备名称。
1.如果我在卧室，只能控制当前卧室位置的设备。
2.如果我不在卧室，只能控制非卧室区域的设备。
返回格式: ["设备名称1","设备名称2"]`, deviceNames, areaName)
		} else {
			content = fmt.Sprintf(`这是我的全部设备信息%v，根据我的意图严格返回完整的设备名称。
返回格式: ["名称1","名称2"]`, x.MustMarshalEscape2String(deviceNames))
		}

		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{Role: "user", Content: content},
			{Role: "user", Content: message},
		})
		if err != nil {
			ava.Error(err)
			return "没有找到对应的设备信息"
		}

		var resultEntities = make(map[string]*data.Entity, 10)
		for k, v := range deviceKeyMapTwo {
			if strings.Contains(result, k) {
				if v1, ok := data.GetEntitiesByDeviceId()[v.ID]; ok {
					for _, v2 := range v1 {
						resultEntities[v2.EntityID] = v2
					}
				}
			}
		}

		if len(resultEntities) == 0 {
			return "没有找到对应的设备信息"
		}

		//找到实体，过滤实体
		resultFiter, _ := getFilterEntities(message, resultEntities)
		if len(resultEntities) == 0 {
			return "没有找到对应设备实体"
		}

		var sendDeviceEntity = make([]*runDeviceEntitity, 0, 10)
		var sendCommandData = make([]*runDeviceCommand, 0, 10)
		for _, v := range resultFiter {
			prefix := strings.Split(v.EntityID, ".")
			if len(prefix) == 0 {
				continue
			}
			domain := prefix[0]
			command, ok := data.GetService()[domain]
			if !ok {
				continue
			}
			sendDeviceEntity = append(sendDeviceEntity, &runDeviceEntitity{
				Domain:     domain,
				EntityId:   v.EntityID,
				EntityName: v.DeviceName,
			})

			sendCommandData = append(sendCommandData, &runDeviceCommand{
				Domain:  domain,
				Command: command,
			})
		}

		//根据实体前缀找到设备指令
		result2, err := chatCompletionInternal([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`这是我的设备信息%s，设备对应的指令信息%s，domain表示对应指令类型，action是domain和具体指令的结合，根据我的意图按照格式进行返回。
1.sub_domain,message是必要字段不能遗漏。
2.fields是具体的要发送控制指令的内容，根据指令信息数据判断是否为空。
3.message是你作为智能家居助理用人性化的语言反馈的内容。
4.target是指令信息数据fields中是否包含target。
5.sub_domain是指令信息domain更下一级的指令划分。
6.必须严格根据设备信息中的domain去寻找指令信息。
返回JSON格式：[{"entity_id":"实体id","target":true,"fields":{"rgb_color":[255, 100, 100]},"sub_domain":"turn_on","message":"xx灯已打开，颜色调整为xx"}]`, x.MustMarshalEscape2String(sendDeviceEntity), x.MustMarshalEscape2String(sendCommandData))},
			{Role: "user", Content: message},
		})
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请重新来一次" + err.Error()
		}

		s := x.FindJSON(result2)
		if len(s) == 0 {
			return "没有发现任何设备"
		}

		var resultData []*runDeviceResultData

		err = x.Unmarshal([]byte(s), &resultData)
		if err != nil {
			ava.Errorf("err=%v |data=%s", err, s)
			return "没有发现任何设备"
		}

		var resultMessage string
		//执行设备操作
		for _, v := range resultData {
			if v.EntityID == "" {
				continue
			}

			if strings.Contains(result2, "on") {
				r, err := data.GetState(v.EntityID)
				if err != nil {
					ava.Error(err)
					continue
				}
				if strings.Contains(strings.ToLower(r.State), "on") {
					resultMessage += data.GetEntityByEntityId()[v.EntityID].DeviceName + "是开着的，"
					continue
				}
			}

			if strings.Contains(result2, "off") {
				r, err := data.GetState(v.EntityID)
				if err != nil {
					ava.Error(err)
					continue
				}
				if strings.Contains(strings.ToLower(r.State), "off") {
					resultMessage += data.GetEntityByEntityId()[v.EntityID].DeviceName + "是关着的，"
					continue
				}
			}

			if strings.Contains(message, "热水器") {
				//判断是不是美的，美的有bug
				v1, ok := data.GetEntityByEntityId()[v.EntityID]
				if !ok {
					continue
				}

				if v.Fields == nil {
					continue
				}

				if v1.Platform == "midea_ac_lan" && v.Fields["temperature"] != nil {
					tmp, ok := v.Fields["temperature"].(float64)
					if !ok {
						continue
					}
					tmp *= 2
					v.Fields["temperature"] = tmp
				}
			}
			v.Fields["entity_id"] = v.EntityID
			domain := strings.Split(v.EntityID, ".")
			if len(domain) == 0 {
				continue
			}
			do := domain[0]

			var isOpenTv bool
			if strings.Contains(result2, "on") && strings.Contains(message, "电视") {
				isOpenTv = turnOnTv(v.EntityID)
			}

			if !isOpenTv {
				err = x.Post(ava.Background(), fmt.Sprintf("%s/api/services/%s", data.GetHassUrl(), do+"/"+v.SubDomain), data.GetToken(), v.Fields, nil)
				if err != nil {
					ava.Error(err)
					continue
				}
			}

			if resultMessage != "" {
				if x.Similarity(resultMessage, v.Message) < 0.5 {
					resultMessage += v.Message
					continue
				} else {
					continue
				}
			}

			resultMessage += v.Messagea + "，"
		}

		if resultMessage == "" {
			return "没有任何操作指令"
		}

		return resultMessage
	}

	return CoreDelay(message, aiMessage, deviceId, f)
}

type runDeviceEntitity struct {
	Domain     string `json:"domain"`
	EntityId   string `json:"entity_id"`
	EntityName string `json:"entity_name"`
}

type runDeviceCommand struct {
	Domain  string      `json:"domain"`
	Command interface{} `json:"command"`
}

type runDeviceResultData struct {
	EntityID  string                 `json:"entity_id"`
	Target    bool                   `json:"target"`
	Fields    map[string]interface{} `json:"fields"`
	SubDomain string                 `json:"sub_domain"`
	Message   string                 `json:"message"`
}

// 打开电视的逻辑,解决ha中一些电视无法打开，只可以关闭的问题
// 1.根据设备id，找到设备
// 2.判断当前设备的状态是否为关闭
// 3.找到ir红外控制设备的名称是否和当前电视名称一致,如果一致，使用红外控制打开电视
func turnOnTv(entityId string) bool {
	d, ok := data.GetEntityByEntityId()[entityId]
	if !ok {
		return false
	}

	ir, ok := data.GetEntityCategoryMap()[data.CategoryIrTV]
	if !ok {
		return false
	}

	for _, v := range ir {
		if strings.Contains(v.DeviceName, d.DeviceName) {
			if !strings.Contains(v.OriginalName, "开机") {
				continue
			}
			//使用红外控制打开电视
			err := x.Post(ava.Background(), fmt.Sprintf("%s/api/services/button/press", data.GetHassUrl()), data.GetToken(), &data.HttpServiceData{
				EntityId: v.EntityID,
			}, nil)
			if err != nil {
				ava.Error(err)
				return false
			}
			ava.Debugf("-------使用红外控制打开电视------")
			return true
		}
	}

	return false
}
