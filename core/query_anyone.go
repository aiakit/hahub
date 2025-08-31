package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"strings"

	"github.com/aiakit/ava"
)

// 判断是否有人
// 某个区域是否有人
// 存在传感器，或者灯
func isAnyoneHere(message, aiMessage, deviceId string) string {
	var resultMessage = ""
	defer func() {
		if resultMessage == "" {
			resultMessage = "条件不满足，没有灯或者人体传感器设备，我无法判断。"
		}
	}()
	var areaName string

	func() {
		var areaNames = data.GetAreaNames()
		//目标区域名称
		for _, v := range areaNames {
			if strings.Contains(message, v) {
				areaName = v
				break
			}
		}

		if areaName == "" {
			//发送所有自动化简短数据给ai
			result, err := chatCompletionInternal([]*chat.ChatMessage{{
				Role:    "user",
				Content: fmt.Sprintf(`根据我的意图，从所有位置区域信息数据%s中，找出最匹配的区域，严格按照名称返回给我。返回格式："区域名称"`, areaNames),
			}, {Role: "user", Content: message}})
			if err != nil {
				ava.Error(err)
				return
			}

			for _, v := range areaNames {
				if strings.Contains(result, v) {
					areaName = v
					break
				}
			}
		}

		if areaName == "" {
			return
		}

		//找到区域
		var areaId string
		for k, v := range data.GetAreaMap() {
			if strings.Contains(v, areaName) {
				areaId = k
				break
			}
		}

		if areaId == "" {
			return
		}

		//判断当前位置是否有名字中带有厕所，卫生间名称设备，灯和传感器
		entities, ok := data.GetEntityAreaMap()[areaId]
		if !ok {
			return
		}

		var exsit bool
		for _, e := range entities {
			if e.Category == data.CategoryLight || e.Category == data.CategoryLightGroup || e.Category == data.CategoryHumanPresenceSensor {
				state, err := data.GetState(e.EntityID)
				if err != nil {
					continue
				}

				if strings.Contains(strings.ToLower(state.State), "on") {
					exsit = true
					break
				}
			}
		}

		if exsit {
			resultMessage = areaName + "有人"
		} else {
			resultMessage = areaName + "没有人"
		}
	}()

	return resultMessage
}
