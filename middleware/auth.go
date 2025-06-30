package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 鉴权中间件，示例：校验Authorization Bearer Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}
		// 期望格式：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证信息格式错误"})
			c.Abort()
			return
		}
		token := parts[1]
		// 这里写死一个示例token，实际应替换为你的校验逻辑
		if !checkToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func checkToken(token string) bool {
	return token == "vio-deepseek-token"
}


