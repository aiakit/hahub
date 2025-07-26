package automation

import (
	"fmt"
	"hahub/hub/core"
	"strings"
	"sync"

	"github.com/aiakit/ava"
)

var prefixUrlCreateAutomation = "%s/api/config/automation/config/%s"
var prefixUrlTurnOnAutomation = "%s/api/services/automation/turn_on"
var prefixUrlTurnOffAutomation = "%s/api/services/automation/turn_off"

type lxConfig struct {
	name string
	Lx   float64
	l    *lx
}

// 区域流明配置
var lxAreaConfig = []lxConfig{
	{"卫生间", 60, nil},
	{"浴室", 61, nil},
	{"洗手盆", 49, nil},
	{"厨房", 99, nil},
	{"餐厅", 98, nil},
	{"书房", 97, nil},
	{"电竞房", 96, nil},
	{"办公室", 95, nil},
	{"工作", 94, nil},
	{"玄关", 92, nil},
	{"茶室", 100, nil},
	{"招待", 101, nil},
	{"会客", 102, nil},
	{"阳台", 91, nil},
	{"客厅", 93, nil},
	{"卧室", 103, nil},
	{"主卧", 104, nil},
	{"次卧", 105, nil},
	{"小孩房", 106, nil},
	{"老人房", 107, nil},
	{"客房", 108, nil},
	{"厢房", 109, nil},
	{"儿童房", 110, nil},
}

type lx struct {
	EntityId string
	Lx       float64
	AreaName string
	ByArea   string
}

var (
	lxByAreaId = make(map[string]*lx, 2)
	lxLock     sync.RWMutex // 新增：读写锁
)

// 自动化配置
type Automation struct {
	Alias       string        `json:"alias"`                //自动化名称
	Description string        `json:"description"`          //自动化描述
	Triggers    []Triggers    `json:"triggers"`             //触发条件
	Conditions  []Conditions  `json:"conditions,omitempty"` //限制条件
	Actions     []interface{} `json:"actions,omitempty"`    //执行动作
	Mode        string        `json:"mode"`                 //执行模式
}

type IfThenELSEAction struct {
	If   []ifCondition `json:"if"`
	Then []interface{} `json:"then"`
	Else []interface{} `json:"else,omitempty"`
}

type ifCondition struct {
	Condition  string        `json:"condition"`
	Conditions []interface{} `json:"conditions"`
}

// 获取 lxByAreaId 中的值，使用读锁
func getLxConfig(areaId string) *lx {
	lxLock.RLock()
	defer lxLock.RUnlock()
	return lxByAreaId[areaId]
}

// 初始化实体ID与流明配置的映射关系，使用写锁
func initEntityIdByLx(c *ava.Context) {
	lxLock.Lock()
	defer lxLock.Unlock()

	for _, areaId := range core.GetAreas() {
		e, ok := core.LxArea[areaId]
		if !ok {
			continue
		}
		for k, config := range lxAreaConfig { // 遍历lxConfig切片
			areaName := core.SpiltAreaName(e.AreaName)
			if strings.Contains(areaName, config.name) {
				data := &lx{
					EntityId: e.EntityID,
					Lx:       config.Lx,
					AreaName: e.AreaName,
					ByArea:   e.AreaName,
				}
				lxByAreaId[areaId] = data // 给lxByAreaId赋值
				lxAreaConfig[k].l = data
				break
			}
		}
	}

	for _, areaId := range core.GetAreas() {
		if v := lxByAreaId[areaId]; v == nil {
			// 查找当前区域在 lxAreaConfig 中的索引位置
			var startIdx int
			var matched bool
			var lux float64
			for idx, config := range lxAreaConfig {
				areaName := core.SpiltAreaName(core.GetAreaName(areaId))
				if strings.Contains(areaName, config.name) {
					startIdx = idx
					matched = true
					lux = config.Lx
					break
				}
			}
			if !matched {
				continue // 无匹配配置项，跳过
			}

			// 从下一个配置项开始循环查找
			tries := 0
			maxTries := len(lxAreaConfig)
			currentIdx := (startIdx + 1) % len(lxAreaConfig)

			for tries < maxTries {
				currentConfig := lxAreaConfig[currentIdx]
				if currentConfig.l != nil {
					lxByAreaId[areaId] = &lx{
						EntityId: currentConfig.l.EntityId,
						Lx:       lux,
						AreaName: core.GetAreaName(areaId),
						ByArea:   currentConfig.l.AreaName,
					}
					break
				}
				currentIdx = (currentIdx + 1) % len(lxAreaConfig)
				tries++
			}

			if tries == maxTries {
				// 所有配置项均无有效设备
				c.Debugf("未找到区域 %s 的有效流明设备", areaId)
			}
		}
	}
}

type Triggers struct {
	Condition string      `json:"condition,omitempty"`
	Type      string      `json:"type,omitempty"`
	DeviceID  string      `json:"device_id,omitempty"`
	EntityID  string      `json:"entity_id,omitempty"`
	State     interface{} `json:"state,omitempty"`
	Domain    string      `json:"domain,omitempty"`
	Trigger   string      `json:"trigger,omitempty"`
	Attribute string      `json:"attribute,omitempty"`
	Above     float64     `json:"above,omitempty"`
	Below     float64     `json:"below,omitempty"`
	For       *For        `json:"for,omitempty"`
}

