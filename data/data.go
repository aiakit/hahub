package data

import (
	"hahub/x"
	"math"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

//向远程获取方案，将各种方案缓存到本地，本地执行

var idServiceDefault = math.MaxInt64 - 100000000

func GetServiceIncreaseId() int {
	gHub.idLock.Lock()
	idServiceDefault++
	id := idServiceDefault
	gHub.idLock.Unlock()
	return id
}

type areaInfo struct {
	AreaId string `json:"area_id"`
	Name   string `json:"name"`
}

type areaList struct {
	Id      int         `json:"id"`
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Total   int         `json:"total"`
	Result  []*areaInfo `json:"result"`
}

// todo 小米的ha插件有 bug,当删除房间之后，websocket返回的数据还是有删除的房间，后期检查
func callAreaList() {
	var to = map[string]interface{}{
		"type": "config/area_registry/list",
	}

	sendWsRequest(to, func(msg []byte) {
		var data areaList
		err := x.Unmarshal(msg, &data)
		if err != nil {
			ava.Errorf("Unmarshal areaList error: %v", err)
			return
		}
		var prefixMap = make(map[string]int)
		for _, a := range data.Result {
			s := strings.Split(a.Name, " ")
			if len(s) == 0 {
				continue
			}

			if len(s) > 1 {
				prefixMap[s[0]]++
			}
		}

		var prefix string
		var flag int
		for k, v := range prefixMap {
			if flag > v {
				prefix = k
				break
			}
			flag = v
		}

		for _, a := range data.Result {
			s := strings.Split(a.Name, " ")

			if len(s) == 1 && s[0] != prefix {
				continue
			}

			gHub.areas = append(gHub.areas, a.AreaId)
			gHub.areaName[a.AreaId] = a.Name
		}

		data.Total = len(data.Result)
		writeToFile("area.json", &data)

		// 标记区域数据已加载
		initMutex.Lock()
		initState.areasLoaded = true
		initMutex.Unlock()
		checkInitComplete()
	})
}

// 获取区域数据
func getAreaData(message string) (*areaList, error) {
	var data areaList
	err := x.Unmarshal([]byte(message), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// 获取全量设备数据
type deviceList struct {
	ID      int64     `json:"id"`
	Type    string    `json:"type"`
	Success bool      `json:"success"`
	Total   int       `json:"total"`
	Result  []*Device `json:"result"`
}

type Device struct {
	AreaID     string  `json:"area_id"`     //区域id
	AreaName   string  `json:"area_name"`   //区域名称
	CreatedAt  float64 `json:"created_at"`  //创建时间
	ModifiedAt float64 `json:"modified_at"` //修改时间
	ID         string  `json:"id"`          //设备id
	Model      string  `json:"model"`       //使用的模型
	Name       string  `json:"name"`        //用户命名的产品名称
	SwVersion  string  `json:"sw_version"`  //固件版本

	NameByUser string `json:"name_by_user"` //修改设备名称后，ha会用这个字段表示修改后的名称
}

func callDeviceList() {
	var to = map[string]interface{}{
		"type": "config/device_registry/list",
	}

	sendWsRequest(to, func(msg []byte) {
		var data deviceList
		err := x.Unmarshal(msg, &data)
		if err != nil {
			ava.Errorf("Unmarshal deviceList error: %v", err)
			return
		}
		var filtered []*Device
		gHub.lock.RLock()
		for _, d := range data.Result {
			if d.NameByUser != "" {
				d.Name = d.NameByUser
			}
			if d.AreaID == "" {
				continue
			}
			if name, ok := gHub.areaName[d.AreaID]; ok {
				d.AreaName = name
			}
			filtered = append(filtered, d)
			gHub.deviceMap[d.ID] = d
		}
		gHub.lock.RUnlock()

		data.Result = filtered
		data.Total = len(filtered)
		ava.Debugf("total Device=%d", len(filtered))
		writeToFile("Device.json", data)
		initMutex.Lock()
		initState.devicesLoaded = true
		initMutex.Unlock()
		checkInitComplete()
	})
}

// 获取全量实体数据
type EntityList struct {
	ID      int64     `json:"id"`
	Type    string    `json:"type"`
	Success bool      `json:"success"`
	Total   int       `json:"total"`
	Result  []*Entity `json:"result"`
}

type EntityListTest struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Total   int    `json:"total"`
	Result  any    `json:"result"`
}

type Entity struct {
	DeviceID     string `json:"device_id"`     //设备id
	EntityID     string `json:"entity_id"`     //实体id
	ID           string `json:"id"`            //真实实体id，数据id
	OriginalName string `json:"original_name"` //实体名称
	Platform     string `json:"platform"`      //产自什么平台
	UniqueID     string `json:"unique_id"`     //唯一id

	Category    string `json:"category"`     //设备类型
	SubCategory string `json:"sub_category"` //设备二级类型
	AreaID      string `json:"area_id"`      //区域id
	AreaName    string `json:"area_name"`    //区域名称
	DeviceName  string `json:"device_name"`  //设备名称（从设备数据获取）
	DeviceMode  string `json:"device_mode"`  //设备类型
	Name        string `json:"name"`         //修改名称之后，ha会用这个字段表示名称,ha修改OriginalName
}

var callbackEntityMap = make([]func(e *Entity), 0)

func RegisterEntityCallback(callback func(e *Entity)) {
	callbackEntityMap = append(callbackEntityMap, callback)
}

func callbackEntity(entity *Entity) {
	for _, callback := range callbackEntityMap {
		callback(entity)
	}
}

func callEntityList() {
	var to = map[string]interface{}{
		"type": "config/entity_registry/list",
	}

	sendWsRequest(to, func(msg []byte) {
		var data EntityList
		var dataALL EntityListTest
		err := x.Unmarshal(msg, &data)
		if err != nil {
			ava.Errorf("Unmarshal EntityList error: %v", err)
			return
		}
		err = x.Unmarshal(msg, &dataALL)
		if err != nil {
			ava.Errorf("Unmarshal EntityListTest error: %v", err)
			return
		}
		var filtered []*Entity
		for _, e := range data.Result {

			callbackEntity(e)

			if strings.Contains(e.OriginalName, "接近远离") && strings.HasPrefix(e.EntityID, "binary_sensor.") {
				filtered = append(filtered, e)
				continue
			}

			if strings.Contains(e.OriginalName, "厂家设置") || strings.Contains(e.OriginalName, "厂商") || strings.Contains(e.OriginalName, "恢复出厂设置") {
				continue
			}

			if strings.Contains(e.OriginalName, "背光") || strings.Contains(e.OriginalName, "指示灯") || strings.Contains(e.OriginalName, "倒计时") {
				continue
			}

			if strings.Contains(e.OriginalName, "拓展") || strings.Contains(e.OriginalName, "双击") || strings.Contains(e.OriginalName, "长按") {
				continue
			}

			if strings.Contains(e.OriginalName, "遥控") || strings.Contains(e.OriginalName, "超时") {
				continue
			}

			if strings.Contains(e.OriginalName, "最大功率开关") || strings.Contains(e.OriginalName, "提醒") || strings.Contains(e.OriginalName, "充电保护") {
				continue
			}

			if strings.Contains(e.OriginalName, "开关状态切换") {
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
		writeToFile("entity_test.json", &dataALL)

		shortEntities := FilterEntities(filtered, GetDevice())
		shortData := EntityList{ID: data.ID, Type: data.Type, Success: data.Success, Result: shortEntities}
		shortData.Total = len(shortEntities)

		writeToFile("entity_short.json", &shortData)

		gHub.lock.Lock()
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
		gHub.lock.Unlock()

		initMutex.Lock()
		initState.entitiesLoaded = true
		initMutex.Unlock()
		checkInitComplete()
	})
}

// 获取服务数据
type serviceList struct {
	ID      int64                  `json:"id"`
	Type    string                 `json:"type"`
	Success bool                   `json:"success"`
	Total   int                    `json:"total"`
	Result  map[string]interface{} `json:"result"`
}

func callServices() {
	var to = map[string]interface{}{
		"type": "get_services",
	}

	sendWsRequest(to, func(msg []byte) {
		var data serviceList
		err := x.Unmarshal(msg, &data)
		if err != nil {
			ava.Errorf("Unmarshal serviceList error: %v", err)
			return
		}
		data.Total = len(data.Result)
		ava.Debugf("total services=%d", len(data.Result))
		writeToFile("services.json", &data)

		for k, v := range data.Result {
			if k == "notify" {
				v1, ok := v.(map[string]interface{})
				if ok {
					for key := range v1 {
						if strings.HasPrefix(key, "mobile_") {
							gHub.notifyPhone = append(gHub.notifyPhone, key)
						}
					}
				}

			}
		}
		initMutex.Lock()
		initState.servicesLoaded = true
		gHub.service = data.Result
		initMutex.Unlock()
		checkInitComplete()
	})
}

// 获取实体详细信息
// 结构参考Home Assistant get_states返回
// https://developers.home-assistant.io/docs/api/websocket/#get_states

type stateList struct {
	ID      int64    `json:"id"`
	Type    string   `json:"type"`
	Success bool     `json:"success"`
	Total   int      `json:"total"`
	Result  []*State `json:"result"`
	//Result []interface{} `json:"result"`
}

type State struct {
	EntityID   string `json:"entity_id"`
	State      string `json:"State"`
	Attributes struct {
		Mode                string   `json:"mode"`
		Current             int      `json:"current"`
		FriendlyName        string   `json:"friendly_name"`
		SupportedColorModes []string `json:"supported_color_modes"`
		ID                  string   `json:"id"`
		LastTriggered       any      `json:"last_triggered"`
	} `json:"attributes"`
	LastChanged  time.Time `json:"last_changed"`
	LastReported time.Time `json:"last_reported"`
	LastUpdated  time.Time `json:"last_updated"`
}

func callStates() {
	var to = map[string]interface{}{
		"type": "get_states",
	}

	sendWsRequest(to, func(msg []byte) {
		var data stateList
		err := x.Unmarshal(msg, &data)
		if err != nil {
			ava.Errorf("Unmarshal stateList error: %v", err)
			return
		}
		var filter = make([]*State, 0, 1024)
		gHub.lock.RLock()
		for _, v := range data.Result {
			tmp := x.MustMarshal(v)
			id := x.Json.Get(tmp, "entity_id").ToString()
			if _, ok := gHub.entityIdMap[id]; ok {
				filter = append(filter, v)
			}
		}
		gHub.lock.RUnlock()

		data.Result = filter
		data.Total = len(data.Result)
		ava.Debugf("total states=%d", len(data.Result))
		writeToFile("states.json", &data)
		//writeBytesToFile("states_native.json", msg)
		initMutex.Lock()
		initState.statesLoaded = true
		initMutex.Unlock()
		checkInitComplete()
	})
}

// 获取区域名称
func SpiltAreaName(name string) string {
	s := strings.Split(name, " ")
	if len(s) > 1 {
		return s[1]
	}

	return s[0]
}

// 获取家庭名称,如果没有返回房间名称
func SpiltHomeName(name string) string {
	s := strings.Split(name, " ")
	if len(s) > 1 {
		return s[0]
	}

	return s[0]
}

type HttpServiceData struct {
	DeviceId string `json:"device_id,omitempty"`
	EntityId string `json:"entity_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Message  string `json:"message,omitempty"`
	Value    string `json:"value,omitempty"`
}

type HttpServiceDataPlay struct {
	DeviceId    string  `json:"device_id,omitempty"`
	EntityId    string  `json:"entity_id,omitempty"`
	VolumeLevel float64 `json:"volume_level"`
}

type HttpServiceDataPlayPause struct {
	DeviceId      string `json:"device_id,omitempty"`
	EntityId      string `json:"entity_id,omitempty"`
	IsVolumeMuted bool   `json:"is_volume_muted"`
}

type StateChangedSimple struct {
	Type  string `json:"type"`
	Event struct {
		EventType string `json:"event_type"`
		Data      struct {
			EntityID string `json:"entity_id"`
			OldState struct {
				EntityID    string    `json:"entity_id"`
				State       string    `json:"state"`
				LastChanged time.Time `json:"last_changed"`
				LastUpdated time.Time `json:"last_updated"`
			} `json:"old_state"`
			NewState struct {
				Attributes struct {
					FriendlyName string `json:"friendly_name"`
				} `json:"attributes"`
				EntityID    string    `json:"entity_id"`
				State       string    `json:"state"`
				LastChanged time.Time `json:"last_changed"`
				LastUpdated time.Time `json:"last_updated"`
			} `json:"new_state"`
		} `json:"data"`
	} `json:"event"`
	ID int `json:"id"`
}
