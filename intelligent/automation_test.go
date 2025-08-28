package intelligent

import (
	"fmt"
	"hahub/x"
	"testing"

	"github.com/aiakit/ava"
)

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
