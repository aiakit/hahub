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
	category string
	id       string
	Name     string `json:"name"`
	State    string `json:"state,omitempty"`
	AreaName string `json:"area_name,omitempty"`
}

// todo 不要传开关所有实体,馨光实体不传
func QueryDevice(message, aiMessage, deviceId string) string {
	var allEntities = data.GetEntityByEntityId()
	resultEntities, _ := getFilterEntities(allEntities)

	//获取所有实体状态,再匹配实体
	states, err := data.GetStates()
	if err != nil {
		ava.Error(err)
		return "没有找到设备数据"
	}

	//状态
	//所有设备
	var entities = make(map[string][]shortDevice, 20) //device_name:实体
	//离线设备
	var offlineDevices = make(map[string]bool, 5)
	//在线设备
	var onlineDevices = make(map[string]bool, 5)

	var deviceNameMap = make(map[string]bool)

	var entitiesMap = make(map[string]data.StateAll, 20)

	for _, state := range states {
		en, ok := resultEntities[state.EntityID]
		if !ok {
			continue
		}

		if strings.HasPrefix(state.EntityID, "automation") || strings.HasPrefix(state.EntityID, "script") || strings.HasPrefix(state.EntityID, "scene") {
			continue
		}
		name := en.DeviceName
		areaName := data.SpiltAreaName(en.AreaName)

		if areaName == "" {
			continue
		}

		if name == "" {
			continue
		}

		// 修改name的构建方式，加入EntityID以避免重复
		name = areaName + "_" + name + "_" + en.DeviceID

		if len(entities[name]) > 3 {
			continue
		}

		if !deviceNameMap[name] {
			if state.State == "unavailable" {
				offlineDevices[name] = true
			} else {
				onlineDevices[name] = true
			}
		}

		entities[name] = append(entities[name], shortDevice{
			id:       en.EntityID,
			Name:     name,
			AreaName: data.SpiltAreaName(en.AreaName),
			category: en.Category,
		})

		state.DeviceName = name
		entitiesMap[state.EntityID] = state
	}

	var countTotal = len(entities)

	var devicess = data.GetDevice()
	//如果是具体某个设备

	var entitiesMapTwo = make(map[string][]shortDevice, 4)
	//直接找设备
	for _, d := range devicess {
		if strings.Contains(message, d.Name) {
			name := d.Name
			areaName := data.SpiltAreaName(d.AreaName)
			name = areaName + "_" + name + "_" + d.ID
			if v, ok := entities[name]; ok {
				entitiesMapTwo[name] = v
			}
		}
	}

	if len(entitiesMapTwo) == 0 {
		for _, d := range devicess {
			var areaName = data.SpiltAreaName(d.AreaName)
			if areaName == "" {
				continue
			}

			// 同样修改这里name的构建方式
			name := areaName + "_" + d.Name + "_" + d.ID

			if _, ok := entities[name]; !ok && !strings.Contains(d.Model, "ir") {
				ava.Debugf("漏掉计算的设备 |name=%s |mode=%s", name, d.Model)
			}

			area := data.GetAreas()
			var areaNames = make([]string, 0, 2)
			for _, v := range area {
				nn := data.SpiltAreaName(data.GetAreaName(v))
				if strings.Contains(message, nn) {
					areaNames = append(areaNames, nn)
				}
			}

			if v, ok := entities[name]; ok {
				for _, v1 := range v {
					if len(areaNames) > 0 {
						for _, aa := range areaNames {
							if !strings.Contains(v1.AreaName, aa) {
								delete(entities, name)
								continue
							}
						}
					}

					if !strings.Contains(message, "灯") {
						if v1.category == data.CategoryLight {
							delete(entities, name)
							continue
						}

						if v1.category == data.CategoryLightGroup {
							delete(entities, name)
							continue
						}
					}

					if !strings.Contains(message, "开关") {
						if v1.category == data.CategorySwitch {
							delete(entities, name)
							continue
						}
					}

					if !strings.Contains(message, "空调") {
						if v1.category == data.CategoryAirConditioner {
							delete(entities, name)
							continue
						}
					}

					//if v1.category == data.CateOther {
					//	delete(entities, name)
					//	continue
					//}

					if v1.category == data.CategoryVirtualEvent {
						delete(entities, name)
						continue
					}
				}
			}
		}
	}

	if len(entitiesMapTwo) > 0 {
		entities = nil
		entities = entitiesMapTwo
	}

	var sendData = make([]string, 0, 20)
	if strings.Contains(aiMessage, "query_offline_number") {
		return "你有" + strconv.Itoa(len(offlineDevices)) + "个设备离线了"
	}

	if strings.Contains(aiMessage, "query_online_number") {
		return "你有" + strconv.Itoa(len(onlineDevices)) + "个设备在线"
	}

	if strings.Contains(aiMessage, "query_all_number") {
		return "你总共有" + strconv.Itoa(countTotal) + "个设备"
	}

	if strings.Contains(aiMessage, "query_offline_state") {
		for k := range entities {
			if _, ok := offlineDevices[k]; ok {
				sendData = append(sendData, k)
			}
		}
	} else if strings.Contains(aiMessage, "query_online_state") {
		for k := range entities {
			if _, ok := onlineDevices[k]; ok {
				sendData = append(sendData, k)
			}
		}
	} else {
		for k := range entities {
			sendData = append(sendData, k)
		}
	}

	if len(sendData) == 0 {
		return "没有对应状态的设备"
	}

	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{Role: "user", Content: fmt.Sprintf(`这是我的全部设备%s，根据我的意图找出对应的设备名称给我。
返回格式: ["设备名称1","设备名称2"]`, x.MustMarshalEscape2String(sendData))},
		{Role: "user", Content: message},
	})
	if err != nil {
		ava.Error(err)
		return "没有找到对应的设备信息"
	}

	var entity = make([]data.StateAll, 0, 2)

	for k, v := range entities {
		for _, v1 := range v {
			if strings.Contains(result, k) {
				dd := entitiesMap[v1.id]
				dd.EntityID = ""
				entity = append(entity, entitiesMap[v1.id])
			}
		}
	}

	//查询所有或者30个设备以上比较难搞，未来优化为内存缓存设备状态，直接计算数据状态返回不经过ai
	if len(entity) > 30 {
		return "你要查寻的设备过多，请到app中查看设备状态"
	}

	result2, err := chatCompletionInternal([]*chat.ChatMessage{
		{Role: "user", Content: fmt.Sprintf(`这是我的设备状态信息%s，根据我的意图用20字以内人性化的语言回答我设备状态是怎么样的。`, x.MustMarshalEscape2String(entity))},
		{Role: "user", Content: message},
	})
	if err != nil {
		ava.Error(err)
		return "服务器开小差了，请重新来一次"
	}

	return result2
}

