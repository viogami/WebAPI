package server

import (
	"WebAPI/conf"
	"WebAPI/middleware"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    *conf.Config
	router *gin.Engine
}

// 添加中间件
func (s *Server) UseMiddleware() {
	s.router.Use(middleware.RateLimitMiddleware())
}

// Run 启动 HTTP 服务
func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Printf("Server running on %s", addr)
	if err := s.router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// 创建一个 Server 实例
func NewServer(cfg *conf.Config) *Server {
	// 设置 Gin 的运行模式
	gin.SetMode(cfg.GinMode)

	g := gin.Default()
	
	// 创建 Server 实例
	return &Server{
		router: g,
		cfg:    cfg,
	}
}