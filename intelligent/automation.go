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
	{"卧室", 50, nil},
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
	id          string
	Alias       string        `json:"alias"`                //自动化名称
	Description string        `json:"description"`          //自动化描述
	Triggers    []*Triggers   `json:"triggers"`             //触发条件
	Conditions  []*Conditions `json:"conditions,omitempty"` //限制条件
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
	Conditions []interface{} `json:"conditions,omitempty"`
	State      string        `json:"state,omitempty"`
	EntityId   string        `json:"entity_id,omitempty"`
	Type       string        `json:"type,omitempty"`
	DeviceId   string        `json:"device_id,omitempty"`
	Domain     string        `json:"domain,omitempty"`
	Above      float64       `json:"above,omitempty"`
	Below      float64       `json:"below,omitempty"`
	After      string        `json:"after,omitempty"`
	Before     string        `json:"before,omitempty"`
	Attribute  string        `json:"attribute,omitempty"`
	Weekday    []string      `json:"weekday,omitempty"`
	For        *For          `json:"for,omitempty"`
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

	for _, areaId := range data.GetAreas() {
		e, ok := data.LxArea[areaId]
		if !ok {
			continue
		}
		for k, config := range lxAreaConfig { // 遍历lxConfig切片
			areaName := data.SpiltAreaName(e.AreaName)
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

	for _, areaId := range data.GetAreas() {
		if v := lxByAreaId[areaId]; v == nil {
			// 查找当前区域在 lxAreaConfig 中的索引位置
			var startIdx int
			var matched bool
			var lux float64
			for idx, config := range lxAreaConfig {
				areaName := data.SpiltAreaName(data.GetAreaName(areaId))
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
						AreaName: data.SpiltAreaName(data.GetAreaName(areaId)),
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
	Name      string      `json:"name,omitempty"`
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
	Name           string        `json:"name,omitempty"` //设备名称
}

type For struct {
	Hours   float64 `json:"hours"`
	Minutes float64 `json:"minutes"`
	Seconds float64 `json:"seconds"`
}

type ActionLight struct {
	Condition     string           `json:"condition,omitempty"`
	Type          string           `json:"type,omitempty"`
	Action        string           `json:"action,omitempty"`
	DeviceID      string           `json:"device_id,omitempty"`
	EntityID      string           `json:"entity_id,omitempty"`
	Domain        string           `json:"domain,omitempty"`
	BrightnessPct float64          `json:"brightness_pct,omitempty"`
	Data          *actionLightData `json:"data,omitempty"`
	Target        *targetLightData `json:"target,omitempty"`
	Option        string           `json:"option,omitempty"`
	Delay         *delay           `json:"delay,omitempty"`
	Value         interface{}      `json:"value,omitempty"`

	subCategory string
}

type delay struct {
	Hours        int `json:"hours"`
	Minutes      int `json:"minutes"`
	Seconds      int `json:"seconds"`
	Milliseconds int `json:"milliseconds"`
}

type actionLightData struct {
	ColorTempKelvin   int         `json:"color_temp_kelvin,omitempty"`
	BrightnessPct     float64     `json:"brightness_pct,omitempty"`
	BrightnessStepPct float64     `json:"brightness_step_pct,omitempty"`
	RgbColor          []int       `json:"rgb_color,omitempty"`
	State             interface{} `json:"state,omitempty"`
}

type targetLightData struct {
	DeviceId string `json:"device_id,omitempty"`
	EntityId string `json:"entity_id,omitempty"`
}

type ActionCommon struct {
	Type     string      `json:"type,omitempty"`
	DeviceID string      `json:"device_id,omitempty"`
	EntityID string      `json:"entity_id,omitempty"`
	Domain   string      `json:"domain,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	Option   interface{} `json:"option,omitempty"`
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
	Delay *delay `json:"delay"`
}

type Response struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

func ChaosAutomation() {

	var now = time.Now()

	c := ava.Background()

	//删除所有自动化
	DeleteAllAutomations(c)
	ava.Debug("清除所有系统创建的自动化")
	//处理光照数据
	initEntityIdByLx(ava.Background())
	ava.Debug("光照数据已经处理")

	//人体传感器
	WalkBodySensor(c)
	ava.Debug("创建人体传感器相关自动化")

	//人体存在传感器
	WalkPresenceSensor(c)
	WalkPresenceSensorKeting(c)
	ava.Debug("创建人在传感器相关自动化")

	//布防
	Defense(c)

	//警报
	attention(c)

	ava.Debug("创建布防和警报自动化")

	//灯光控制
	LightControl(c)
	ava.Debug("创建开关控制灯自动化")

	//插座打开就开灯
	WalkBodySocketSensor(c)
	ava.Debug("创建插座灯自动化")

	WalkPresenceSensorAir(c)

	CreateAutomation(ava.Background())

	ava.Debugf("latency=%.2f |all automation created done! |total=%d", time.Since(now).Seconds(), len(autos))

	autos = make([]*Automation, 0, 10)

	//重新缓存一遍数据
	data.CallService().WaitForCallService()

	//开关自动关闭规则
	//switchRule()
}

var autos []*Automation
var scripts []*Script

func Chaos() {
	ScriptChaos()
	ChaosAutomation()
}

// 发起自动化创建
// 在所有homeassistant自动化名称中，不能出现名称一样的自动化

func CreateAutomation(c *ava.Context) {
	var pool, _ = ants.NewPool(8)
	var wg sync.WaitGroup

	for _, auto := range autos {
		wg.Add(1)
		automation := auto // 解决闭包问题，创建局部变量

		// 提交任务到协程池
		_ = pool.Submit(func() {
			defer wg.Done()

			// 检查是否已存在同名自动化
			arealdy, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
			if ok {
				for _, v := range arealdy {
					if v.UniqueID == automation.id && (strings.Contains(v.OriginalName, "*") || strings.Contains(v.Name, "*")) {
						return
					}
				}
			}

			var response Response
			err := x.Post(c, fmt.Sprintf(prefixUrlCreateAutomation, data.GetHassUrl(), automation.id), data.GetToken(), automation, &response)
			if err != nil {
				ava.Error(err)
				return
			}

			// 布防/撤防类自动化不需要开启
			if strings.Contains(automation.Alias, "布防") || strings.Contains(automation.Alias, "撤防") {
				return
			}

			err = TurnOnAutomation(c, automation.id)
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

func AddAutomation2Queue(c *ava.Context, automation *Automation) string {
	// 自动化名称和实体ID检测，确保唯一
	alias := automation.Alias
	baseEntityId := "automation." + x.ChineseToPinyin(alias)

	automation.id = baseEntityId

	autos = append(autos, automation)

	return baseEntityId
}

func TurnOnAutomation(c *ava.Context, entityId string) error {
	var response Response
	err := x.Post(c, fmt.Sprintf(prefixUrlTurnOnAutomation, data.GetHassUrl()), data.GetToken(), &data.HttpServiceData{EntityId: entityId}, &response)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}

func TurnOffAutomation(c *ava.Context, entityId string) error {
	var response Response
	err := x.Post(c, fmt.Sprintf(prefixUrlTurnOffAutomation, data.GetHassUrl()), data.GetToken(), &data.HttpServiceData{EntityId: entityId}, &response)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}

// 删除所有自动化
func DeleteAllAutomations(c *ava.Context) {
	entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
	if !ok {
		return
	}

	var pool, _ = ants.NewPool(8)
	var wg sync.WaitGroup

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		if strings.Contains(entity.OriginalName, "*") || entity.Name != "" {
			continue
		}

		wg.Add(1)
		entityItem := entity // 解决闭包问题，创建局部变量

		// 提交任务到协程池
		_ = pool.Submit(func() {
			defer wg.Done()

			url := fmt.Sprintf(prefixUrlCreateAutomation, data.GetHassUrl(), entityItem.UniqueID)
			var response Response
			err := x.Del(c, url, data.GetToken(), &response)
			if response.Result != "ok" || err != nil {
				c.Debugf("delete automation |response=%v |id=%s |err=%v", &response, x.MustMarshal2String(entityItem), err)
				return
			}
		})
	}

	// 等待所有任务完成
	wg.Wait()

	// 释放协程池资源
	pool.Release()
}

// 注册虚拟事件触发自动化或者脚本
// s_名称，表示场景虚拟化事件,a_名称，表示自动化虚拟化事件
func registerVirtualEvent(simple *data.StateChangedSimple, body []byte) {
	var event virtualEvent
	err := x.Unmarshal(body, &event)
	if err != nil {
		return
	}

	if v, ok := event.Event.Data.NewState.Attributes["event_type"].(string); ok && v == "虚拟事件发生" {
		name, ok := event.Event.Data.NewState.Attributes["事件名称"].(string)
		if !ok {
			return
		}
		//获取所有的场景和自动化
		func() {
			if strings.HasPrefix(name, "场景") {
				name = strings.Trim(name, "场景")
				entities, ok := data.GetEntityCategoryMap()[data.CategoryScript]
				if !ok {
					return
				}

				var index int
				var tmp float64
				for k, entity := range entities {
					s := x.Similarity(name, entity.OriginalName)
					if s > tmp {
						index = k
						tmp = s
					}
				}

				e := entities[index]
				err = RunSript(e.EntityID)
				if err != nil {
					ava.Error(err)
					return
				}

				ava.Debugf("收到虚拟事件发生=%s |执行场景=%s", name, e.OriginalName)
			}
		}()

		//获取所有的场景和自动化
		func() {
			if strings.HasPrefix(name, "自动化") {
				name = strings.Trim(name, "自动化")

				entities, ok := data.GetEntityCategoryMap()[data.CategoryAutomation]
				if !ok {
					return
				}

				var index int
				var tmp float64
				for k, entity := range entities {
					s := x.Similarity(name, entity.OriginalName)
					if s > tmp {
						index = k
						tmp = s
					}
				}

				e := entities[index]
				err = RunAutomation(e.EntityID)
				if err != nil {
					ava.Error(err)
					return
				}
				ava.Debugf("收到虚拟事件发生=%s |执行自动化=%s", name, e.OriginalName)
			}
		}()
	}
}

type virtualEvent struct {
	Type  string `json:"type"`
	Event struct {
		EventType string `json:"event_type"`
		Data      struct {
			EntityID string `json:"entity_id"`
			NewState struct {
				EntityID     string                 `json:"entity_id"`
				State        time.Time              `json:"state"`
				Attributes   map[string]interface{} `json:"attributes"`
				LastChanged  time.Time              `json:"last_changed"`
				LastReported time.Time              `json:"last_reported"`
				LastUpdated  time.Time              `json:"last_updated"`
				Context      struct {
					ID       string `json:"id"`
					ParentID any    `json:"parent_id"`
					UserID   any    `json:"user_id"`
				} `json:"context"`
			} `json:"new_state"`
		} `json:"data"`
		Origin    string    `json:"origin"`
		TimeFired time.Time `json:"time_fired"`
		Context   struct {
			ID       string `json:"id"`
			ParentID any    `json:"parent_id"`
			UserID   any    `json:"user_id"`
		} `json:"context"`
	} `json:"event"`
	ID int `json:"id"`
}
