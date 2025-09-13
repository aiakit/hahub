package core

import (
	"hahub/data"
	"hahub/internal/chat"
	"hahub/x"
	"sync"
	"time"

	"github.com/aiakit/ava"
)

// 内存记忆功能

// MemoryManager 管理对话历史记录
type MemoryManager struct {
	lock       sync.Mutex
	history    map[string][]*chat.ChatMessage // 按设备ID存储历史记录
	maxHistory int                            // 最大历史记录数
}

var memoryManager *MemoryManager

func init() {
	memoryManager = &MemoryManager{
		history:    make(map[string][]*chat.ChatMessage),
		maxHistory: 5,
	}
}

// GetHistory 获取指定设备的历史记录
func GetHistory(deviceId string) []*chat.ChatMessage {
	memoryManager.lock.Lock()
	defer memoryManager.lock.Unlock()

	history, exists := memoryManager.history[deviceId]
	if !exists {
		return nil
	}
	return history
}

// AddUserMessage 添加用户消息到历史记录
func AddUserMessage(deviceId, content string) {
	addMessage(deviceId, &chat.ChatMessage{
		Role:    "user",
		Content: content,
	})
}

// AddAIMessage 添加AI消息到历史记录
func AddAIMessage(deviceId, content string) {
	addMessage(deviceId, &chat.ChatMessage{
		Role:    "assistant",
		Content: content,
	})
}

// AddSystemMessage 添加系统消息到历史记录
func AddSystemMessage(deviceId, content string) {
	addMessage(deviceId, &chat.ChatMessage{
		Role:    "system",
		Content: content,
	})
}

// addMessage 内部方法，添加消息到历史记录
func addMessage(deviceId string, message *chat.ChatMessage) {
	memoryManager.lock.Lock()
	defer memoryManager.lock.Unlock()

	// 如果该设备的历史记录不存在，则创建
	if _, exists := memoryManager.history[deviceId]; !exists {
		memoryManager.history[deviceId] = make([]*chat.ChatMessage, 0)
	}

	// 如果历史记录已满，移除最旧的记录
	if len(memoryManager.history[deviceId]) >= memoryManager.maxHistory {
		memoryManager.history[deviceId] = memoryManager.history[deviceId][1:]
	}

	// 添加新记录
	memoryManager.history[deviceId] = append(memoryManager.history[deviceId], message)
}

// ClearHistory 清除指定设备的历史记录
func ClearHistory(deviceId string) {
	memoryManager.lock.Lock()
	defer memoryManager.lock.Unlock()

	delete(memoryManager.history, deviceId)
}

// ClearAllHistory 清除所有设备的历史记录
func ClearAllHistory() {
	memoryManager.lock.Lock()
	defer memoryManager.lock.Unlock()

	memoryManager.history = make(map[string][]*chat.ChatMessage)
}

// GetLastUpdated 获取指定设备的最后更新时间

func PlayTextActionByEntityId(entityId, message string) {

	// 使用 []rune 计算字符长度
	runes := []rune(message)
	if len(runes) > 800 {
		// 按800个字符拆分消息
		for i := 0; i < len(runes); i += 800 {
			end := i + 800
			if end > len(runes) {
				end = len(runes)
			}

			// 将切片转换回字符串
			chunk := string(runes[i:end])
			err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
				EntityId: entityId,
				Message:  chunk,
			}, nil)
			if err != nil {
				ava.Error(err)
				continue
			}
			// 暂停，等待播放完成
			time.Sleep(GetPlaybackDuration(chunk))
		}
	} else {
		err := x.Post(ava.Background(), data.GetHassUrl()+"/api/services/notify/send_message", data.GetToken(), &data.HttpServiceData{
			EntityId: entityId,
			Message:  message,
		}, nil)
		if err != nil {
			ava.Error(err)
			return
		}
		// 暂停，等待播放完成
		time.Sleep(GetPlaybackDuration(message))
		ava.Debugf("PlayTextAction |text=%s |latency=%v", message, GetPlaybackDuration(message))
	}

	//使用1秒补正
	time.Sleep(time.Second)
}
