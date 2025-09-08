package intelligent

import (
	"fmt"
	"hahub/data"
	"hahub/x"
	"strings"
	"sync"
	"time"

	"github.com/aiakit/ava"
	"github.com/panjf2000/ants/v2"
)

var prefixUrlCreateScript = "%s/api/config/script/config/%s"

func ScriptChaos() {

	var now = time.Now()

	c := ava.Background()
	//删除所有场景
	DeleteAllScript(c)
	ava.Debug("场景已清除")

	//初始化开关选择：场景按键，开关按键类型
	InitSwitchSelect(c)
	ava.Debug("开关按键初始化场景已创建")

	//灯光初始化
	InitLight(c)
	ava.Debug("灯光初始化场景已创建")

	//初始化开关
	InitSwitch(c)
	ava.Debug("开关初始化场景已创建")

	//馨光灯初始化
	InitXinGuang(c)
	ava.Debug("预处理馨光场景")

	//初始化灯光场景
	LightScriptSetting(c)
	ava.Debug("创建灯光场景")
	//卧室睡觉场景
	GoodNightScript(c)
	ava.Debug("创建睡觉/晚安场景")
	//起床场景
	GoodMorningScript(c)
	ava.Debug("创建起床/早安场景")
	//调光场景
	dimmmingIncrease(c)
	dimmmingReduce(c)
	ava.Debug("创建调光场景")

	//客厅灯光按照序号开启
	Display(c)
	ava.Debug("创建灯光展示场景")

	//回家离家
	InitHoming(c)
	InitLevingHome(c)
	ava.Debug("创建回家/离家场景")
	//洗澡
	TakeAShower(c)

	//区域直接开关灯
	Panel(c)

	//sos
	Sos(c)

	CreateScript(ava.Background())

	ava.Debugf("latency=%.2f |all script created done! |total=%d", time.Since(now).Seconds(), len(scripts))

	scripts = make([]*Script, 0, 10)

	//刷新实体
	data.CallService().WaitForCallService()

	//switchRule()
}

// 场景，Sequence和automation的actions一致
type Script struct {
	id          string
	Alias       string        `json:"alias"`       //自动化名称
	Description string        `json:"description"` //自动化描述
	Sequence    []interface{} `json:"sequence"`    //执行动作
}

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

func CreateScript(c *ava.Context) {
	var pool, _ = ants.NewPool(8)
	var wg sync.WaitGroup

	for _, script := range scripts {
		wg.Add(1)
		scriptItem := script // 解决闭包问题，创建局部变量

		// 提交任务到协程池
		_ = pool.Submit(func() {
			defer wg.Done()

			// 检查是否已存在同名场景
			arealdy, ok := data.GetEntityCategoryMap()[data.CategoryScript]
			if ok {
				for _, v := range arealdy {
					if v.UniqueID == scriptItem.id && (strings.Contains(v.OriginalName, "*") || strings.Contains(v.OriginalName, "*")) {
						return
					}
				}
			}

			var response Response
			err := x.Post(c, fmt.Sprintf(prefixUrlCreateScript, data.GetHassUrl(), scriptItem.id), data.GetToken(), scriptItem, &response)
			if err != nil {
				ava.Error(err)
				return
			}
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 释放协程池资源
	pool.Release()
}

func AddScript2Queue(c *ava.Context, script *Script) string {

	// 自动化名称和实体ID检测，确保唯一
	baseEntityId := x.ChineseToPinyin(script.Alias)

	script.id = baseEntityId

	scripts = append(scripts, script)
	return "script." + baseEntityId
}

// 删除所有场景
func DeleteAllScript(c *ava.Context) {
	entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
	if !ok {
		return
	}

	var pool, _ = ants.NewPool(8)
	var wg sync.WaitGroup

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		if strings.Contains(entity.OriginalName, "*") {
			continue
		}

		wg.Add(1)
		entityItem := entity // 解决闭包问题，创建局部变量

		// 提交任务到协程池
		_ = pool.Submit(func() {
			defer wg.Done()

			url := fmt.Sprintf(prefixUrlCreateScript, data.GetHassUrl(), entityItem.UniqueID)
			var response Response
			err := x.Del(c, url, data.GetToken(), &response)
			if response.Result != "ok" || err != nil {
				c.Debugf("delete scene |response=%v |id=%s |err=%v", &response, x.MustMarshal2String(entityItem), err)
				return
			}
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 释放协程池资源
	pool.Release()
}
