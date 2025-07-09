package hub

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
	"unsafe"

	"github.com/RussellLuo/timingwheel"
	jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigFastest

// todo jsoniter有bug
func Unmarshal(b []byte, v interface{}) error {
	_ = Json.Unmarshal(b, v)
	return nil
}

func MustMarshalEscape(v interface{}) []byte {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 设置不转义 HTML

	if err := encoder.Encode(v); err != nil {
		return nil // 返回编码错误
	}

	return buf.Bytes() // 返回解析结果
}

func MustMarshalEscape2String(v interface{}) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 设置不转义 HTML

	if err := encoder.Encode(v); err != nil {
		return "" // 返回编码错误
	}

	return buf.String() // 返回解析结果
}

func MustMarshal(v interface{}) []byte {
	b, _ := Json.Marshal(v)
	return b
}

func MustMarshal2String(v interface{}) string {
	b, _ := Json.Marshal(v)
	return string(b)
}

func StringToBytes(s string) (b []byte) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func FindJSON(str string) string {
	var a, b, endA, endB int
	a = strings.Index(str, "[")
	b = strings.Index(str, "{")

	if a != -1 && b != -1 {
		if a < b {
			endA = strings.LastIndex(str, "]")
			if endA != -1 {
				return str[a : endA+1]
			}
		} else {
			endB = strings.LastIndex(str, "}")
			if endB != -1 {
				return str[b : endB+1]
			}
		}
	} else if a != -1 {
		endA = strings.LastIndex(str, "]")
		if endA != -1 {
			return str[a : endA+1]
		}
	} else if b != -1 {
		endB = strings.LastIndex(str, "}")
		if endB != -1 {
			return str[b : endB+1]
		}
	}

	return ""
}

var tw = timingwheel.NewTimingWheel(time.Second, 20)

func init() {
	tw.Start()
}

type EveryScheduler struct {
	Interval time.Duration
}

func (s *EveryScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

func TimingwheelAfter(t time.Duration, f func()) {
	tw.AfterFunc(t, f)
}

func TimingwheelTicker(t time.Duration, f func()) *timingwheel.Timer {
	return tw.ScheduleFunc(&EveryScheduler{Interval: t}, f)
}
