package automation

import (
	"fmt"
	"hahub/hub/internal"
	"strings"

	"github.com/aiakit/ava"
)

var prefixUrlCreateAutomation = "%s/api/config/automation/config/%s"
var prefixUrlTurnOnAutomation = "%s/api/services/automation/turn_on"
var prefixUrlTurnOffAutomation = "%s/api/services/automation/turn_off"

// 自动化配置
type Automation struct {
	Alias       string       `json:"alias"`       //自动化名称
	Description string       `json:"description"` //自动化描述
	Triggers    []Triggers   `json:"triggers"`    //触发条件
	Conditions  []Conditions `json:"conditions"`  //限制条件
	Actions     []Actions    `json:"actions"`     //执行动作
	Mode        string       `json:"mode"`        //执行模式
}

type Triggers struct {
	Type     string `json:"type"`
	DeviceID string `json:"device_id"`
	EntityID string `json:"entity_id"`
	Domain   string `json:"domain"`
	Trigger  string `json:"trigger"`
}

type Conditions struct {
	Condition string `json:"condition"`
	Type      string `json:"type"`
	DeviceID  string `json:"device_id"`
	EntityID  string `json:"entity_id"`
	Domain    string `json:"domain"`
}

type Actions struct {
	Type          string  `json:"type"`
	DeviceID      string  `json:"device_id"`
	EntityID      string  `json:"entity_id"`
	Domain        string  `json:"domain"`
	BrightnessPct float64 `json:"brightness_pct,omitempty"`
}

type Response struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

func Chaos() {
	//创建自动化
	c := ava.Background()
	//人体存在传感器
	walkPresenceSensor(c)

	//重新缓存一遍数据
	internal.CallService()
}

// 发起自动化创建
// 在所有homeassistant自动化名称中，不能出现名称一样的自动化
// var skipExistAutomation = false //是否跳过相同名称自动化
// var coverExistAutomation = true //是否覆盖名称相关自动化
func CreateAutomation(c *ava.Context, automation *Automation, skip, cover bool) {
	// 自动化名称和实体ID检测，确保唯一
	alias := automation.Alias
	entityMap := internal.GetEntityIdMap()
	baseEntityId := "automation." + internal.ChineseToPinyin(alias)
	conflictCount := 0

	for _, entity := range entityMap {
		if entity == nil {
			continue
		}
		if !strings.HasPrefix(entity.EntityID, "automation.") {
			continue
		}

		if entity.OriginalName == alias && skip { //名称相同则不创建
			return
		}

		// 名称冲突
		if entity.OriginalName == alias || entity.EntityID == baseEntityId {
			if cover { //直接覆盖
				continue
			}
			conflictCount++ //重新建一个
		}
	}

	finalAlias := alias
	finalEntityId := baseEntityId
	if conflictCount > 0 {
		finalAlias = fmt.Sprintf("%s%d", alias, conflictCount)
		finalEntityId = fmt.Sprintf("%s%d", baseEntityId, conflictCount)
		automation.Alias = finalAlias
	}

	var response Response
	err := internal.Post(c, fmt.Sprintf(prefixUrlCreateAutomation, internal.GetHassUrl(), finalEntityId), internal.GetToken(), automation, &response)
	if err != nil {
		c.Error(err)
		return
	}

	err = TurnOnAutomation(c, finalEntityId)
	if err != nil {
		c.Error(err)
		return
	}
}

func TurnOnAutomation(c *ava.Context, entityId string) error {
	var response Response
	err := internal.Post(c, fmt.Sprintf(prefixUrlTurnOnAutomation, internal.GetHassUrl()), internal.GetToken(), &internal.HttpServiceData{EntityId: entityId}, &response)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}
