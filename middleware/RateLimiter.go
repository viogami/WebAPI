package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

var rateLimiter = NewRateLimiter(1 * time.Second)

// RateLimiter 结构体
type RateLimiter struct {
    mu          sync.Mutex
    visitors    map[string]map[string]time.Time // 每个 IP 的路由访问记录
    blacklist   map[string]bool                 // 黑名单
    interval    time.Duration
    requests    map[string]int                  // 每个 IP 的请求计数
    maxRequests int                             // 最大请求次数
}

// NewRateLimiter 创建一个新的 RateLimiter 实例
func NewRateLimiter(interval time.Duration) *RateLimiter {
    rl := &RateLimiter{
        visitors:    make(map[string]map[string]time.Time),
        blacklist:   make(map[string]bool),
        interval:    interval,
        requests:    make(map[string]int),
        maxRequests: 5, // 设置最大请求次数（可根据需求调整）
    }

    // 启动一个后台协程定期清理过期记录
    go rl.cleanupVisitors()

    return rl
}

// Allow 判断是否允许访问
func (rl *RateLimiter) Allow(ip string, route string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    // 检查是否在黑名单中
    if rl.blacklist[ip] {
        return false
    }

    // 初始化 IP 的路由访问记录
    if _, exists := rl.visitors[ip]; !exists {
        rl.visitors[ip] = make(map[string]time.Time)
    }

    // 检查当前路由的访问时间
    if lastVisit, exists := rl.visitors[ip][route]; exists {
        if time.Since(lastVisit) < rl.interval {
            // 增加请求计数
            rl.requests[ip]++
            if rl.requests[ip] > rl.maxRequests {
                // 超过最大请求次数，加入黑名单
                rl.blacklist[ip] = true
                return false
            }
            return false
        }
    }

    // 更新访问时间
    rl.visitors[ip][route] = time.Now()
    rl.requests[ip]++
    return true
}

// cleanupVisitors 定期清理过期的 IP 记录
func (rl *RateLimiter) cleanupVisitors() {
    for {
        time.Sleep(1 * time.Minute) // 每分钟清理一次
        rl.mu.Lock()
        now := time.Now()
        for ip, routes := range rl.visitors {
            for route, lastSeen := range routes {
                if now.Sub(lastSeen) > rl.interval {
                    delete(routes, route) // 删除过期的路由记录
                }
            }
            // 如果该 IP 的所有路由记录都被清理，则删除该 IP
            if len(routes) == 0 {
                delete(rl.visitors, ip)
                delete(rl.requests, ip)
            }
        }
        rl.mu.Unlock()
    }
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        route := c.FullPath() // 获取当前路由路径
        if !rateLimiter.Allow(ip, route) {
            if rateLimiter.blacklist[ip] {
                c.JSON(http.StatusForbidden, gin.H{
                    "error": "Your IP has been permanently banned.Please contact me if you think this is a mistake.",
                })
            } else {
                c.JSON(http.StatusTooManyRequests, gin.H{
                    "error": "Too many requests. Please try again later.",
                })
            }
            c.Abort()
            return
        }
        c.Next()
    }
}