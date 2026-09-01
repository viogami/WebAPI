package handlers

import (
	"net/http"

	chsdk "auto-memories-doll/ch/sdk"
	"webapi/conf"

	"github.com/gin-gonic/gin"
)

type CHHandler struct {
	handler http.Handler
	closeFn func()
}

func NewCHHandler(cfg conf.CHAPIConfig) (*CHHandler, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	handler, closeFn, err := chsdk.NewHTTPHandler(chsdk.Config{
		DatabaseURL:     cfg.DatabaseURL,
		SessionTTLHours: cfg.SessionTTLHours,
		PasswordPepper:  cfg.PasswordPepper,
		AllowedOrigin:   cfg.AllowedOrigin,
	})
	if err != nil {
		return nil, err
	}

	return &CHHandler{handler: handler, closeFn: closeFn}, nil
}

func (h *CHHandler) RegisterRoutes(r *gin.Engine) {
	wrapped := gin.WrapH(http.StripPrefix("/CH", h.handler))
	r.Any("/CH", wrapped)
	r.Any("/CH/*any", wrapped)
}

func (h *CHHandler) Close() {
	if h != nil && h.closeFn != nil {
		h.closeFn()
	}
}
