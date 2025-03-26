package handlers

import (
	"net/http"
	"strconv"
	"WebAPI/conf"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	
}

func(h *Handler) HelloHandler(c *gin.Context) {
	c.String(http.StatusOK, conf.AppConfig.TextConfig.HelloText)
}

func NewHandler() *Handler {
	return &Handler{}
}

// 创建辅助函数处理不同类型参数获取
func getFloatParam(c *gin.Context, param string, defaultValue float64) float64 {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseFloat(paramStr, 64); err == nil {
			return val
		}
	}
	return defaultValue
}

func getBoolParam(c *gin.Context, param string, defaultValue bool) bool {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseBool(paramStr); err == nil {
			return val
		}
	}
	return defaultValue
}

func getIntParam(c *gin.Context, param string, defaultValue int) int {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.Atoi(paramStr); err == nil {
			return val
		}
	}
	return defaultValue
}