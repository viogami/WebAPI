package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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
	// 仅允许 GET / HEAD
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}

	proxyPath := strings.TrimPrefix(c.Param("proxyPath"), "/")

	targetHost, targetPath, err := jump.ParseGithubProxyPath(proxyPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetURL, err := url.Parse("https://" + targetHost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target host"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义 Director，防止 Gin 改写
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = targetHost
		req.URL.Path = targetPath
		req.URL.RawQuery = c.Request.URL.RawQuery
		req.Host = targetHost

		// 可选：清理部分 Header
		req.Header.Del("Referer")
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
