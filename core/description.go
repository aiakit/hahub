package core

import (
	"fmt"
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"

	"github.com/aiakit/ava"
)

// 展示，支持暂停，继续，停止，循环
// 0.介绍灯光
// 1.简单播报情况，设备类型，设备数量，场景，自动化
// 2.逐个介绍场景和自动化
func Evaluate(message, aiMessage, deviceId string) string {
	//智能家居评估C,B,A,S,SS,SSS，热水器，空调，灯光，扫地机，地暖，插座，电视，开关，浴霸，人体传感器，水浸传感器，烟雾报警器，燃气报警器，门锁，门，窗帘，床，洗衣机，冰箱，洗碗机，温度，湿度,等等进行综合评估
	//智能家居建议
	//设备数量，开启情况
	//场景数量介绍,大致有哪些场景
	//自动化情况介绍，大致有哪些自动化
	device := data.GetDevice()
	scene := data.GetEntityCategoryMap()[data.CategoryScript]
	auto := data.GetEntityCategoryMap()[data.CategoryAutomation]
	result, err := chatCompletionInternal([]*chat.ChatMessage{
		{
			Role: "system",
			Content: fmt.Sprintf(`你是一个智能家居专家，现在你需要根据当前智能家居情况进行人性化的描述，需要突出重点，然后对智能家居等级进行评估，并给出建议，按照C,B,A,S,SS,SSS等级进行评估。
当前设备信息:%s
当前场景信息:%s
当前自动化信息：%s`, x.MustMarshalEscape2String(device), x.MustMarshalEscape2String(scene), x.MustMarshalEscape2String(auto)),
		},
	})
	if err != nil {
		ava.Error(err)
		return "服务器出错了"
	}

	return result
}

//灯光展示,按照编号依次开灯
