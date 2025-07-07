package hub

import (
	"github.com/aiakit/ava"

	jsoniter "github.com/json-iterator/go"
)

var (
	defaultAreaInfo *areaList
)

func init() {

	callback()
}

// 初始化数据，设备信息-区域信息
// 版本信息，获取版本号，替换缓存数据
func callback() {
	for msg := range channelMessage {

		id := jsoniter.Get(msg, "id").ToInt64()
		tpe := jsoniter.Get(msg, "type").ToString()
		success := jsoniter.Get(msg, "success").ToBool()

		if tpe == "result" && success == true {
			switch id {
			case getAreaInfoId: //获取区域数据
			case getDeviceListId: //获取设备数据
			case getEntityListId: //获取实体数据
			default:
				ava.Debugf("no id function |data=%v", BytesToString(msg))
			}
		}

		if tpe == "event" {
			handleData(msg)
		}

	}
}

func handleData(data []byte) {

}
