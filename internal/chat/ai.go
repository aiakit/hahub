package chat

import (
	"context"
	"hahub/x"

	"github.com/aiakit/ava"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

var DefaultQianwenApiKey = "sk-08cdfea5547040209ea0e2d874fff912"
var DefaultQianwenUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
var DefaultQianwenModel = "qwen-turbo-2025-07-15"

//var DefaultQianwenModel = "qwen-turbo-2024-11-01"

var DefaultProvider = NewOpenAIProvider(DefaultQianwenModel)

//var DefaultProvider = NewDoubaoProvider("doubao-seed-1-6-250615")

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// AIProvider 定义AI服务提供者的接口
type AIProvider interface {
	ChatCompletion(messages []*ChatMessage) (string, error)
}

// OpenAIProvider 实现OpenAI服务提供者
type OpenAIProvider struct {
	client *openai.Client
	m      string
}

func NewOpenAIProvider(m string) *OpenAIProvider {
	config := openai.DefaultConfig(DefaultQianwenApiKey)
	config.BaseURL = DefaultQianwenUrl

	return &OpenAIProvider{client: openai.NewClientWithConfig(config), m: m}
}

func (o *OpenAIProvider) ChatCompletion(messages []*ChatMessage) (string, error) {
	openaiMessages := make([]openai.ChatCompletionMessage, 0, 2)
	for _, msg := range messages {
		mm := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if msg.Name != "" {
			mm.Name = msg.Name
		}

		openaiMessages = append(openaiMessages, mm)
	}

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    o.m,
			Messages: openaiMessages,
		},
	)

	if err != nil {
		ava.Errorf("openai ChatCompletion error=%v", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// DoubaoProvider 实现豆包服务提供者
type DoubaoProvider struct {
	client *arkruntime.Client
	m      string
}

func NewDoubaoProvider(m string) *DoubaoProvider {
	return &DoubaoProvider{client: arkruntime.NewClientWithApiKey("5b60ed7d-71cd-486a-96bd-cb1ad2a9a9d2"), m: m}
}

func (d *DoubaoProvider) ChatCompletion(messages []*ChatMessage) (string, error) {

	openaiMessages := make([]*model.ChatCompletionMessage, 0, 2)
	for _, msg := range messages {
		mm := &model.ChatCompletionMessage{
			Role:    msg.Role,
			Content: &model.ChatCompletionMessageContent{StringValue: volcengine.String(msg.Content)},
		}
		if msg.Name != "" {
			mm.Name = volcengine.String(msg.Name)
		}
		openaiMessages = append(openaiMessages, mm)
	}

	req := model.CreateChatCompletionRequest{
		// 将推理接入点 <Model>替换为 Model ID
		Model:    d.m,
		Messages: openaiMessages,
	}

	// 发送聊天完成请求，并将结果存储在 resp 中，将可能出现的错误存储在 err 中
	resp, err := d.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		// 若出现错误，打印错误信息并终止程序
		ava.Errorf("doubao ChatCompletion error=%v", err)
		return "", err
	}

	return *resp.Choices[0].Message.Content.StringValue, nil
}

// ChatCompletionMessage 通用的聊天完成函数
func ChatCompletionMessage(messages []*ChatMessage) (string, error) {
	s, err := DefaultProvider.ChatCompletion(messages)
	if err != nil {
		ava.Debugf("TO=%s| err=%v", x.MustMarshalEscape2String(messages), err)
		return s, err
	}
	ava.Debugf("TO=%s| FROM=%s", x.MustMarshalEscape2String(messages), s)
	return s, err
}
