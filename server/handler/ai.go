package handlers

import (
	"net/http"
	"webapi/core/ai/deepseek"
	"webapi/core/ai/openai"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
}

func NewAIHandler() *AIHandler {
	return &AIHandler{}
}

// OpenAI ChatGPT
func (h *AIHandler) ProcessMessage(c *gin.Context) {
	message := c.PostForm("message")
	reply := openai.GetInstance().InvokeChatGPTAPI(message)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func (h *AIHandler) ProcessMessageWithRole(c *gin.Context) {
	message := c.PostForm("message")
	role := c.PostForm("role")
	reply := openai.GetInstance().InvokeChatGPTAPIWithRole(message, role)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// DeepSeek
func (h *AIHandler) ProcessSharpReviews(c *gin.Context) {
	message := c.PostForm("message")
	reply := deepseek.GetInstance().InvokeDeepSeekAPI(message)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}