// 获取过滤的实体,开关是不使用的
// 实体过滤
func getFilterEntities(entities map[string]*data.Entity) (map[string]*data.Entity, map[string]*data.Entity) {
	var result = make(map[string]*data.Entity, 50)
	var resultSkip = make(map[string]*data.Entity, 50)
	var cacheTemp = make(map[string]bool)
	var cacheHumanBody = make(map[string]bool)
	var cacheLx = make(map[string]bool)

	for _, e := range entities {

		if e.Category == data.CategoryXiaomiMiotSpeaker || e.Category == data.CategorySwitchMode ||
			e.Category == data.CategoryXinGuang || e.Category == data.CategorySwitchClickOnce || e.Category == data.CategorySwitchScene {
			continue
		}

		if strings.Contains(e.DeviceMode, ".light.") && !strings.HasPrefix(e.EntityID, "light") {
			continue
		}
		if e.Category == data.CategoryXiaomiHomeSpeaker && !strings.HasPrefix(e.EntityID, "media_player") {
			continue
		}

		if e.Category == data.CategoryTemperatureSensor {
			if data.GetAreaName(e.AreaID) != "" {
				if _, ok := cacheTemp[e.AreaID]; !ok {
					cacheTemp[e.AreaID] = true
					result[e.EntityID] = e
				} else {
					continue
				}
			} else {
				continue
			}
		}

		if e.Category == data.CategoryHumiditySensor {
			if data.GetAreaName(e.AreaID) != "" {
				if _, ok := cacheHumanBody[e.AreaID]; !ok {
					cacheHumanBody[e.AreaID] = true
					result[e.EntityID] = e
				} else {
					continue
				}
			} else {
				continue
			}
		}

		if e.Category == data.CategoryLxSensor {
			if data.GetAreaName(e.AreaID) != "" {
				if _, ok := cacheLx[e.AreaID]; !ok && strings.Contains(e.DeviceName, "存在传感器") {
					cacheLx[e.AreaID] = true
					e.OriginalName = data.SpiltAreaName(e.AreaName) + "光照传感器"
					result[e.EntityID] = e
				} else {
					continue
				}
			} else {
				continue
			}
		}

		if strings.Contains(e.DeviceName, "热水器") && !strings.HasPrefix(e.EntityID, "water_heater") {
			continue
		}

		if strings.Contains(e.DeviceName, "浴霸") && !strings.HasPrefix(e.EntityID, "climate") {
			continue
		}

		if strings.Contains(e.DeviceMode, "airc") && strings.Contains(e.DeviceName, "空调") && !strings.HasPrefix(e.EntityID, "climate") {
			continue
		}

		if strings.Contains(e.DeviceName, "新风") && !strings.HasPrefix(e.EntityID, "switch") {
			continue
		}

		if strings.Contains(e.DeviceName, "新风") && strings.HasPrefix(e.EntityID, "switch") {
			continue
		}

		if strings.Contains(e.DeviceName, "地暖") && !strings.HasPrefix(e.EntityID, "switch") {
			continue
		}

		if strings.Contains(e.DeviceName, "地暖") && strings.HasPrefix(e.EntityID, "switch") {
			continue
		}

		if e.Category == data.CategoryIrTV {
			continue
		}

		if e.Category == data.CategoryHaTV && !strings.HasPrefix(e.EntityID, "media_player") {
			continue
		}

		if e.Category == data.CategoryHaTV {
			e.DeviceName = e.OriginalName
			continue
		}
		result[e.EntityID] = e

		//if e.Category == data.CategoryLightGroup {
		//	continue
		//}

		if e.Category == data.CategorySwitch {
			continue
		}

		if e.Category == data.CateOther {
			continue
		}

		if e.Category == data.CategoryVirtualEvent {
			continue
		}

		resultSkip[e.EntityID] = e

	}
	fmt.Println("----过滤后实体数量1---", len(result))
	fmt.Println("----过滤后实体数量2---", len(resultSkip))

	return result, resultSkip
}
