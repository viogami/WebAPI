package handlers

import (
	"github.com/viogami/WebAPI/core/ai/deepseek"
	"github.com/viogami/WebAPI/core/ai/openai"
	"net/http"

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


