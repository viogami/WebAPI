package handlers

import (
	"WebAPI/core/tool"

	"github.com/gin-gonic/gin"
)

type ToolHandler struct {
}

func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

func (h *ToolHandler) BestWorkTime(c *gin.Context) {
	avgMonthlyHours := getFloatParam(c, "avg_monthly_hours")
	remainingWorkDays := getIntParam(c, "remaining_work_days")

	reply := tool.CalculateBestWorkTime(avgMonthlyHours, remainingWorkDays)

	c.JSON(200, gin.H{"best_work_time": reply})
}
