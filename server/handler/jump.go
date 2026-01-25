package handlers

import (
	"io"
	"net/http"
	"strings"
	"time"
	"webapi/core/jump"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-gonic/gin"
)

type JumpHandler struct {
	cache *ristretto.Cache
}

type cachedResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

func NewJumpHandler() *JumpHandler {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100 << 20, // 最大 100 MB
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	return &JumpHandler{
		cache: cache,
	}
}

func (h *JumpHandler) GithubProxy(c *gin.Context) {
	proxyPath := strings.TrimPrefix(c.Param("proxyPath"), "/")

	targetURL, err := jump.BuildGithubURL(proxyPath, c.Request.URL.RawQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := "github:" + targetURL

	// ---------- 1. 尝试走缓存 ----------
	if v, ok := h.cache.Get(key); ok {
		cached := v.(*cachedResponse)

		c.Status(cached.StatusCode)
		for k, v := range cached.Headers {
			c.Header(k, v)
		}
		c.Writer.Write(cached.Body)
		return
	}

	// ---------- 2. 回源 ----------
	resp, err := http.Get(targetURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// ---------- 3. 透传 ----------
	c.Status(resp.StatusCode)

	headers := make(map[string]string)
	for _, hkey := range []string{
		"Content-Type",
		"Cache-Control",
	} {
		if v := resp.Header.Get(hkey); v != "" {
			headers[hkey] = v
			c.Header(hkey, v)
		}
	}

	c.Writer.Write(body)

	// ---------- 4. 写缓存（只缓存成功响应） ----------
	if resp.StatusCode == http.StatusOK {
		h.cache.SetWithTTL(
			key,
			&cachedResponse{
				StatusCode: resp.StatusCode,
				Headers:    headers,
				Body:       body,
			},
			int64(len(body)), // cost = body size
			24*time.Hour,     // TTL = 1 天
		)
	}
}
