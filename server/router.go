package server

import (
	h "WebAPI/server/handlers"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() *gin.Engine {
	r := s.router

	// 创建handler
	h1 := h.NewHandler()
	h2 := h.NewP5ccHandler()
	h3 := h.NewWxapiHandler()

	r.GET("/", h1.HelloHandler)

	// p5cc
	r.GET("/p5cc/:text", h2.GET)
	r.POST("/p5cc", h2.POST)

	// wxapi
	r.GET("/wxapi", h3.Hello)
	r.GET("/wxapi/v1", h3.Redirect)
	r.GET("/wxapi/v1/oa", h3.WXCheckSignature)
	r.POST("/wxapi/v1/oa", h3.WXMsgReceive)
	r.GET("/wxapi/v1/oa/menu", h3.CheckMenu)
	//获取token
	r.GET("/wxapi/v1/oa/basic/get_access_token", h3.GetAccessToken)
	//获取微信callback IP
	r.GET("/wxapi/v1/oa/basic/get_callback_ip", h3.GetCallbackIP)
	//获取微信API接口 IP
	r.GET("/wxapi/v1/oa/basic/get_api_domain_ip", h3.GetAPIDomainIP)
	//清理接口调用次数
	r.GET("/wxapi/v1/oa/basic/clear_quota", h3.ClearQuota)

	return r
}
