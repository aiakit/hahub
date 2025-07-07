package hub

//向远程获取方案，将各种方案缓存到本地，本地执行

// 获取区域数据
const getAreaInfoId = 999999999999999991
const getDeviceListId = 999999999999999992
const getEntityListId = 999999999999999993

type areaInfo struct {
	AreaId string `json:"area_id"`
	Name   string `json:"name"`
}

type areaList struct {
	Result []*areaInfo `json:"result"`
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
func callEntityList() {
	var to struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
	}
	to.Id = getEntityListId
	to.Type = "config/device_registry/list"
	CallServiceWs(&to)
}
