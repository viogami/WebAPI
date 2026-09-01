package server

import h "webapi/server/handler"

func (s *Server) RegisterRoutes() {
	r := s.router

	baseHandler := h.NewHandler(s.cfg.Text.HelloText)
	p5ccHandler := h.NewP5ccHandler(s.cfg.P5cc)
	aiHandler := h.NewAIHandler(s.cfg.AI)
	jumpHandler := h.NewJumpHandler()

	r.GET("/", baseHandler.HelloHandler)
	r.GET("/healthz", baseHandler.HealthHandler)

	r.GET("/p5cc/:text", p5ccHandler.GET)
	r.POST("/p5cc", p5ccHandler.POST)

	r.POST("/gpt", aiHandler.ProcessMessage)
	r.POST("/deepseek", aiHandler.ProcessSharpReviews)

	r.GET("/jump/github/*proxyPath", jumpHandler.GithubProxy)

	if s.chHandler != nil {
		s.chHandler.RegisterRoutes(r)
	}
}
