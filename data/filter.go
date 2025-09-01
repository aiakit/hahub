package data

import (
	"strings"
	"sync"

	"github.com/panjf2000/ants/v2"
)

// 区域流明配置
var LxArea = make(map[string]*Entity)

const (
	CategoryXiaomiHomeSpeaker   = "xiaomi_home_speaker"   // 小米音箱
	CategoryXiaomiMiotSpeaker   = "xiaomi_miot_speaker"   // 小米MIOT音箱
	CategoryAirConditioner      = "air_conditioner"       // 空调
	CategoryFloorHeating        = "floor_heating"         // 地暖
	CategoryFreshAir            = "flesh_air"             // 地暖
	CategoryVirtualEvent        = "virtual_event"         // 虚拟事件
	CategorySwitch              = "switch"                // 开关
	CategoryWiredSwitch         = "wired_switch"          // 有线开关
	CategorySwitchClickOnce     = "click_once_switch"     // 开关传感器,单击事件
	CategorySwitchScene         = "scene_switch"          // 开关场景按键
	CategorySwitchMode          = "switch_mode"           // 开关模式：判断有线开关和无线开关
	CategoryLight               = "light"                 // 灯总
	CategoryLightTemp           = "light_temp"            // 灯总
	CategoryLightRgb            = "light_rgb"             // 灯rgb
	CategoryLightRgbAndTemp     = "light_rgb_temp"        // 灯rgb和色温
	CategoryLightModel          = "light_mode"            // 灯
	CategoryXinGuang            = "light_xinguang"        // 馨光灯，不含light实体
	CategoryLightGroup          = "light_group"           // 灯组
	CategoryCurtain             = "curtain"               // 窗帘
	CategoryHumanPresenceSensor = "human_presence_sensor" // 存在传感器
	CategorySocket              = "socket"                // 插座
	CategoryHumanBodySensor     = "human_body_sensor"     // 人体传感器
	CategoryTemperatureSensor   = "temperature_sensor"    // 温度传感器
	CategoryHumiditySensor      = "humidity_sensor"       // 湿度传感器
	CategoryLxSensor            = "lx_sensor"             // 光照传感器
	CategoryIrTV                = "ir_tv"                 // 红外电视
	CategoryHaTV                = "ha_tv"                 // 电视品牌插件
	CategoryAutomation          = "automation"            // 自动化
	CategoryScene               = "scene"                 // 场景
	CategoryScript              = "script"                // 场景
	CategoryGas                 = "gas"                   // 天然气
	CategoryWater               = "water"                 // 水侵
	CategoryFire                = "fire"                  // 火灾
	CategoryWaterHeater         = "water_heater"          //热水器
	CategoryPowerconsumption    = "power_consumption"     //用电功率
	CategoryPowe                = "power"                 //电池电量
	CateroyBathroomHeater       = "bathroom_heater"       //浴霸
	CategoryBed                 = "bed"                   //床
	CategoryDoor                = "door"                  //门
	CateOther                   = "other"                 //其他的类型
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
//名字中带“存在传感器”，实体id是“binary_sensor.”开头，实体详情中，State:off表示无人，on表示有人，名字中带有“光照”，“有人持续”，“无人持续”

//插座
//名字中带插座,实体名称中带有“开关状态”

//人体传感器:
//开关上的人体传感器：名字中带有"接近远离"，“感应有人”，“感应无人”

//空调、地暖、温度、湿度
//名字中带有“温度”、“湿度”，开关通过指令直接控制

// 实体过滤函数，按注释规则筛选
// areaMap: area_id -> area_name
// deviceMap: device_id -> *Device

func FilterEntities(entities []*Entity, deviceMap map[string]*Device) []*Entity {
	var filtered []*Entity
	var lock sync.RWMutex

	var areaLxStruct = map[string]struct {
		lx string
		e  *Entity
	}{}

	var waterHeater = make([]*Entity, 0)

	var entityIdMap = make(map[string]*Entity, 20)
	var pool, _ = ants.NewPool(8)
	var wg sync.WaitGroup

	for _, entity := range entities {

		e := entity
		wg.Add(1)

		_ = pool.Submit(func() {

			var category string
			var subCategory string

			func(entity *Entity) {
				defer wg.Done()

				var deviceData *Device
				//脚本和自动化数据是没有设备id的，注意不要动这里的代码
				if v, ok := deviceMap[e.DeviceID]; ok {
					deviceData = v
				}

				if deviceData != nil {
					e.DeviceName = deviceData.Name
					e.AreaID = deviceData.AreaID
					e.AreaName = deviceData.AreaName
					e.DeviceMode = deviceData.Model
					if e.OriginalName == "" && e.Name == "" {
						e.OriginalName = deviceData.Name
					}
				}

				if e.Name != "" {
					e.OriginalName = e.Name
				}

				// 1. 音箱
				if deviceData != nil && strings.Contains(deviceData.Model, ".wifispeaker.") {
					if !strings.Contains(e.OriginalName, "电视") {
						if e.Platform == "xiaomi_home" {
							category = CategoryXiaomiHomeSpeaker
						} else if e.Platform == "xiaomi_miot" {
							category = CategoryXiaomiMiotSpeaker
						}
					}
					return
				}

				// 2. 空调
				if strings.HasPrefix(e.EntityID, "climate.") && strings.Contains(e.OriginalName, "空调") && strings.Contains(e.DeviceName, "空调") {
					category = CategoryAirConditioner
					return
				}
				// 2. 地暖
				if strings.HasPrefix(e.DeviceName, "地暖") && strings.Contains(e.OriginalName, "地暖") && strings.Contains(e.DeviceName, "地暖") {
					category = CategoryFloorHeating
					return
				}

				// 3. 新风
				if strings.HasPrefix(e.DeviceName, "新风") && strings.Contains(e.OriginalName, "新风") && strings.Contains(e.DeviceName, "新风") {
					category = CategoryFreshAir
					return
				}

				// 3. 虚拟事件
				if deviceData != nil && strings.Contains(e.OriginalName, "虚拟事件") && strings.Contains(deviceData.Model, ".gateway.") {
					category = CategoryVirtualEvent
					return
				}

				// 4. 开关,设备和实体都是开关
				if deviceData != nil && strings.Contains(deviceData.Model, ".switch.") && strings.Contains(e.EntityID, "switch.") {
					if deviceData != nil && strings.Contains(deviceData.Model, ".switch.") &&
						!strings.Contains(e.OriginalName, "指示灯") &&
						!strings.Contains(e.OriginalName, "背光") && !strings.Contains(e.OriginalName, "拓展") {
						category = CategorySwitch
					}
					return
				}

				// 4.1 有线开关标记
				if deviceData != nil && strings.Contains(deviceData.Model, ".switch.") && strings.Contains(e.EntityID, "select.") && strings.Contains(e.EntityID, "_mode_") {
					if deviceData != nil && strings.Contains(deviceData.Model, ".switch.") {
						st, err := GetState(e.EntityID)
						if err != nil {
							return
						}
						if strings.Contains(st.State, "有线") {
							category = CategorySwitchMode
							return
						}
					}
				}

				// 4.2 切换类开关实体
				if strings.Contains(e.OriginalName, "开关传感器 单击") && strings.Contains(e.EntityID, "event.") {
					if deviceData != nil && strings.Contains(deviceData.Model, ".switch.") {
						category = CategorySwitchClickOnce
						return
					}
				}

				// 4.3 开关场景按键
				if strings.Contains(e.OriginalName, "场景") && strings.Contains(e.EntityID, "event.") {
					category = CategorySwitchScene
					return
				}

				// 5. 灯
				if strings.HasPrefix(e.EntityID, "light.") && !strings.Contains(e.EntityID, "_group_") && !strings.Contains(e.OriginalName, "指示灯") {
					//有些开关里面的实体有light,直接不用
					if !strings.Contains(e.DeviceName, "开关") {
						category = CategoryLight
						if deviceData != nil {

							state, _ := GetState(e.EntityID)
							if state != nil {
								var existTemp bool
								var existRgb bool
								for _, v := range state.Attributes.SupportedColorModes {
									if v == "rgb" {
										existRgb = true
									}
									if v == "color_temp" {
										existTemp = true
									}
								}
								if existRgb && !existTemp {
									subCategory = CategoryLightRgb
								}

								if !existRgb && existTemp {
									subCategory = CategoryLightTemp
								}

								if existRgb && existTemp {
									subCategory = CategoryLightRgbAndTemp
								}
							}
						}
					}
					return
				}

				////5.1 灯组
				if strings.HasPrefix(e.EntityID, "light.") && strings.Contains(e.EntityID, "_group_") {
					category = CategoryLightGroup
					if deviceData != nil {
						state, _ := GetState(e.EntityID)
						if state != nil {
							var existTemp bool
							var existRgb bool
							for _, v := range state.Attributes.SupportedColorModes {
								if v == "rgb" {
									existRgb = true
								}
								if v == "color_temp" {
									existTemp = true
								}
							}
							if existRgb && !existTemp {
								subCategory = CategoryLightRgb
							}

							if !existRgb && existTemp {
								subCategory = CategoryLightTemp
							}

							if existRgb && existTemp {
								subCategory = CategoryLightRgbAndTemp
							}
						}
					}
					return
				}

				//5.2灯光模式
				if strings.Contains(e.OriginalName, "默认状态 渐变时间设置，字节[0]开灯渐变时间，字节[1]关灯渐变时间，字节[2]模式渐变时间") {
					category = CategoryLightModel
					return
				}
				// 6. 窗帘
				if strings.Contains(e.OriginalName, "窗帘") && strings.HasPrefix(e.EntityID, "cover.") {
					category = CategoryCurtain
					return
				}

				// 7. 存在传感器,包含了binary_sensor
				if strings.Contains(e.EntityID, "sensor.") && (strings.Contains(e.OriginalName, "人在") || strings.Contains(e.OriginalName, "有人无人") || strings.Contains(e.OriginalName, "人体感应")) {
					category = CategoryHumanPresenceSensor
					return
				}

				// 8. 插座
				if deviceData != nil && strings.Contains(deviceData.Model, "plug.") && strings.Contains(deviceData.Name, "插座") && strings.Contains(e.OriginalName, "开关 开关") && strings.HasPrefix(e.EntityID, "switch.") {
					category = CategorySocket
					return
				}

				// 9. 人体传感器,binary_sensor
				if (strings.HasPrefix(e.EntityID, "event.") && strings.Contains(e.OriginalName, "有人")) ||
					(strings.Contains(e.OriginalName, "接近远离") && strings.HasPrefix(e.EntityID, "binary_sensor.")) {
					if deviceData != nil && strings.Contains(deviceData.Name, "-") {
						category = CategoryHumanBodySensor
						return
					}
				}
				// 10. 温度/湿度
				if strings.HasPrefix(e.EntityID, "sensor.") && strings.Contains(e.OriginalName, "温湿度传感器 温度") {
					category = CategoryTemperatureSensor
					return
				}
				if strings.HasPrefix(e.EntityID, "sensor.") && strings.Contains(e.OriginalName, "温湿度传感器 相对湿度") {
					category = CategoryHumiditySensor
					return
				}
				// 11. 光照,如果一个房间有多个，取当前光照值最高的那个
				if strings.HasPrefix(e.EntityID, "sensor.") && strings.Contains(e.OriginalName, "光照") {
					s, err := GetState(e.EntityID)
					if err == nil {
						lock.Lock()
						if v, ok := areaLxStruct[deviceData.AreaID]; ok {
							if strings.Compare(s.State, v.lx) > 0 {
								areaLxStruct[deviceData.AreaID] = struct {
									lx string
									e  *Entity
								}{lx: s.State, e: e}
							}
						} else {
							areaLxStruct[deviceData.AreaID] = struct {
								lx string
								e  *Entity
							}{lx: s.State, e: e}
						}
						lock.Unlock()
					}
					return
				}

				// 12. 红外电视
				if strings.Contains(e.OriginalName, "红外电视") && strings.EqualFold(e.Platform, "xiaomi_home") {
					category = CategoryIrTV
					return
				}

				// 12.1 电视
				if strings.Contains(e.OriginalName, "电视") && strings.Contains(e.EntityID, "media_player.") && !strings.Contains(e.OriginalName, "红外") {
					category = CategoryHaTV
					return
				}

				// 13. 自动化
				if strings.HasPrefix(e.EntityID, "automation.") {
					category = CategoryAutomation
					return
				}
				// 14. 场景
				if strings.HasPrefix(e.EntityID, "scene.") {
					category = CategoryScene
					return
				}

				//15.馨光类型,场景设置中需要实体id
				if deviceData != nil && strings.Contains(deviceData.Name, "馨光") {
					category = CategoryXinGuang
					return
				}

				//16.天然气报警
				if strings.Contains(e.OriginalName, "天然气浓度") {
					category = CategoryGas
					return
				}

				//17.烟雾
				if strings.Contains(e.OriginalName, "检测到高浓度烟雾") {
					category = CategoryFire
					return
				}

				//18.水
				if strings.Contains(e.OriginalName, "检测到") && strings.Contains(e.OriginalName, "水") {
					category = CategoryWater
					return
				}

				//20.热水器
				if deviceData != nil && strings.Contains(deviceData.Name, "热水器") {
					lock.Lock()
					waterHeater = append(waterHeater, e)
					lock.Unlock()
					return
				}

				//21.脚本
				if strings.HasPrefix(e.EntityID, "script.") {
					category = CategoryScript
					return
				}

				//22.功率实体
				if strings.Contains(e.OriginalName, "功耗参数") && strings.HasPrefix(e.EntityID, "sensor.") {
					category = CategoryPowerconsumption
					return
				}

				if strings.Contains(e.OriginalName, "电池电量") && strings.HasPrefix(e.EntityID, "sensor.") {
					category = CategoryPowe
					return
				}

				//23.浴霸
				if deviceData != nil && strings.Contains(deviceData.Name, "浴霸") {
					if strings.HasPrefix(e.EntityID, "climate.") || strings.Contains(e.OriginalName, "换气") {
						category = CateroyBathroomHeater
						return
					}
				}

				//24.床
				if deviceData != nil && strings.Contains(deviceData.Model, ".bed.") {
					category = CategoryBed
					return
				}

				//门
				if deviceData != nil && strings.Contains(deviceData.Model, ".door.") && (strings.Contains(e.OriginalName, "门状态") || strings.Contains(e.OriginalName, "电池电量")) {
					category = CategoryDoor
					return
				}
			}(e)

			if category != "" {
				// 赋值区域信息
				e.Category = category
				e.SubCategory = subCategory
				lock.Lock()
				entityIdMap[e.EntityID] = e
				filtered = append(filtered, e)
				lock.Unlock()
			}
		})
	}
	wg.Wait()
	pool.Release()

	for _, e := range areaLxStruct {
		e.e.Category = CategoryLxSensor
		if deviceMap != nil {
			if dev, ok := deviceMap[e.e.DeviceID]; ok && dev != nil {
				e.e.AreaID = dev.AreaID
				e.e.AreaName = dev.AreaName
				e.e.DeviceName = dev.Name // 新增：赋值设备名称
			}
		}
		entityIdMap[e.e.EntityID] = e.e
		filtered = append(filtered, e.e)
		LxArea[e.e.AreaID] = e.e
	}

	for _, e := range waterHeater {
		e.Category = CategoryWaterHeater
		if deviceMap != nil {
			if dev, ok := deviceMap[e.DeviceID]; ok && dev != nil {
				e.AreaID = dev.AreaID
				e.AreaName = dev.AreaName
				e.DeviceName = dev.Name // 新增：赋值设备名称
			}
		}
		entityIdMap[e.EntityID] = e
		filtered = append(filtered, e)
	}

	// 先构建 device_id -> []*Entity 的映射，方便查找
	deviceEntityMap := make(map[string][]*Entity)
	for _, e := range filtered {
		if e.Category == CategorySwitch || e.Category == CategorySwitchMode {
			deviceEntityMap[e.DeviceID] = append(deviceEntityMap[e.DeviceID], e)
		}
	}

	// 优化：只遍历 deviceEntityMap，每次只处理同一个 device_id 下的实体
	for _, entities := range deviceEntityMap {
		var modeEntities []*Entity
		var switchEntities []*Entity
		// 分类
		for _, e := range entities {
			if e.Category == CategorySwitchMode {
				modeEntities = append(modeEntities, e)
			} else if e.Category == CategorySwitch {
				switchEntities = append(switchEntities, e)
			}
		}

		// 使用并发处理每个switch_mode实体
		var wg sync.WaitGroup
		pool, _ := ants.NewPool(8)

		// 遍历所有 switch_mode
		for _, mmm := range modeEntities {
			wg.Add(1)
			modeEntity := mmm

			_ = pool.Submit(func() {
				defer wg.Done()
				modeState, err := GetState(modeEntity.EntityID)
				if err != nil {
					return
				}
				modeFriendly := modeState.Attributes.FriendlyName
				modePrefix := getPrefix(modeFriendly)
				if !strings.Contains(modeState.State, "有线") {
					return
				}
				// 遍历所有 switch
				for _, swEntity := range switchEntities {

					if swEntity.AreaID != entityIdMap[modeEntity.EntityID].AreaID {
						continue
					}

					swState, err := GetState(swEntity.EntityID)
					if err != nil {
						continue
					}
					swFriendly := swState.Attributes.FriendlyName
					swPrefix := getPrefix(swFriendly)
					if modePrefix == swPrefix {
						if v, ok := deviceMap[swEntity.DeviceID]; ok {
							if strings.Contains(v.Name, "#") {
								swEntity.SubCategory = CategoryWiredSwitch
								swEntity.Category = CategoryLightGroup
							}
							break
						}
					}
				}
			})
		}
		wg.Wait()
		// 释放pool资源
		pool.Release()
	}

	for k, v := range filtered {
		if v.Name != "" {
			filtered[k].OriginalName = v.Name
		}
		gHub.deviceIdState[v.DeviceID] = append(gHub.deviceIdState[v.DeviceID], v)
	}

	return filtered
}

// getPrefix 用于提取 friendly_name 前缀（去掉最后一个空格及其后内容）
func getPrefix(name string) string {
	name = strings.TrimSpace(name)
	lastSpace := strings.LastIndex(name, " ")
	if lastSpace == -1 {
		return name
	}
	return name[:lastSpace]
}
