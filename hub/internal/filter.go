package internal

import "strings"

const (
	CategoryXiaomiHomeSpeaker   = "xiaomi_home_speaker"   // 小米音箱
	CategoryXiaomiMiotSpeaker   = "xiaomi_miot_speaker"   // 小米MIOT音箱
	CategoryAppleTV             = "apple_tv"              // Apple TV
	CategoryAirConditioner      = "air_conditioner"       // 空调
	CategoryVirtualEvent        = "virtual_event"         // 虚拟事件
	CategorySwitch              = "switch"                // 开关
	CategoryWiredSwitch         = "switch_wired_switch"   // 有线开关
	CategoryLight               = "light"                 // 灯
	CategoryLightGroup          = "light_group"           // 灯组
	CategoryCurtain             = "curtain"               // 窗帘
	CategoryHumanPresenceSensor = "human_presence_sensor" // 存在传感器
	CategorySocket              = "socket"                // 插座
	CategoryHumanBodySensor     = "human_body_sensor"     // 人体传感器
	CategoryTemperatureSensor   = "temperature_sensor"    // 温度传感器
	CategoryHumiditySensor      = "humidity_sensor"       // 湿度传感器
	CategoryLxSensor            = "lx_sensor"             // 光照传感器
	CategoryIrTV                = "ir_tv"                 // 红外电视
	CategoryAutomation          = "automation"            // 自动化
	CategoryScene               = "scene"                 // 场景
)

//过滤实体,并在实体中增加字段标注设备类型，设备数据中也加上，在实体数据中加上设备id,区域id，区域名称

//开关
//从每个开关中取出当名字带有“开关”且model字段包含switch，1.取出entity_id是“button.”开头的实体;2.取出当开关实体详情数据state包含“有线”的实体

//灯，是否是灯组，在场景和自动化中，只需要对灯组进行控制,名字中带有“灯组”
//1.双色温，2.rbg灯(灯带，灯泡等)，3.馨光
//双色温灯直接通过指令控制亮度和色温和开关
//rgb直接通过指令控制亮度、颜色和开关
//馨光灯需要实体带有select和mode，表示模式切换，然后指令控制亮度，静态模式下颜色控制模拟色温，开关灯

//窗帘
//名字中带有窗帘，取出实体id开头是"cover."的实体

//存在传感器:
//名字中带“存在传感器”，实体id是“binary_sensor.”开头，实体详情中，state:off表示无人，on表示有人，名字中带有“光照”，“有人持续”，“无人持续”

//插座
//名字中带插座,实体名称中带有“开关状态”

//人体传感器:
//开关上的人体传感器：名字中带有"接近远离"，“感应有人”，“感应无人”

//空调、地暖、温度、湿度
//名字中带有“温度”、“湿度”，开关通过指令直接控制

// 实体过滤函数，按注释规则筛选
// areaMap: area_id -> area_name
// deviceMap: device_id -> *device
func FilterEntities(entities []*Entity, deviceMap map[string]*device) []*Entity {
	var filtered []*Entity
	// 先处理音箱和apple_tv设备，收集所有相关设备id
	speakerDeviceIDs := make(map[string]*device) // device_id -> category
	for _, dev := range deviceMap {
		if strings.Contains(dev.Name, "音箱") {
			speakerDeviceIDs[dev.ID] = dev
		}
	}
	for _, e := range entities {
		var (
			name     = e.OriginalName
			id       = e.EntityID
			platform = e.Platform
			category = ""
		)

		// 1. 音箱
		if _, ok := speakerDeviceIDs[e.DeviceID]; ok {
			if platform == "xiaomi_home" {
				category = CategoryXiaomiHomeSpeaker
			} else if platform == "xiaomi_miot" {
				category = CategoryXiaomiMiotSpeaker
			}
		}
		// 2. apple_tv
		if platform == "apple_tv" {
			category = CategoryAppleTV
		}
		// 2. 空调
		if strings.HasPrefix(id, "climate.") && strings.Contains(name, "空调") {
			category = CategoryAirConditioner
		}
		// 3. 虚拟事件
		if strings.Contains(name, "虚拟事件") {
			category = CategoryVirtualEvent
		}
		// 4. 开关
		if strings.Contains(name, "开关") && strings.Contains(e.EntityID, "switch") {
			if strings.HasPrefix(id, "button.") {
				category = CategorySwitch
			}
		}
		// 4.1 有线开关
		if strings.Contains(name, "开关") && strings.Contains(name, "有线") {
			if strings.HasPrefix(id, "select.") {
				category = CategoryWiredSwitch
			}
		}

		// 5. 灯
		if strings.HasPrefix(id, "light.") && !strings.Contains(e.EntityID, "_group_") {
			category = CategoryLight
		}
		//5.1 灯组
		if strings.HasPrefix(id, "light.") && strings.Contains(e.EntityID, "_group_") {
			category = CategoryLightGroup
		}

		// 6. 窗帘
		if strings.Contains(name, "窗帘") && strings.HasPrefix(id, "cover.") {
			category = CategoryCurtain
		}
		// 7. 存在传感器
		if strings.Contains(name, "存在传感器") && strings.HasPrefix(id, "binary_sensor.") {
			category = CategoryHumanPresenceSensor
		}
		// 8. 插座
		if strings.HasPrefix(id, "plug.") && strings.Contains(name, "插座") && strings.Contains(name, "开关状态") {
			category = CategorySocket
		}
		// 9. 人体传感器
		if (strings.HasPrefix(id, "sensor.") || strings.HasPrefix(id, "event.")) && (strings.Contains(name, "人体传感器") ||
			(strings.Contains(name, "接近远离") || strings.Contains(name, "感应有人") || strings.Contains(name, "感应无人") ||
				strings.Contains(name, "接近") || strings.Contains(name, "远离"))) {
			category = CategoryHumanBodySensor
		}
		// 10. 温度/湿度
		if strings.HasPrefix(id, "sensor.") && strings.Contains(name, "温度") {
			category = CategoryTemperatureSensor
		}
		if strings.HasPrefix(id, "sensor.") && strings.Contains(name, "湿度") {
			category = CategoryHumiditySensor
		}
		// 11. 光照
		if strings.HasPrefix(id, "sensor.") && strings.Contains(name, "光照") {
			category = CategoryLxSensor
		}
		// 12. 红外电视
		if strings.Contains(name, "红外电视") {
			category = CategoryIrTV
		}
		// 13. 自动化
		if strings.HasPrefix(id, "automation.") {
			category = CategoryAutomation
		}
		// 14. 场景
		if strings.HasPrefix(id, "scene.") {
			category = CategoryScene
		}
		if category != "" {
			// 赋值区域信息
			e.Category = category
			if deviceMap != nil {
				if dev, ok := deviceMap[e.DeviceID]; ok && dev != nil {
					e.AreaID = dev.AreaID
					e.AreaName = dev.AreaName
					e.Name = dev.Name // 新增：赋值设备名称
				}
			}
			filtered = append(filtered, e)
		}
	}
	return filtered
}
