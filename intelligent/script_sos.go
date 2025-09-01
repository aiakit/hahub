package intelligent

import "github.com/aiakit/ava"

func Sos(c *ava.Context) {
	var script = &Script{
		Alias:       "sos",
		Description: "触发sos求救事件",
	}
	script.Sequence = append(script.Sequence, virtualEventNotify("sos"))
	script.Sequence = append(script.Sequence, persistentNotification("sos", "有紧急求救事件发生"))
	AddScript2Queue(c, script)
}
