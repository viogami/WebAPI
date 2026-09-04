package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"webapi/conf"
	"webapi/middleware"
	h "webapi/server/handler"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg       *conf.Config
	router    *gin.Engine
	chHandler *h.CHHandler
}

func NewServer(cfg *conf.Config) (*Server, error) {
	gin.SetMode(cfg.Server.GinMode)

	router := gin.Default()
	if err := router.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("configure trusted proxies: %w", err)
	}
	router.Use(middleware.NewRateLimiter(middleware.RateLimitConfig{
		RequestsPerSecond: cfg.Server.RateLimit.RequestsPerSecond,
		Burst:             cfg.Server.RateLimit.Burst,
		IdleTTL:           time.Duration(cfg.Server.RateLimit.IdleTTLSeconds) * time.Second,
	}).Middleware())

	chHandler, err := h.NewCHHandler(cfg.CH)
	if err != nil {
		return nil, fmt.Errorf("initialize CH handler: %w", err)
	}

	return &Server{
		cfg:       cfg,
		router:    router,
		chHandler: chHandler,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	slog.Info("server running", "address", addr)
	defer s.Close()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	}
}

func (s *Server) Close() {
	if s.chHandler != nil {
		s.chHandler.Close()
	}
}
