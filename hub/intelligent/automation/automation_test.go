package automation

import (
	"fmt"
	"hahub/hub/core"
	"testing"

	"github.com/aiakit/ava"
)

func init() {
	core.WaitForInit()
}

func TestDelAllAutomation(t *testing.T) {
	DeleteAllAutomations(ava.Background())
}

func TestInitEntityIdByLx(t *testing.T) {
	initEntityIdByLx(ava.Background())
	fmt.Println(core.MustMarshal2String(lxByAreaId))
}
