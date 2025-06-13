package handlers

import (
	"WebAPI/core/AI/deepseek"
	myopenai "WebAPI/core/AI/openai"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
}

// GPT
func (h *AIHandler) ProcessMessage(c *gin.Context) {
	message := c.PostForm("message")
	reply := myopenai.GetInstance().InvokeChatGPTAPI(message)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}
func (h *AIHandler) ProcessMessageWithRole(c *gin.Context) {
	message := c.PostForm("message")
	role := c.PostForm("role")
	reply := myopenai.GetInstance().InvokeChatGPTAPIWithRole(message, role)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// DeepSeek
func (h *AIHandler) ProcessSharpReviews(c *gin.Context) {
	message := c.PostForm("message")
	reply := deepseek.GetInstance().InvokeDeepSeekAPI(message)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func NewAIHandler() *AIHandler {
	return &AIHandler{}
}
