package hub

import (
	"os"

	"github.com/aiakit/ava"

	"strings"

	jsoniter "github.com/json-iterator/go"
)

var (
	defaultAreaInfo *areaList
)

func init() {
	// 先启动回调监听
	go callback()
	// 主动获取三类数据
	callAreaList()
	callDeviceList()
	callEntityList()
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
	_ = os.WriteFile(filename, data, 0755)
}

func handleData(data []byte) {

}
