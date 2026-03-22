package handlers

import (
	"log"
	"net/http"
	"strings"
	"webapi/conf"

	chsdk "auto-memories-doll/ch/sdk"

	"github.com/gin-gonic/gin"
)

type CHHandler struct {
	handler http.Handler
	closeFn func()
}

func NewCHHandler(cfg conf.CHApiConfig) (*CHHandler, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		log.Println("CHApi enabled but databaseURL is empty, skip route registration")
		return nil, nil
	}

	h, closeFn, err := chsdk.NewHTTPHandler(chsdk.Config{
		DatabaseURL:     cfg.DatabaseURL,
		SessionTTLHours: cfg.SessionTTLHours,
		PasswordPepper:  cfg.PasswordPepper,
		AllowedOrigin:   cfg.AllowedOrigin,
	})
	if err != nil {
		return nil, err
	}

	return &CHHandler{handler: h, closeFn: closeFn}, nil
}

func (h *CHHandler) RegisterRoutes(r *gin.Engine) {
	if h == nil || h.handler == nil {
		return
	}

	wrapped := gin.WrapH(http.StripPrefix("/CH", h.handler))
	r.Any("/CH", wrapped)
	r.Any("/CH/*any", wrapped)
}

func (h *CHHandler) Close() {
	if h == nil || h.closeFn == nil {
		return
	}
	h.closeFn()
}
