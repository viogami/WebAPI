package openai

import (
	config "WebAPI/conf"
	"context"
	"log/slog"
	"os"
	"sync"

	goOpenai "github.com/sashabaranov/go-openai"
)

var (
	instance *ChatGPTService
	once     sync.Once
)

func GetInstance() *ChatGPTService {
	once.Do(func() {
		instance = NewChatGPTService()
	})
	return instance
}

type ChatGPTService struct {
	Role             string
	Character        string
	characterSetting string

	client *goOpenai.Client
}

func NewChatGPTService() *ChatGPTService {
	URL_PROXY := config.AppConfig.AIConfig.ChatGPTUrlProxy
	APIKey := os.Getenv("ChatGPTAPIKey")

	s := new(ChatGPTService)
	s.Role = goOpenai.ChatMessageRoleUser
	s.Character = "vio"
	s.characterSetting = gpt_preset[s.Character]

	conf := goOpenai.DefaultConfig(APIKey)
	conf.BaseURL = URL_PROXY

	s.client = goOpenai.NewClientWithConfig(conf)

	return s
}

func (s *ChatGPTService) SetCharacter(character string) {
	s.Character = character
	s.characterSetting = gpt_preset[character]
}

func (s *ChatGPTService) InvokeChatGPTAPI(text string) string {
	return s.InvokeChatGPTAPIWithRole(text, s.Role)
}

func (s *ChatGPTService) InvokeChatGPTAPIWithRole(text string, role string) string {
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		goOpenai.ChatCompletionRequest{
			Model: goOpenai.GPT4o20241120,
			Messages: []goOpenai.ChatCompletionMessage{
				{
					Role:    goOpenai.ChatMessageRoleSystem,
					Content: s.characterSetting,
				},
				{
					Role:    role,
					Content: text,
				},
			},
		},
	)
	if err != nil {
		slog.Error("Error calling ChatGPT API", "error", err)
		Resp := "AI调用失败了😥 error:\n" + err.Error()
		return Resp
	}
	return resp.Choices[0].Message.Content
}
