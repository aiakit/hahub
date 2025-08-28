package data

import (
	"fmt"
	"hahub/x"
	"time"

	"github.com/aiakit/ava"
)

var prefixUrl = "%s/api/states/%s"
var prefixUrlAll = "%s/api/states"

// 获取设备状态
func GetState(entityId string) (*State, error) {
	var response State
	var c = ava.Background()
	err := x.Get(c, fmt.Sprintf(prefixUrl, GetHassUrl(), entityId), GetToken(), &response)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return &response, err
}

type StateAll struct {
	DeviceName  string                 `json:"device_name,omitempty"`
	EntityID    string                 `json:"entity_id,omitempty"`
	State       string                 `json:"State"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
}

func GetStates() ([]StateAll, error) {
	var response []StateAll
	var c = ava.Background()
	err := x.Get(c, fmt.Sprintf(prefixUrlAll, GetHassUrl()), GetToken(), &response)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return response, err
}
