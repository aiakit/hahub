package internal

import (
	"math"
	"strings"
)

//向远程获取方案，将各种方案缓存到本地，本地执行

// 获取区域数据
const (
	getAreaInfoId = (math.MaxInt64 - 1000) + iota
	getDeviceListId
	getEntityListId
	getServicesId
	getStatesId
)

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

func callAreaList() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getAreaInfoId
	to.Type = "config/area_registry/list"
	CallServiceWs(&to)
}

// 获取区域数据
func getAreaData(message string) (*areaList, error) {
	var data areaList
	err := Unmarshal([]byte(message), &data)
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
	Result  []*device `json:"result"`
}

type device struct {
	AreaID     string  `json:"area_id"`     //区域id
	AreaName   string  `json:"area_name"`   //区域名称
	CreatedAt  float64 `json:"created_at"`  //创建时间
	ModifiedAt float64 `json:"modified_at"` //修改时间
	ID         string  `json:"id"`          //设备id
	Model      string  `json:"model"`       //使用的模型
	Name       string  `json:"name"`        //用户命名的产品名称
	SwVersion  string  `json:"sw_version"`  //固件版本
}

func callDeviceList() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getDeviceListId
	to.Type = "config/device_registry/list"
	CallServiceWs(&to)
}

// 获取全量实体数据
type EntityList struct {
	ID      int64     `json:"id"`
	Type    string    `json:"type"`
	Success bool      `json:"success"`
	Total   int       `json:"total"`
	Result  []*Entity `json:"result"`
}

type Entity struct {
	DeviceID     string `json:"device_id"`     //设备id
	EntityID     string `json:"entity_id"`     //实体id
	ID           string `json:"id"`            //真实实体id，数据id
	OriginalName string `json:"original_name"` //实体名称
	Platform     string `json:"platform"`      //产自什么平台
	UniqueID     string `json:"unique_id"`     //唯一id

	Category string `json:"category"`  //设备类型
	AreaID   string `json:"area_id"`   //区域id
	AreaName string `json:"area_name"` //区域名称
	Name     string `json:"name"`      //设备名称（从设备数据获取）
}

func callEntityList() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getEntityListId
	to.Type = "config/entity_registry/list"
	CallServiceWs(&to)
}

// 获取服务数据
type serviceList struct {
	ID      int64          `json:"id"`
	Type    string         `json:"type"`
	Success bool           `json:"success"`
	Total   int            `json:"total"`
	Result  map[string]any `json:"result"`
}

func callServices() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getServicesId
	to.Type = "get_services"
	CallServiceWs(&to)
}

// 获取实体详细信息
// 结构参考Home Assistant get_states返回
// https://developers.home-assistant.io/docs/api/websocket/#get_states

type stateList struct {
	ID      int64         `json:"id"`
	Type    string        `json:"type"`
	Success bool          `json:"success"`
	Total   int           `json:"total"`
	Result  []interface{} `json:"result"`
}

func callStates() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getStatesId
	to.Type = "get_states"
	CallServiceWs(&to)
}

func SpiltAreaName(name string) string {
	s := strings.Split(name, " ")
	if len(s) > 1 {
		return s[1]
	}

	return s[0]
}

type HttpServiceData struct {
	EntityId string `json:"entity_id"`
}
