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
	DeviceName string `json:"device_name"`
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

	func() {
		entities := data.GetEntityCategoryMap()[data.CategorySwitchScene]
		//场景按键
		for _, e := range entities {
			areaName := data.SpiltAreaName(e.AreaName)
			bn := strings.Trim(e.Name, " ")

			//如果是场景开关
			var ss = &switchSelect{
				ButtonName: bn,
				Category:   data.CategorySwitchScene,
				EntityID:   e.EntityID,
				DeviceID:   e.DeviceID,
				SeqButton:  0,
				AreaID:     e.AreaID,
				AreaName:   areaName,
				DeviceName: e.DeviceName,
			}
			key := e.AreaID + "_" + bn
			switchSelectSameName[key] = append(switchSelectSameName[key], ss)
		}

	}()

	func() {
		entitiesClick, ok := data.GetEntityCategoryMap()[data.CategorySwitchClickOnce]
		if !ok {
			return
		}

		var entitiesMap = map[string][]*data.Entity{}
		for _, e := range entitiesClick {
			entitiesMap[e.DeviceID] = append(entitiesMap[e.DeviceID], e)
		}

		entitiesSwitch, ok := data.GetEntityCategoryMap()[data.CategorySwitch]
		if !ok {
			return
		}

		for _, e := range entitiesSwitch {
			if !strings.Contains(e.OriginalName, "键") {
				continue
			}
			entitiesMap[e.DeviceID] = append(entitiesMap[e.DeviceID], e)
		}

		//处理单开
		var entityMap = make(map[string][]*data.Entity)
		for _, ee := range entitiesSwitch {
			if ee.Category == data.CategorySwitch {
				entityMap[ee.DeviceID] = append(entityMap[ee.DeviceID], ee)
			}
		}

		for _, e := range entitiesSwitch {
			if strings.Contains(e.OriginalName, "-") {
				continue
			}
			name := strings.Split(e.OriginalName, " ")
			buttonName := ""
			buttonFlag := ""
			areaName := data.SpiltAreaName(e.AreaName)
			SeqButton := 1

			var ss = &switchSelect{
				Category:   data.CategorySwitchClickOnce,
				EntityID:   e.EntityID,
				DeviceID:   e.DeviceID,
				AreaID:     e.AreaID,
				AreaName:   areaName,
				DeviceName: e.DeviceName,
				ButtonName: buttonName,
				Attribute:  "按键类型",
			}

			for _, v1 := range name {
				if v1 != " " && v1 != "" {
					buttonName = v1
					break
				}
			}

			if len(name) >= 1 {
				buttonFlag = name[len(name)-1]
			} else {
				continue
			}

			ss.ButtonName = buttonName

			if len(entityMap[e.DeviceID]) == 1 {
				ss.ButtonName = e.DeviceName
			}

			switch {
			case strings.Contains(e.OriginalName, "按键1"):
				SeqButton = 1
			case strings.Contains(e.OriginalName, "按键2"):
				SeqButton = 2
			case strings.Contains(e.OriginalName, "按键3"):
				SeqButton = 3
			case strings.Contains(e.OriginalName, "按键4"):
				SeqButton = 4
			}

			ss.SeqButton = SeqButton

			if strings.Contains(e.OriginalName, "按键") || buttonFlag == "开关" {
				for _, ee := range entitiesMap[e.DeviceID] {
					if strings.Contains(ee.OriginalName, "开关传感器 单击") {
						key := ee.AreaID + "_" + ss.ButtonName
						switchSelectSameName[key] = append(switchSelectSameName[key], ss)
						break
					}
				}
			} else {
				for _, ee := range entitiesMap[e.DeviceID] {
					if strings.Contains(ee.OriginalName, buttonFlag) {
						key := ee.AreaID + "_" + ss.ButtonName
						switchSelectSameName[key] = append(switchSelectSameName[key], ss)
						break
					}
				}
			}
		}

	}()
}
