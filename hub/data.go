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
	Id      int         `json:"id"`
	Type    string      `json:"type"`
	Success bool        `json:"success"`
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
type entityList struct {
	ID      int64     `json:"id"`
	Type    string    `json:"type"`
	Success bool      `json:"success"`
	Result  []*entity `json:"result"`
}

type entity struct {
	DeviceID     string `json:"device_id"`     //设备id
	EntityID     string `json:"entity_id"`     //实体id
	ID           string `json:"id"`            //真实实体id，数据id
	OriginalName string `json:"original_name"` //实体名称
	Platform     string `json:"platform"`      //产自什么平台
	UniqueID     string `json:"unique_id"`     //唯一id
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
