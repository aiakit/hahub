package core

// 智能家居描述
func Description(message, aiMessage, deviceId string) string {
	//智能家居评估C,B,A,S,SS,SSS，热水器，空调，灯光，扫地机，地暖，插座，电视，开关，浴霸，人体传感器，水浸传感器，烟雾报警器，燃气报警器，门锁，门，窗帘，床，洗衣机，冰箱，洗碗机，温度，湿度,等等进行综合评估
	//智能家居建议
	//设备数量，开启情况
	//场景数量介绍,大致有哪些场景
	//自动化情况介绍，大致有哪些自动化
	return ""
}

// 展示，支持暂停，继续，停止，循环
// 0.介绍灯光
// 1.简单播报情况，设备类型，设备数量，场景，自动化
// 2.逐个介绍场景和自动化
func Display(message, aiMessage, deviceId string) string {
	return ""
}