type Conditions struct {
	Attribute      string        `json:"attribute,omitempty"`
	Condition      string        `json:"condition,omitempty"`
	Type           string        `json:"type,omitempty"`
	DeviceID       string        `json:"device_id,omitempty"`
	EntityID       string        `json:"entity_id,omitempty"`
	Domain         string        `json:"domain,omitempty"`
	Above          float64       `json:"above,omitempty"` //大于
	Below          float64       `json:"below,omitempty"` //小于
	For            *For          `json:"for,omitempty"`
	After          string        `json:"after,omitempty"`
	Before         string        `json:"before,omitempty"`
	Weekday        []string      `json:"weekday,omitempty"`
	ConditionChild []interface{} `json:"conditions,omitempty"`
	State          interface{}   `json:"state,omitempty"`
}

type For struct {
	Hours   float64 `json:"hours"`
	Minutes float64 `json:"minutes"`
	Seconds float64 `json:"seconds"`
}

type ActionLight struct {
	Type          string           `json:"type,omitempty"`
	Action        string           `json:"action,omitempty"`
	DeviceID      string           `json:"device_id,omitempty"`
	EntityID      string           `json:"entity_id,omitempty"`
	Domain        string           `json:"domain,omitempty"`
	BrightnessPct float64          `json:"brightness_pct,omitempty"`
	Data          *actionLightData `json:"data,omitempty"`
	Target        *targetLightData `json:"target,omitempty"`
	Option        string           `json:"option,omitempty"`
}

type actionLightData struct {
	ColorTempKelvin int     `json:"color_temp_kelvin,omitempty"`
	BrightnessPct   float64 `json:"brightness_pct,omitempty"`
	RgbColor        []int   `json:"rgb_color,omitempty"`
}

type targetLightData struct {
	DeviceId string `json:"device_id"`
}

type ActionCommon struct {
	Type     string `json:"type,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
	EntityID string `json:"entity_id,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type ActionService struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data,omitempty"`
	Target *struct {
		EntityId string `json:"entity_id"`
	} `json:"target,omitempty"`
}

type ActionNotify struct {
	Action string `json:"action,omitempty"`
	Data   struct {
		Message string `json:"message,omitempty"`
		Title   string `json:"title,omitempty"`
	} `json:"data,omitempty"`
	Target struct {
		DeviceID string `json:"device_id,omitempty"`
	} `json:"target,omitempty"`
}

type ActionTimerDelay struct {
	Delay struct {
		Hours        int `json:"hours"`
		Minutes      int `json:"minutes"`
		Seconds      int `json:"seconds"`
		Milliseconds int `json:"milliseconds"`
	} `json:"delay"`
}

type Response struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

var deleteAllAutomationSwitch = true

func Chaos() {
	c := ava.Background()

	//删除所有自动化
	if deleteAllAutomationSwitch {
		DeleteAllAutomations(c)
	}

	//处理光照数据
	initEntityIdByLx(ava.Background())

	//人体传感器
	walkBodySensor(c)

	//人体存在传感器
	walkPresenceSensor(c)
	walkPresenceSensorKeting(c)

	//布防
	arm(c)

	//警报
	attention(c)

	//回家
	initHoming(c)

	//灯光控制
	lightControl(c)

	//重新缓存一遍数据
	core.CallService()

	//开关自动关闭规则
	//switchRule()
}

// 发起自动化创建
// 在所有homeassistant自动化名称中，不能出现名称一样的自动化
var skipExistAutomation = false //是否跳过相同名称自动化
var coverExistAutomation = true //是否覆盖名称相关自动化
func CreateAutomation(c *ava.Context, automation *Automation) {
	// 自动化名称和实体ID检测，确保唯一
	alias := automation.Alias
	entityMap := core.GetEntityIdMap()
	baseEntityId := "automation." + core.ChineseToPinyin(alias)
	conflictCount := 0

	for _, entity := range entityMap {
		if entity == nil {
			continue
		}
		if !strings.HasPrefix(entity.EntityID, "automation.") {
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
		automation.Alias = finalAlias
	}

	var response Response
	err := core.Post(c, fmt.Sprintf(prefixUrlCreateAutomation, core.GetHassUrl(), finalEntityId), core.GetToken(), automation, &response)
	if err != nil {
		c.Error(err)
		return
	}

	if response.Result != "ok" {
		c.Errorf("data=%v", core.MustMarshal2String(automation))
	}

	if strings.Contains(automation.Alias, "布防") || strings.Contains(automation.Alias, "撤防") {
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
	err := core.Post(c, fmt.Sprintf(prefixUrlTurnOnAutomation, core.GetHassUrl()), core.GetToken(), &core.HttpServiceData{EntityId: entityId}, &response)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}

func TurnOffAutomation(c *ava.Context, entityId string) error {
	var response Response
	err := core.Post(c, fmt.Sprintf(prefixUrlTurnOffAutomation, core.GetHassUrl()), core.GetToken(), &core.HttpServiceData{EntityId: entityId}, &response)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}

// 删除所有自动化
func DeleteAllAutomations(c *ava.Context) {
	entityMap := core.GetEntityIdMap()
	for _, entity := range entityMap {
		if entity == nil {
			continue
		}
		if entity.Category != core.CategoryAutomation {
			continue
		}
		if core.IsAllDigits(entity.EntityID) {
			continue
		}
		//url := fmt.Sprintf(prefixUrlCreateAutomation, core.GetHassUrl(), entity.EntityID)，ha这个id生成规则有bug
		url := fmt.Sprintf(prefixUrlCreateAutomation, core.GetHassUrl(), entity.UniqueID)
		var response Response
		err := core.Del(c, url, core.GetToken(), &response)
		if response.Result != "ok" || err != nil {
			c.Debugf("delete automation |response=%v |id=%s |err=%v", &response, core.MustMarshal2String(entity), err)
			continue
		}
	}
}
