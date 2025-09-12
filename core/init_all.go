package core

import (
	"fmt"
	"hahub/data"
	"hahub/intelligent"
	"hahub/internal/chat"
	"hahub/x"

	"github.com/aiakit/ava"
)

func InitALL(message, aiMessage, deviceId string) string {
	go func() {
		var resultMessage string
		defer func() {
			PlayTextAction(deviceId, fmt.Sprintf("系统初始化完成，%s", resultMessage))
		}()
		intelligent.Chaos()
		//智能家居评估C,B,A,S,SS,SSS，热水器，空调，灯光，扫地机，地暖，插座，电视，开关，浴霸，人体传感器，水浸传感器，烟雾报警器，燃气报警器，门锁，门，窗帘，床，洗衣机，冰箱，洗碗机，温度，湿度,等等进行综合评估
		//智能家居建议
		//设备数量，开启情况
		//场景数量介绍,大致有哪些场景
		//自动化情况介绍，大致有哪些自动化
		device := data.GetDevices()
		var d = make([]*shortDevice, 0)
		for _, e := range device {
			d = append(d, &shortDevice{
				Name: e.Name,
				id:   e.ID,
			})
		}
		scene := data.GetEntityCategoryMap()[data.CategoryScript]
		var s = make([]*shortScene, 0)
		for _, e := range scene {
			s = append(s, &shortScene{
				Alias: e.OriginalName,
				id:    e.EntityID,
			})
		}

		auto := data.GetEntityCategoryMap()[data.CategoryAutomation]
		var a = make([]*shortScene, 0)
		for _, e := range auto {
			a = append(a, &shortScene{
				Alias: e.OriginalName,
				id:    e.EntityID,
			})
		}

		result, err := chatCompletionInternal([]*chat.ChatMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`你是一个智能家居专家，现在你需要根据当前智能家居情况进行人性化的描述，300字左右，需要突出重点，然后按照C,B,A,S,SS,SSS等级对当前智能家居系统进行评估。
当前设备信息:%s。
当前场景信息:%s。
当前是否使用AI助手：是。
当前自动化信息：%s`, x.MustMarshal2String(d), x.MustMarshal2String(s), x.MustMarshal2String(a)),
			},
		})
		if err != nil {
			ava.Error(err)
			resultMessage = "服务器出错了"
			return
		}

		resultMessage = fmt.Sprintf("系统初始化完成，%s", result)
	}()

	return "系统初始化即将开始，数据加载中，即将根据家里的环境和设备通过人工智能自动创建场景和自动化等其他设置，检测到初始化需要至少十分钟的时间，在这段时间内，请不要进行任何操作，直到听到初始化完成，否则可能将造成系统紊乱。"
}
