package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Server 结构体封装 Gin 服务
type Server struct {
	router *gin.Engine
	cfg    *Config
}

// 创建一个 Server 实例
func NewServer(cfg *Config) *Server {
	// 初始化 Gin
	g := gin.Default()

	// 创建 Server 实例
	return &Server{
		router: g,
		cfg:    cfg,
	}
}

// 注册路由
func (s *Server) RegisterRoutes() {
	s.router.GET("/", helloHandler)
	// 注册 /p5cc 路由动态路由
	s.router.GET("/p5cc/:text", p5ccHandler)
	s.router.POST("/p5cc", UpdateP5ccHandler)
}

// Run 启动 HTTP 服务
func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	log.Printf("Server running on %s", addr)
	if err := s.router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
