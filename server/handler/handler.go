package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webapi/conf"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func (h *Handler) HelloHandler(c *gin.Context) {
	c.String(http.StatusOK, conf.AppConfig.TextConfig.HelloText)
}

func NewHandler() *Handler {
	return &Handler{}
}

// 创建辅助函数处理不同类型参数获取
func getFloatParam(c *gin.Context, param string, defaultValue ...float64) float64 {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseFloat(paramStr, 64); err == nil {
			return val
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func getBoolParam(c *gin.Context, param string, defaultValue ...bool) bool {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.ParseBool(paramStr); err == nil {
			return val
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

func getIntParam(c *gin.Context, param string, defaultValue ...int) int {
	if paramStr := c.PostForm(param); paramStr != "" {
		if val, err := strconv.Atoi(paramStr); err == nil {
			return val
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func authedUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get("ch_user_id")
	if !ok {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}

func parseBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func hashPassword(password, salt, pepper string) string {
	sum := sha256.Sum256([]byte(password + ":" + salt + ":" + pepper))
	return hex.EncodeToString(sum[:])
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func decodeJSONBody(c *gin.Context, target any) error {
	defer c.Request.Body.Close()
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func parseLimit(raw string, fallback, min, max int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func parseRFC3339Nullable(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return t
}
