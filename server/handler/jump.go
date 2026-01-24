package handlers

import (
	"io"
	"net/http"
	"strings"
	"webapi/core/jump"

	"github.com/gin-gonic/gin"
)

type JumpHandler struct {
}

func NewJumpHandler() *JumpHandler {
	return &JumpHandler{}
}

func (h *JumpHandler) GithubProxy(c *gin.Context) {
	proxyPath := strings.TrimPrefix(c.Param("proxyPath"), "/")

	targetURL, err := jump.BuildGithubURL(proxyPath, c.Request.URL.RawQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 透传状态码
	c.Status(resp.StatusCode)

	// 透传 Content-Type（SVG / PNG / JSON 都靠它）
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		c.Header("Content-Type", ct)
	}

	// 可选：Cache-Control
	if cc := resp.Header.Get("Cache-Control"); cc != "" {
		c.Header("Cache-Control", cc)
	}

	// 原样返回 body
	_, _ = io.Copy(c.Writer, resp.Body)
}
