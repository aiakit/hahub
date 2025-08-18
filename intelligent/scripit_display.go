package intelligent

import (
	"hahub/data"
	"hahub/x"
	"sort"
	"strings"

	"github.com/aiakit/ava"
)

type sortLight struct {
	Entity *data.Entity
	Number int
}

// 按照编号执行灯光展示，非卧室，带有编号的灯
func Display(c *ava.Context) {
	var entities, ok = data.GetEntityCategoryMap()[data.CategoryLight]
	if !ok {
		return
	}

	var entities2, ok1 = data.GetEntityCategoryMap()[data.CategoryXinGuang]
	if !ok1 {
		return
	}

	entityMapMode := make(map[string]*data.Entity)
	var entityMode, ok3 = data.GetEntityCategoryMap()[data.CategoryLightModel]
	if ok3 {
		for _, e := range entityMode {
			entityMapMode[e.DeviceID] = e
		}
	}

	for _, e := range entities2 {
		if strings.HasPrefix(e.EntityID, "light.") {
			entities = append(entities, e)
		}
	}

	var script = &Script{
		Alias:       "客厅逐个亮灯展示场景",
		Description: "客厅区域按照灯名称顺序逐个将灯打开",
		Sequence:    nil,
	}

	sXinguang := InitModeThree(c, 1)

	if sXinguang != nil && len(sXinguang.Sequence) > 0 {
		script.Sequence = append(script.Sequence, sXinguang.Sequence...)
	}

	// 找出所有灯带编号，按照从1开始顺序加入到一个数组中
	var lightStripNumbers []*sortLight

	// 遍历所有灯实体
	for _, entity := range entities {
		if strings.Contains(entity.AreaName, "卧室") {
			continue
		}

		i := x.ExtractAndCombineNumbers(entity.DeviceName)
		if i == 0 {
			continue
		}

		lightStripNumbers = append(lightStripNumbers, &sortLight{
			Entity: entity,
			Number: i,
		})
	}

	// 按照Number进行排序
	sort.Slice(lightStripNumbers, func(i, j int) bool {
		return lightStripNumbers[i].Number < lightStripNumbers[j].Number
	})

	//先把灯设置成开灯很快
	for _, s := range lightStripNumbers {
		if v, ok := entityMapMode[s.Entity.DeviceID]; ok {
			actionCommon := handleDefaultGradientTimeSettings(v, 2)
			if actionCommon != nil {
				script.Sequence = append(script.Sequence, actionCommon)
			}
		}
	}

	//开灯
	for index, s := range lightStripNumbers {
		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Milliseconds: 100},
		})

		if strings.Contains(s.Entity.DeviceName, "彩") || strings.Contains(s.Entity.DeviceName, "楼梯") {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
				},
				Target: &targetLightData{DeviceId: s.Entity.DeviceID},
			})
			continue
		}

		if s.Entity.Category != data.CategoryXinGuang {
			entity := s.Entity
			colorTemp := 5800
			if index%2 == 1 {
				colorTemp = 2700
			}
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					ColorTempKelvin: colorTemp,
					BrightnessPct:   100,
				},
				Target: &targetLightData{DeviceId: entity.DeviceID},
			})
			continue
		}

		if s.Entity.Category == data.CategoryXinGuang {
			script.Sequence = append(script.Sequence, ActionLight{
				Action: "light.turn_on",
				Data: &actionLightData{
					BrightnessPct: 100,
					RgbColor:      GetRgbColor(5000),
				},
				Target: &targetLightData{DeviceId: s.Entity.DeviceID},
			})
		}
	}

	//再改回去
	for _, s := range lightStripNumbers {
		if v, ok := entityMapMode[s.Entity.DeviceID]; ok {
			actionCommon := handleDefaultGradientTimeSettings(v, 1)
			if actionCommon != nil {
				script.Sequence = append(script.Sequence, actionCommon)
			}
		}
	}

	if len(script.Sequence) > 0 {
		CreateScript(c, script)
	}
}
