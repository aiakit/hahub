package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"
	"strings"

	"github.com/aiakit/ava"
)

// 定时周期执行的动作：设备，场景，自动化
func RunTming(message, aiMessage, deviceId string) string {
	var auto = &intelligent.Automation{}

	if strings.Contains(aiMessage, "control_device") {
		var devices = data.GetDeviceByName()

		//获取所有实体状态,再匹配实体
		states, err := data.GetStates()
		if err != nil {
			ava.Error(err)
			return "没有找到设备数据"
		}

		var entities = make([]shortDevice, 0, 20)
		var entitiesMap = make(map[string]data.StateAll, 20)

		for _, state := range states {
			name, ok := state.Attributes["friendly_name"].(string)
			if !ok {
				continue
			}

			n := strings.Split(name, " ")
			if len(n) == 0 {
				continue
			}

			deviceName := n[0]
			d, ok := devices[deviceName]
			if !ok {
				continue
			}

			var isExist bool
			for _, v := range d {
				if v.EntityID == state.EntityID {
					isExist = true
					break
				}
			}

			if !isExist {
				continue
			}

			entities = append(entities, shortDevice{
				Name:  deviceName,
				id:    state.EntityID,
				State: state.State,
			})

			entitiesMap[state.EntityID] = state
		}

		result, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`这是我的全部设备信息%s，选择对应的设备信息按照格式返回给我，
返回格式: [{"name":"设备名称","id":"设备id"}]`, x.MustMarshalEscape2String(entities))},
			{Role: "user", Content: message},
		}, deviceId)
		if err != nil {
			ava.Error(err)
			return "没有找到对应的设备信息"
		}

		var entity = make([]data.StateAll, 0, 2)
		var isServiceExist = make(map[string]bool)
		var command = make(map[string]interface{})

		for _, v := range entities {
			if strings.Contains(result, v.id) {
				entity = append(entity, entitiesMap[v.id])

				var sp = strings.Split(v.id, ".")
				if len(sp) == 0 {
					continue
				}
				cc := sp[0]
				if _, ok := isServiceExist[cc]; ok {
					continue
				}

				v1, ok := data.GetService()[cc]
				if !ok {
					continue
				}
				command[cc] = v1
			}
		}

		result1, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`分析我提供的数据，根据我的意图，按照格式返回数据给我。

设备数据：%s;
指令数据：%s;
返回JSON格式:{
        "alias": "自动化名称",
        "description": "自动化功能描述",
        "triggers": [
            {
                "trigger": "time",
				"at":"22:23:05"
            }
        ],
		"conditions": [
		   {
            "condition": "time",
            "weekday": [
                "mon",
                "tue"
            ]
      	  }
       ],    
		"actions": [
            {
                "type": "turn_on",
				"domain":"light",
				"brightness_pct":100;
				"entity_id":"68d419bf3cc1a0a94e1d82fe3c5bbda3",
            }
        ]
    }

数据格式说明：
alias：根据意图，给这个自动化命名，要求简短。
description: 对这个自动化的功能进行描述。
triggers：表示在某个时间点执行，如果我的意图比较模糊例如：下午就帮我选一个下午的时间点赋值到at中。
conditions：表示周期性的weekday有7个值可选，分别是周一(mon)、二(tue)、三(wed)、四(thu)、五(fri)、六(sat)、日(sun)。
actions：表示要控制的设备，通过设备数据和指令数据得到。`, entity, command)},
		}, deviceId)

		if err != nil {
			ava.Error(err)
			return "出了点小问题"
		}

		err = x.Unmarshal([]byte(x.FindJSON(result1)), auto)
		if err != nil {
			ava.Error(err)
			return "创建自动化失败了"
		}

		auto.Mode = "single"
		auto.Alias += "*"
		intelligent.AddAutomation2Queue(ava.Background(), auto)

		return "已为你创建了" + auto.Alias
	}

	if strings.Contains(aiMessage, "scene") {
		var gShortScenes = make(map[string]*shortScene)

		entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
		if !ok {
			return ""
		}

		for _, e := range entities {
			gShortScenes[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		//发送所有场景简短数据给ai
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role: "user",
			Content: fmt.Sprintf(`这是我的全部场景信息%s，总数是%d个，根据对话内容将信息返回给我：
1.查询某个场景："id":""
2.查询场景数量："你总共有x个场景"`, x.MustMarshalEscape2String(gShortScenes), len(gShortScenes)),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string

		for _, v := range gShortScenes {
			if strings.Contains(result, v.id) {
				id = v.id
			}
		}

		if id == "" {
			return result
		}

		result1, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`分析我提供的数据，根据我的意图，按照格式返回数据给我。

场景数据：entity_id=%s;
返回JSON格式:{
        "alias": "自动化名称",
        "description": "自动化功能描述",
        "triggers": [
            {
                "trigger": "time",
				"at":"22:23:05"
            }
        ],
		"conditions": [
		   {
            "condition": "time",
            "weekday": [
                "mon",
                "tue"
            ]
      	  }
       ],    
		"actions": [
            {
                "action": "script.turn_on",
				"target":{"entity_id":"script.dian_jing_fang_diao_di_deng_guang_liang_du"},
            }
        ]
    }

数据格式说明：
alias：根据意图，给这个自动化命名，要求简短。
description: 对这个自动化的功能进行描述。
triggers：表示在某个时间点执行，如果我的意图比较模糊例如：下午就帮我选一个下午的时间点赋值到at中。
conditions：表示周期性的weekday有7个值可选，分别是周一(mon)、二(tue)、三(wed)、四(thu)、五(fri)、六(sat)、日(sun)。
actions：表示要启动的脚本场景entity_id。`, id)},
		}, deviceId)

		if err != nil {
			ava.Error(err)
			return "出了点小问题"
		}

		err = x.Unmarshal([]byte(x.FindJSON(result1)), auto)
		if err != nil {
			ava.Error(err)
			return "创建自动化失败了"
		}

		auto.Mode = "single"
		auto.Alias += "*"
		intelligent.AddAutomation2Queue(ava.Background(), auto)

		return "已为你创建了" + auto.Alias

	}

	if strings.Contains(aiMessage, "automation") {
		var gShortAutomations = make(map[string]*shortScene)
		entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
		if !ok {
			return ""
		}

		for _, e := range entities {
			gShortAutomations[e.UniqueID] = &shortScene{
				id:    e.EntityID,
				Alias: e.OriginalName,
			}
		}

		//发送所有自动化简短数据给ai
		result, err := chatCompletionHistory([]*chat.ChatMessage{{
			Role: "user",
			Content: fmt.Sprintf(`这是我的全部自动化信息%s，总数是%d个，根据对话内容将信息返回给我：
1.查询某个自动化："id":""
2.查询自动化数量："你总共有x个自动化"`, x.MustMarshalEscape2String(gShortAutomations), len(gShortAutomations)),
		}, {Role: "user", Content: message}}, deviceId)
		if err != nil {
			ava.Error(err)
			return "服务器开小差了，请等一会儿再试试"
		}
		var id string

		for _, v := range gShortAutomations {
			if strings.Contains(result, v.id) {
				id = v.id
			}
		}

		if id == "" {
			return result
		}

		result1, err := chatCompletionHistory([]*chat.ChatMessage{
			{Role: "user", Content: fmt.Sprintf(`分析我提供的数据，根据我的意图，按照格式返回数据给我。

自动化数据：entity_id=%s;
返回JSON格式:{
        "alias": "自动化名称",
        "description": "自动化功能描述",
        "triggers": [
            {
                "trigger": "time",
				"at":"22:23:05"
            }
        ],
		"conditions": [
		   {
            "condition": "time",
            "weekday": [
                "mon",
                "tue"
            ]
      	  }
       ],    
		"actions": [
            {
                "action": "automation.trigger",
				"data":{"skip_condition":true},
				"target":{"entity_id":"script.dian_jing_fang_diao_di_deng_guang_liang_du"},
            }
        ]
    }

数据格式说明：
alias：根据意图，给这个自动化命名，要求简短。
description: 对这个自动化的功能进行描述。
triggers：表示在某个时间点执行，如果我的意图比较模糊例如：下午就帮我选一个下午的时间点赋值到at中。
conditions：表示周期性的weekday有7个值可选，分别是周一(mon)、二(tue)、三(wed)、四(thu)、五(fri)、六(sat)、日(sun)。
actions：表示要启动的脚本场景entity_id。`, id)},
		}, deviceId)

		if err != nil {
			ava.Error(err)
			return "出了点小问题"
		}

		err = x.Unmarshal([]byte(x.FindJSON(result1)), auto)
		if err != nil {
			ava.Error(err)
			return "创建自动化失败了"
		}

		auto.Mode = "single"
		auto.Alias += "*"
		intelligent.AddAutomation2Queue(ava.Background(), auto)

		return "已为你创建了" + auto.Alias
	}

	return "未知的动作"
}
