package internal

import (
	"fmt"

	"github.com/aiakit/ava"
)

var prefixUrl = "%s/api/states/%s"

// 获取设备状态
func GetState(entityId string) (*State, error) {
	var response State
	var c = ava.Background()
	err := Get(c, fmt.Sprintf(prefixUrl, GetHassUrl(), entityId), GetToken(), &response)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return &response, err
}
