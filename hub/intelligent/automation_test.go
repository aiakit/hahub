package intelligent

import (
	"fmt"
	"hahub/hub/data"
	"testing"

	"github.com/aiakit/ava"
)

func init() {
	data.WaitForInit()
}

func TestDelAllAutomation(t *testing.T) {
	DeleteAllAutomations(ava.Background())
}

func TestInitEntityIdByLx(t *testing.T) {
	initEntityIdByLx(ava.Background())
	fmt.Println(x.MustMarshal2String(lxByAreaId))
}

func TestTunOn(t *testing.T) {
	//err := TurnOnAutomation(ava.Background(), "automation.li_jia_zi_dong_hua")
	err := TurnOnAutomation(ava.Background(), "automation.xi_yi_fang_lou_shui")
	if err != nil {
		t.Error(err)
	}

}

func TestSpeaker(t *testing.T) {
	//go func() {
	//	time.Sleep(time.Second * 2)
	//	pausePlay("media_player.xiaomi_cn_865393253_lx06")
	//}()
	ChaosSpeaker()
	//PlayControlAction("cd5be8092557a0dcd20162114ad99de3", "text.xiaomi_cn_865393253_lx06_execute_text_directive_a_5_5", "今天天气怎么样")

	select {}
}
