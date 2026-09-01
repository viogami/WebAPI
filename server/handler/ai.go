package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"webapi/conf"
	"webapi/core/ai/deepseek"
	"webapi/core/ai/openai"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	chatGPT  *openai.ChatGPTService
	deepSeek *deepseek.DeepSeekService
}

func NewAIHandler(cfg conf.AIConfig) *AIHandler {
	return &AIHandler{
		chatGPT:  openai.NewChatGPTService(cfg.OpenAIBaseURL, cfg.OpenAIAPIKey),
		deepSeek: deepseek.NewDeepSeekService(cfg.DeepSeekBaseURL, cfg.DeepSeekAPIKey),
	}
}

func (h *AIHandler) ProcessMessage(c *gin.Context) {
	message, ok := requestMessage(c)
	if !ok {
		return
	}

	reply, err := h.chatGPT.InvokeChatGPTAPI(c.Request.Context(), message)
	if err != nil {
		writeAIError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func (h *AIHandler) ProcessSharpReviews(c *gin.Context) {
	message, ok := requestMessage(c)
	if !ok {
		return
	}

	reply, err := h.deepSeek.InvokeDeepSeekAPI(c.Request.Context(), message)
	if err != nil {
		writeAIError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func requestMessage(c *gin.Context) (string, bool) {
	message := strings.TrimSpace(c.PostForm("message"))
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return "", false
	}
	return message, true
}

func writeAIError(c *gin.Context, err error) {
	if errors.Is(err, openai.ErrNotConfigured) || errors.Is(err, deepseek.ErrNotConfigured) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service is not configured"})
		return
	}

	slog.Error("AI upstream request failed", "error", err)
	c.JSON(http.StatusBadGateway, gin.H{"error": "AI upstream request failed"})
}
