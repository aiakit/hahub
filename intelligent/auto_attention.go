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
	var title = "燃气泄漏"
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
	var title = "检测到烟雾"
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
	var title = "漏水了"
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

	if ok {
		for _, v := range speakers {
			if strings.Contains(v.OriginalName, "播放文本") && strings.HasPrefix(v.EntityID, "notify.") {
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
		}
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

	ve := virtualEventNotify(title)
	if len(ve) > 0 {
		for _, v := range ve {
			auto.Actions = append(auto.Actions, v)
		}
	}
}

// 虚拟事件通知
func virtualEventNotify(message string) []*ActionCommon {
	entities, ok := data.GetEntityCategoryMap()[data.CategoryVirtualEvent]
	if !ok {
		return nil
	}
	var result = make([]*ActionCommon, 0)

	for _, e := range entities {
		if strings.Contains(e.OriginalName, "产生虚拟事件") && strings.HasPrefix(e.EntityID, "text.") {
			result = append(result, &ActionCommon{
				Type:     "set_value",
				EntityID: e.EntityID,
				Domain:   "text",
				DeviceID: e.DeviceID,
				Value:    message,
			})
		}
	}

	return result
}
