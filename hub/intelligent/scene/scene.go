package scene

import (
	"fmt"
	"hahub/hub/core"
	"strings"

	"github.com/aiakit/ava"
)

var prefixUrlCreateScene = "%s/api/config/scene/config/%s"

// 场景
func Chaos() {
	c := ava.Background()

	//删除所有场景

	InitLight(c)
	InitSwitch(c)

}

type Scene struct {
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Entities map[string]interface{} `json:"entities,omitempty"`
}

type Response struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

var skipExistAutomation = false //是否跳过相同名称自动化
var coverExistAutomation = true //是否覆盖名称相关自动化
func CreateScene(c *ava.Context, scene *Scene) {
	// 自动化名称和实体ID检测，确保唯一
	alias := scene.Name
	entityMap := core.GetEntityIdMap()
	baseEntityId := "scene." + core.ChineseToPinyin(alias)
	conflictCount := 0

	for _, entity := range entityMap {
		if entity == nil {
			continue
		}
		if !strings.HasPrefix(entity.EntityID, "scene.") {
			continue
		}

		if entity.OriginalName == alias && skipExistAutomation { //名称相同则不创建
			return
		}

		// 名称冲突
		if entity.OriginalName == alias || entity.EntityID == baseEntityId {
			if coverExistAutomation { //直接覆盖
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
		scene.Name = finalAlias
	}

	var response Response
	err := core.Post(c, fmt.Sprintf(prefixUrlCreateScene, core.GetHassUrl(), finalEntityId), core.GetToken(), scene, &response)
	if err != nil {
		c.Error(err)
		return
	}

	if response.Result != "ok" {
		c.Error("result=", response)
		c.Errorf("data=%v", core.MustMarshal2String(scene))
	}
}
