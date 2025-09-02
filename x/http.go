package x

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aiakit/ava"
)

func PostWithOutLog(c *ava.Context, uri, token string, data, v interface{}) error {

	var body = MustMarshalEscape(data)

	var header = map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 添加重试机制，最多重试3次
	var b []byte
	var err error
	for i := 0; i < 3; i++ {
		b, err = post(c, uri, body, header)
		if err == nil {
			break
		}
		c.Debugf("Post请求失败，第%d次重试: %v", i+1, err)
		if i < 2 { // 避免最后一次重试后不必要的延迟
			time.Sleep(time.Duration(i+1) * time.Second) // 逐步增加重试间隔
		}
	}

	if err != nil {
		ava.Error(err)
		return err
	}

	if v == nil {
		return nil
	}

	return Unmarshal(b, v)
}

func Post(c *ava.Context, uri, token string, data, v interface{}) error {
	var now = time.Now()

	var body = MustMarshalEscape(data)

	var header = map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 添加重试机制，最多重试3次
	var b []byte
	var err error
	for i := 0; i < 3; i++ {
		b, err = post(c, uri, body, header)
		if err == nil {
			break
		}
		c.Debugf("Post请求失败，第%d次重试: %v", i+1, err)
		if i < 2 { // 避免最后一次重试后不必要的延迟
			time.Sleep(time.Duration(i+1) * time.Second) // 逐步增加重试间隔
		}
	}

	if err != nil {
		ava.Error(err)
		return err
	}

	if len(string(b)) < 500 {
		c.Debugf("latency=%v秒 |uri=%s |TO=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(body), string(b))
	} else {
		c.Debugf("latency=%v秒 |uri=%s |TO=%v |FROM_LEN=%v", time.Now().Sub(now).Seconds(), uri, string(body), len(string(b)))

	}

	if v == nil {
		return nil
	}

	return Unmarshal(b, v)
}

func Del(c *ava.Context, uri, token string, v interface{}) error {
	//var now = time.Now()

	var header = map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 添加重试机制，最多重试3次
	var b []byte
	var err error
	for i := 0; i < 3; i++ {
		b, err = del(c, uri, nil, header)
		if err == nil {
			break
		}
		c.Debugf("Del请求失败，第%d次重试: %v", i+1, err)
		if i < 2 { // 避免最后一次重试后不必要的延迟
			time.Sleep(time.Duration(i+1) * time.Second) // 逐步增加重试间隔
		}
	}

	if err != nil {
		c.Error(err)
		return err
	}

	//if len(string(b)) < 500 {
	//	c.Debugf("latency=%v秒 |uri=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(b))
	//} else {
	//	c.Debugf("latency=%v秒 |uri=%s |FROM_LEN=%v", time.Now().Sub(now).Seconds(), uri, len(string(b)))
	//}

	if v == nil {
		return nil
	}
	return Unmarshal(b, v)
}

func Get(c *ava.Context, uri, token string, v interface{}) error {
	//var now = time.Now()
	var header = map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 添加重试机制，最多重试3次
	var b []byte
	var err error
	for i := 0; i < 3; i++ {
		b, err = get(c, uri, header)
		if err == nil {
			break
		}
		c.Debugf("Get请求失败，第%d次重试: %v", i+1, err)
		if i < 2 { // 避免最后一次重试后不必要的延迟
			time.Sleep(time.Duration(i+1) * time.Second) // 逐步增加重试间隔
		}
	}

	if err != nil {
		c.Error(err)
		return err
	}

	//if len(string(b)) < 500 {
	//	c.Debugf("latency=%v秒 |uri=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(b))
	//} else {
	//	c.Debugf("latency=%v秒 |uri=%s |FROM_LEN=%v", time.Now().Sub(now).Seconds(), uri, len(string(b)))
	//}

	return Unmarshal(b, v)
}

var (
	Client *http.Client
)

func ClientInstance() *http.Client {
	if Client == nil {
		Client = &http.Client{
			Timeout: 30 * time.Second, // 增加总超时时间
			Transport: &http.Transport{
				DisableKeepAlives: true,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second, // 减少 TCP 连接超时时间
					KeepAlive: 30 * time.Second, // 保持长连接的时间
				}).DialContext,
				MaxIdleConns:          50,
				MaxConnsPerHost:       100,
				MaxIdleConnsPerHost:   50,               // 降低每个 host 的空闲连接数
				ExpectContinueTimeout: 2 * time.Second,  // 等待服务第一响应的超时时间
				IdleConnTimeout:       30 * time.Second, // 降低空闲连接的超时时间
			},
		}
	}
	return Client
}

// CheckRespStatus 状态检查
func CheckRespStatus(resp *http.Response) ([]byte, error) {
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return bodyBytes, nil
	}
	return nil, errors.New(string(bodyBytes))
}

func post(c *ava.Context, url string, data []byte, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

func postForm(
	c *ava.Context,
	url string,
	data url.Values,
	header map[string]string,
) ([]byte, error) {

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

// post 流式
func postFormStreamChan(ctx context.Context, url string, reqData interface{}, HeaderValue http.Header) (chan string, error) {
	cli := &http.Client{
		Timeout: time.Second * 60,
	}

	var reqBody io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, err
	}

	if len(HeaderValue) > 0 {
		for k, v := range HeaderValue {
			req.Header.Set(k, v[0])
		}
	}

	var lines = make(chan string, 4)
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	go func() {
		defer func() {
			resp.Body.Close()
			close(lines)
		}()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				if line == "" {
					continue
				}
				lines <- line
			}

		}
	}()

	return lines, nil
}

func get(c *ava.Context, url string, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

func getWithout(url string, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		ava.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		ava.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		ava.Error(err)
		return nil, err
	}

	return rsp, nil
}

func put(c *ava.Context, url string, data []byte, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

func del(c *ava.Context, url string, data []byte, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := ClientInstance().Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}
