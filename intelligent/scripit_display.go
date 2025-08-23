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

var displayEntityId string

// 按照编号执行灯光展示，非卧室，带有编号的灯
func Display(c *ava.Context) {
	var entities, ok = data.GetEntityCategoryMap()[data.CategoryLight]
	if !ok {
		return
	}

	entityMapMode := make(map[string]*data.Entity)
	var entityMode, ok3 = data.GetEntityCategoryMap()[data.CategoryLightModel]
	if ok3 {
		for _, e := range entityMode {
			entityMapMode[e.DeviceID] = e
		}
	}

	var script = &Script{
		Alias:       "客厅逐个亮灯展示场景",
		Description: "客厅区域按照灯名称顺序逐个将灯打开",
		Sequence:    nil,
	}

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
	var parallel1 = make(map[string][]interface{})

	for _, s := range lightStripNumbers {
		if v, ok := entityMapMode[s.Entity.DeviceID]; ok {
			actionCommon := handleDefaultGradientTimeSettings(v, 2)
			if actionCommon != nil {
				parallel1["parallel"] = append(parallel1["parallel"], actionCommon)
			}
		}
	}

	if len(parallel1) > 0 {
		script.Sequence = append(script.Sequence, parallel1)
	}

	var entitiesLight = make([]*data.Entity, 0, 2)
	for _, e := range lightStripNumbers {
		entitiesLight = append(entitiesLight, e.Entity)
	}

	actions := turnOnLights(entitiesLight, 100, 4800, false)
	if len(actions) == 0 {
		return
	}

	var actionsSort = make([]*ActionLight, 0, len(actions))
	for _, e := range lightStripNumbers {
		for _, e1 := range actions {
			deviceId := e1.DeviceID
			if deviceId == "" && e1.Target != nil {
				deviceId = e1.Target.DeviceId
			}

			if deviceId == "" {
				continue
			}

			if e.Entity.DeviceID == deviceId {
				actionsSort = append(actionsSort, e1)
				break
			}
		}
	}

	var sequence = make(map[string][]interface{}, 2)

	for _, e := range actionsSort {
		sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Milliseconds: 700},
		})
		sequence["sequence"] = append(sequence["sequence"], e)
	}

	if len(sequence) > 0 {
		script.Sequence = append(script.Sequence, ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 2},
		})

		sequence["sequence"] = append(sequence["sequence"], ActionTimerDelay{
			Delay: struct {
				Hours        int `json:"hours"`
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			}{Seconds: 2},
		})

		var parallel2 = make(map[string][]interface{})
		//再改回去
		for _, s := range lightStripNumbers {
			if v, ok := entityMapMode[s.Entity.DeviceID]; ok {
				actionCommon := handleDefaultGradientTimeSettings(v, 2)
				if actionCommon != nil {
					parallel2["parallel"] = append(parallel2["parallel"], actionCommon)
				}
			}
		}

		if len(parallel2) > 0 {
			sequence["sequence"] = append(sequence["sequence"], parallel2)
		}

		script.Sequence = append(script.Sequence, sequence)
	}

	if len(script.Sequence) > 0 {
		displayEntityId = "script." + AddScript2Queue(c, script)
	}
}
