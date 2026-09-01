package openai

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	goOpenai "github.com/sashabaranov/go-openai"
)

var ErrNotConfigured = errors.New("OpenAI API key is not configured")

type ChatGPTService struct {
	characterSetting string
	client           *goOpenai.Client
	apiKey           string
}

func NewChatGPTService(baseURL, apiKey string) *ChatGPTService {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	apiKey = strings.TrimSpace(apiKey)
	if baseURL == "" {
		return &ChatGPTService{apiKey: apiKey}
	}

	config := goOpenai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	config.HTTPClient = &http.Client{Timeout: 30 * time.Second}

	return &ChatGPTService{
		characterSetting: gpt_preset["vio"],
		client:           goOpenai.NewClientWithConfig(config),
		apiKey:           apiKey,
	}
}

func (s *ChatGPTService) InvokeChatGPTAPI(ctx context.Context, text string) (string, error) {
	if s.apiKey == "" || s.client == nil {
		return "", ErrNotConfigured
	}

	resp, err := s.client.CreateChatCompletion(ctx, goOpenai.ChatCompletionRequest{
		Model: goOpenai.GPT4o20241120,
		Messages: []goOpenai.ChatCompletionMessage{
			{
				Role:    goOpenai.ChatMessageRoleSystem,
				Content: s.characterSetting,
			},
			{
				Role:    goOpenai.ChatMessageRoleUser,
				Content: text,
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", errors.New("OpenAI returned no message content")
	}

	return resp.Choices[0].Message.Content, nil
}
