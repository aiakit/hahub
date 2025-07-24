package automation

import (
	"hahub/hub/core"
	"strings"
	"time"

	"github.com/aiakit/ava"
)

var lightMap = make(map[string][]*core.Entity)

func switchRule() {
	light, ok := core.GetEntityCategoryMap()[core.CategoryLightGroup]
	if !ok {
		return
	}
	for _, v := range light {
		_, ok := lightMap[v.AreaID]
		if !ok {
			lightMap[v.AreaID] = make([]*core.Entity, 0)
		}
		lightMap[v.AreaID] = append(lightMap[v.AreaID], v)
	}

	//启动定时器
	core.TimingwheelTicker(10*time.Minute, func() {
		SwitchOff()
	})
}

// 开关关闭逻辑
// 1.开关控制的灯都关闭,2.记录上次灯关时间
func SwitchOff() {
	//找到所有有线开关
	wiredSwitch, ok := core.GetEntityCategoryMap()[core.CategoryWiredSwitch]
	if !ok {
		return
	}

	//找到当前区域的所有名字带有开关按键名字的灯
	for _, v := range wiredSwitch {
		c := ava.Background()
		var now = time.Now()

		v1, err := core.GetState(v.EntityID)
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

			state, err := core.GetState(l.EntityID)
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
			err := core.Post(c, core.GetHassUrl()+"/api/services/switch/turn_off", core.GetToken(), &core.HttpServiceData{EntityId: v.EntityID}, nil)
			if err != nil {
				c.Error(err)
			}
		}
	}
}
