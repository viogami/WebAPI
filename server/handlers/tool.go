package handlers

import (
	"github.com/gin-gonic/gin"
)

type ToolHandler struct {
}

func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

func (h *ToolHandler) BestWorkTime(c *gin.Context) {
	reply := "开发中"

	c.JSON(200, gin.H{"best_work_time": reply})
}
