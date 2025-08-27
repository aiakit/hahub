package intelligent

import (
	"errors"
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// -符号是人体感应专属
// 遍历所有人体传感器，生成自动化
// 如果相同区域有多个前缀相同的传感器触发器，则要互斥
// 开灯30秒后自动关灯
func WalkBodySensor(c *ava.Context) {
	// 查询所有实体，找到名字中带有'-'的实体
	allEntities := data.GetEntityByEntityId()
	var sensors []*data.Entity
	for _, e := range allEntities {
		if strings.Contains(e.DeviceName, "-") && (e.Category == data.CategoryLightGroup || e.Category == data.CategoryLight || e.Category == data.CategoryHumanBodySensor) {
			sensors = append(sensors, e)
		}
	}

	for _, v := range sensors {

		autoOn, err := bodySensorOn(v)
		if err != nil {
			c.Errorf("entity=%s |err=%v", x.MustMarshal2String(v), err)
			continue
		}

		if autoOn != nil {
			AddAutomation2Queue(c, autoOn)
		}
	}
}

// 人体传感器有人时自动开灯/开关
func bodySensorOn(entity *data.Entity) (*Automation, error) {

	areaID := entity.AreaID
	areaName := data.SpiltAreaName(entity.AreaName)
	entities, ok := data.GetEntityAreaMap()[areaID]
	if !ok {
		return nil, errors.New("entity area not found")
	}

	// 1. 取entity.Name中'-'前的前缀
	prefix := entity.DeviceName
	suffix := ""
	if idx := strings.Index(prefix, "-"); idx > 0 {
		suffix = prefix[idx+1:]
		prefix = prefix[:idx]
	}

	entitiesFilter := findLightsWithOutLightCategory(prefix, entities)
	if len(entitiesFilter) == 0 {
		return nil, fmt.Errorf("%s区域没有发现灯", areaName)
	}

	actions := turnOnLights(entitiesFilter, 100, 4800, true)

	if len(actions) == 0 {
		return nil, fmt.Errorf("%s区域没有发现灯", areaName)
	}

	// 30秒后关灯
	actions = append(actions, &ActionLight{Delay: &delay{
		Seconds: 30,
	}})

	actions = append(actions, turnOffLights(entitiesFilter)...)

	sensorPrefixStr := prefix

	triggerType := "occupied"
	triggerDomain := "binary_sensor"
	triggerTrigger := "device"
	triggerDeviceId := entity.DeviceID

	var con = make([]*Conditions, 0)
	if strings.HasPrefix(entity.EntityID, "event.") {
		triggerType = ""
		triggerDomain = ""
		triggerDeviceId = ""
		triggerTrigger = "state"
		if name := splitLatest(entity.OriginalName); name != "" {
			con = append(con, &Conditions{
				Condition: "state",
				EntityID:  entity.EntityID,
				Attribute: "event_type",
				State:     name,
			})
		}
	}

	if entity.Category == data.CategoryLight {
		triggerType = "turned_on"
		triggerDomain = "light"
	}

	if strings.EqualFold(areaName, sensorPrefixStr) {
		sensorPrefixStr = ""
	}

	suffixStr := suffix
	if suffixStr != "" {
		suffixStr = strings.TrimSpace(suffixStr)
	}

	condition, action := spiltCondition(entity, actions)

	auto := &Automation{
		Alias:       areaName + prefix + suffixStr + "人来亮灯",
		Description: "当检测到有人，自动打开" + areaName + "下同名前缀的灯和开关",
		Triggers: []*Triggers{{
			Type:     triggerType,
			DeviceID: triggerDeviceId,
			EntityID: entity.EntityID,
			Domain:   triggerDomain,
			Trigger:  triggerTrigger,
		}},
		Conditions: condition,
		Actions:    action,
		Mode:       "single",
	}

	if len(con) > 0 {
		auto.Conditions = append(auto.Conditions, con...)
	}

	if strings.Contains(prefix, "夜") {
		auto.Alias = areaName + suffixStr + "起夜场景"
	}

	// 增加光照条件
	lxConfig := getLxConfig(areaID)
	if lxConfig != nil {
		auto.Conditions = append(auto.Conditions, &Conditions{
			Condition: "numeric_state",
			EntityID:  lxConfig.EntityId,
			Below:     lxConfig.Lx, // 设置光照阈值
		})
	}

	return auto, nil
}

func splitLatest(name string) string {
	s := strings.Split(name, " ")
	if len(s) > 1 {
		return s[len(s)-1]
	}

	return ""
}
