package intelligent

import (
	"hahub/data"
	"hahub/x"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

var lightMap = make(map[string][]*data.Entity)

func switchRule() {
	light, ok := data.GetEntityCategoryMap()[data.CategoryLightGroup]
	if !ok {
		return
	}
	for _, v := range light {
		_, ok := lightMap[v.AreaID]
		if !ok {
			lightMap[v.AreaID] = make([]*data.Entity, 0)
		}
		lightMap[v.AreaID] = append(lightMap[v.AreaID], v)
	}

	//启动定时器
	x.TimingwheelTicker(10*time.Minute, func() {
		SwitchOff()
	})
}

// 开关关闭逻辑
// 1.开关控制的灯都关闭,2.记录上次灯关时间
func SwitchOff() {
	//找到所有有线开关
	wiredSwitch, ok := data.GetEntityCategoryMap()[data.CategoryWiredSwitch]
	if !ok {
		return
	}

	//找到当前区域的所有名字带有开关按键名字的灯
	for _, v := range wiredSwitch {
		c := ava.Background()
		var now = time.Now()

		v1, err := data.GetState(v.EntityID)
		if err != nil {
			c.Error(err)
			continue
		}

		if v1.State == "off" || v1.State == "unavailable" {
			continue
		}

		//判断关闭时间
		if now.Sub(v1.LastChanged) < time.Minute*10 {
			continue
		}

		var closeSwitch bool
		lights, ok := lightMap[v.AreaID]
		if !ok {
			continue
		}

		var count int
		for _, l := range lights {
			if strings.Contains(l.OriginalName, "氛围") || strings.Contains(l.OriginalName, "台灯") {
				continue
			}

			count++

			state, err := data.GetState(l.EntityID)
			if err != nil {
				c.Error(err)
				break
			}

			if state.State == "on" {
				closeSwitch = true
				break
			}

			//判断关闭时间
			if now.Sub(state.LastChanged) < time.Minute*10 {
				closeSwitch = true
				break
			}
		}

		//如果没有灯，表示只有智能开关，这种情况就不去关闭智能开关
		if !closeSwitch && count > 0 {
			//关闭开关
			err := x.Post(c, data.GetHassUrl()+"/api/services/switch/turn_off", data.GetToken(), &data.HttpServiceData{EntityId: v.EntityID}, nil)
			if err != nil {
				c.Error(err)
			}
		}
	}
}

// var switchSelectSceneSwitchMap = make(map[string][]*switchSelect) //key:areaID
// var switchSelectClickOneMap = make(map[string][]*switchSelect)    //key:areaID
var switchSelectSameName = make(map[string][]*switchSelect) //key:click once button name or scene button name

// 开关选择:场景按键，开关按键
type switchSelect struct {
	ButtonName string `json:"button_name"`
	Category   string `json:"category"`
	EntityID   string `json:"entity_id"`
	DeviceID   string `json:"device_id"`
	SeqButton  int    `json:"seq_button"`
	AreaID     string `json:"area_id"`
	AreaName   string `json:"area_name"`
	Attribute  string `json:"attribute"`
}

func InitSwitchSelect(c *ava.Context) {

	entities := data.GetEntityAreaMap()
	for areaID, v := range entities {
		for _, e := range v {
			if e.Category != data.CategorySwitchScene && e.Category != data.CategorySwitchClickOnce {
				continue
			}

			areaName := data.SpiltAreaName(data.GetAreaName(areaID))
			bn := strings.Trim(e.Name, " ")

			//如果是场景开关
			if e.Category == data.CategorySwitchScene {
				var ss = &switchSelect{
					ButtonName: bn,
					Category:   data.CategorySwitchScene,
					EntityID:   e.EntityID,
					DeviceID:   e.DeviceID,
					SeqButton:  0,
					AreaID:     areaID,
					AreaName:   areaName,
				}
				key := areaID + "_" + bn
				switchSelectSameName[key] = append(switchSelectSameName[key], ss)
			}

			//如果是开关按键
			if e.Category == data.CategorySwitchClickOnce {
				name := strings.Split(e.OriginalName, " ")
				var ss = &switchSelect{
					Category: data.CategorySwitchClickOnce,
					EntityID: e.EntityID,
					DeviceID: e.EntityID,
					AreaID:   areaID,
					AreaName: areaName,
				}

				for _, v1 := range name {
					if v1 != " " && v1 != "" {
						ss.ButtonName = v1
						break
					}
				}

				switch {
				case strings.Contains(e.OriginalName, "按键1"):
					ss.SeqButton = 1
				case strings.Contains(e.OriginalName, "按键2"):
					ss.SeqButton = 2
				case strings.Contains(e.OriginalName, "按键3"):
					ss.SeqButton = 3
				case strings.Contains(e.OriginalName, "按键4"):
					ss.SeqButton = 4
				default:
					break
				}

				if ss.SeqButton > 0 {
					ss.Attribute = "按键类型"
				}

				key := areaID + "_" + ss.ButtonName

				switchSelectSameName[key] = append(switchSelectSameName[key], ss)
			}
		}
	}
}
