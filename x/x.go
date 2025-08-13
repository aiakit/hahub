package x

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"
	"unsafe"

	"github.com/RussellLuo/timingwheel"
	jsoniter "github.com/json-iterator/go"
	"github.com/mozillazg/go-pinyin"
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

var defaultPinyinArgs = pinyin.NewArgs()

// 中文转拼音，拼音之间用_连接，英文或数字照常拼接
func ChineseToPinyin(s string) string {
	var result []string
	var prevPos int
	runes := []rune(s)
	for i, r := range runes {
		if unicode.Is(unicode.Han, r) {
			// 处理之前的非汉字部分
			if i > prevPos {
				nonHanziPart := filterNonAlphaNumeric(string(runes[prevPos:i]))
				if nonHanziPart != "" {
					result = append(result, nonHanziPart)
				}
			}
			// 处理汉字部分
			py := pinyin.Pinyin(string(r), defaultPinyinArgs)
			result = append(result, py[0][0])
			prevPos = i + 1
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// 处理英文或数字部分
			continue
		} else {
			// 处理其他非汉字、非英文、非数字的部分
			if i > prevPos {
				result = append(result, string(runes[prevPos:i]))
			}
			prevPos = i + 1
		}
	}
	// 处理最后的非汉字部分
	if prevPos < len(runes) {
		nonHanziPart := filterNonAlphaNumeric(string(runes[prevPos:]))
		if nonHanziPart != "" {
			result = append(result, nonHanziPart)
		}
	}
	return ToLower(strings.Join(result, "_"))
}

func filterNonAlphaNumeric(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func IsAllDigits(str string) bool {
	return strings.IndexFunc(str, func(r rune) bool {
		return !unicode.IsDigit(r)
	}) == -1
}
func ConvertBrightnessToPercentage(brightness int) int {
	if brightness <= 0 {
		return 0
	} else if brightness >= 255 {
		return 100
	}

	return int(float64(brightness)/255.0*100.0 + 0.5)
}

var rRand = rand.New(rand.NewSource(time.Now().UnixNano()))

var randLock sync.Mutex

func Intn(n int) int {
	randLock.Lock()
	r := rRand.Intn(n)
	randLock.Unlock()
	return r
}

func ToLower(s string) string {
	return strings.ToLower(s)
}
