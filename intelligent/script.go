package intelligent

import (
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

var prefixUrlCreateScript = "%s/api/config/script/config/%s"

func ScriptChaos() {
	c := ava.Background()
	//删除所有场景
	DeleteAllScript(c)

	//初始化开关选择：场景按键，开关按键类型
	InitSwitchSelect(c)
	//灯光初始化
	InitLight(c)

	//初始化开关
	InitSwitch(c)

	//馨光灯初始化
	InitXinGuang(c)
	//初始化灯光场景
	lightScriptSetting(c)
	//卧室睡觉场景
	goodNightScript(c)
	//起床场景
	goodMorningScript(c)

	//调光场景
	dimmmingIncrease(c)
	dimmmingReduce(c)

	//回家离家
	initHoming(c)
	initLevingHome(c)

	//洗澡
	TakeAShower(c)

	//刷新实体
	data.CallService()

	ava.Debugf("all script created done! |total=%d", scriptCount)
}

// 场景，Sequence和automation的actions一致
type Script struct {
	Alias       string        `json:"alias"`       //自动化名称
	Description string        `json:"description"` //自动化描述
	Sequence    []interface{} `json:"sequence"`    //执行动作
}

var scriptLock sync.Mutex
var scriptCount int

func RunAutomation(entityId string) error {
	err := x.Post(ava.Background(), fmt.Sprintf("%s/api/services/automation/trigger", data.GetHassUrl()), data.GetToken(), data.HttpServiceData{
		EntityId: entityId,
	}, nil)
	return err
}

func RunSript(entityId string) error {
	err := x.Post(ava.Background(), fmt.Sprintf("%s/api/services/script/turn_on", data.GetHassUrl()), data.GetToken(), data.HttpServiceData{
		EntityId: entityId,
	}, nil)
	return err
}

func GetScript(uniqueId string, v interface{}) error {
	err := x.Get(ava.Background(), fmt.Sprintf(prefixUrlCreateScript, data.GetHassUrl(), uniqueId), data.GetToken(), v)
	if err != nil {
		return err
	}
	return nil
}

func GetAutomation(uniqueId string, v interface{}) error {
	err := x.Get(ava.Background(), fmt.Sprintf(prefixUrlCreateAutomation, data.GetHassUrl(), uniqueId), data.GetToken(), v)
	if err != nil {
		return err
	}
	return nil
}

func CreateScript(c *ava.Context, script *Script) string {
	scriptLock.Lock()
	defer func() {
		time.Sleep(time.Millisecond * 3)
		scriptLock.Unlock()
	}()

	// 自动化名称和实体ID检测，确保唯一
	entityMap := data.GetEntityIdMap()
	baseEntityId := x.ChineseToPinyin(script.Alias)
	//id := strconv.FormatInt(time.Now().UnixMilli(), 10)

	//script.Alias += "_" + id
	for _, entity := range entityMap {
		if entity == nil {
			continue
		}

	}

	var response Response
	err := x.Post(c, fmt.Sprintf(prefixUrlCreateScript, data.GetHassUrl(), baseEntityId), data.GetToken(), script, &response)
	if err != nil {
		c.Error(err)
		return ""
	}

	if response.Result != "ok" {
		c.Errorf("data=%v |result=%s", x.MustMarshal2String(script), x.MustMarshal2String(&response))
	}

	scriptCount++

	return baseEntityId
}

// 删除所有场景
func DeleteAllScript(c *ava.Context) {
	entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
	if !ok {
		return
	}
	for _, entity := range entities {
		if entity == nil {
			continue
		}

		if strings.Contains(entity.OriginalName, "*") {
			continue
		}

		url := fmt.Sprintf(prefixUrlCreateScript, data.GetHassUrl(), entity.UniqueID)
		var response Response
		err := x.Del(c, url, data.GetToken(), &response)
		if response.Result != "ok" || err != nil {
			c.Debugf("delete scene |response=%v |id=%s |err=%v", &response, x.MustMarshal2String(entity), err)
			continue
		}
	}
}
