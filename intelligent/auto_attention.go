package intelligent

import (
	"hahub/data"
	"strings"

	"github.com/aiakit/ava"
)

func attention(c *ava.Context) {
	allEntities := data.GetEntityIdMap()
	for _, entity := range allEntities {
		if entity.Category == data.CategoryGas {
			auto, err := gas(entity)
			if err != nil {
				c.Error(err)
				continue
			}
			CreateAutomation(c, auto)
		}

		if entity.Category == data.CategoryFire {
			auto, err := fire(entity)
			if err != nil {
				c.Error(err)
				continue
			}
			CreateAutomation(c, auto)
		}

		if entity.Category == data.CategoryWater {
			auto, err := water(entity)
			if err != nil {
				c.Error(err)
				continue
			}
			CreateAutomation(c, auto)
		}
	}
}

// 天然气报警
func gas(entity *data.Entity) (*Automation, error) {

	var message = "危险，燃气泄漏了，请立刻处理!!!"
	var title = "危险，危险，燃气泄漏了"
	areaName := data.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "发生燃气泄漏",
		Description: "当" + areaName + "发生燃气泄露时，提醒家人",
		Triggers: []Triggers{{
			Type:     "value",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "sensor",
			Trigger:  "device",
			Above:    1,
			Below:    5,
			For: &For{
				Hours:   0,
				Minutes: 0,
				Seconds: 3,
			},
		}},
		Mode: "single",
	}

	doNotify(title, message, auto)

	return auto, nil
}

func fire(entity *data.Entity) (*Automation, error) {

	var message = "危险，检测到烟雾，请立刻处理!!!"
	var title = "危险，危险，检测到烟雾"
	areaName := data.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "检测到烟雾",
		Description: "当" + areaName + "检测到烟雾，提醒家人",
		Triggers: []Triggers{{
			EntityID: entity.EntityID,
			Trigger:  "state",
			For: &For{
				Hours:   0,
				Minutes: 0,
				Seconds: 3,
			},
		}},
		Mode: "single",
	}

	doNotify(title, message, auto)

	return auto, nil
}

func water(entity *data.Entity) (*Automation, error) {

	var message = "危险,漏水了，请立刻处理!!!"
	var title = "危险，危险，漏水了"
	areaName := data.SpiltAreaName(entity.AreaName)
	auto := &Automation{
		Alias:       areaName + "漏水",
		Description: "当" + areaName + "发生漏水时，提醒家人",
		Triggers: []Triggers{{
			Type:     "moist",
			DeviceID: entity.DeviceID,
			EntityID: entity.EntityID,
			Domain:   "binary_sensor",
			Trigger:  "device",
			For: &For{
				Hours:   0,
				Minutes: 0,
				Seconds: 3,
			},
		}},
		Mode: "single",
	}

	doNotify(title, message, auto)

	return auto, nil
}

func doNotify(title, message string, auto *Automation) {
	notifyPhoneId := data.GetNotifyPhone()

	auto.Actions = append(auto.Actions, &ActionNotify{
		Action: "notify.persistent_notification",
		Data: struct {
			Message string `json:"message,omitempty"`
			Title   string `json:"title,omitempty"`
		}{message, title},
	})

	speakers, ok := data.GetEntityCategoryMap()[data.CategoryXiaomiHomeSpeaker]
	if !ok {
		return
	}

	for _, v := range speakers {
		if !strings.Contains(v.OriginalName, "播放文本") {
			continue
		}
		auto.Actions = append(auto.Actions, &ActionNotify{
			Action: "notify.send_message",
			Data: struct {
				Message string `json:"message,omitempty"`
				Title   string `json:"title,omitempty"`
			}{message, title},
			Target: struct {
				DeviceID string `json:"device_id,omitempty"`
			}{DeviceID: v.DeviceID},
		})
	}

	for _, v := range notifyPhoneId {
		auto.Actions = append(auto.Actions, &ActionNotify{
			Action: "notify." + v,
			Data: struct {
				Message string `json:"message,omitempty"`
				Title   string `json:"title,omitempty"`
			}{message, title},
		})
	}
}
