package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"webapi/conf"
	chcore "webapi/core/c.h"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CHHandler struct {
	cfg         conf.CHApiConfig
	pool        *pgxpool.Pool
	userService *chcore.UserService
	rankService *chcore.RankService
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string `json:"token"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	ExpiresAt string `json:"expires_at"`
}

type historyUploadItem struct {
	AnimeID int64  `json:"anime_id"`
	Name    string `json:"name"`
	Cover   string `json:"cover"`
	AddedAt string `json:"added_at"`
}

type historyQueryItem struct {
	ID        int64  `json:"id"`
	AnimeID   int64  `json:"anime_id"`
	Name      string `json:"name"`
	Cover     string `json:"cover"`
	AddedAt   string `json:"added_at"`
	CreatedAt string `json:"created_at"`
}

type rankUploadRequest struct {
	Title         string          `json:"title"`
	TierBoardName string          `json:"tier_board_name"`
	GridBoardName string          `json:"grid_board_name"`
	Payload       json.RawMessage `json:"payload"`
}

func NewCHHandler(cfg conf.CHApiConfig) (*CHHandler, error) {
	if cfg.Enabled {
		if cfg.DatabaseURL == "" {
			log.Println("CHApi enabled but databaseURL is empty, skip route registration")
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
			if err != nil {
				return nil, fmt.Errorf("parse db config failed: %w", err)
			}

			poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
			poolCfg.ConnConfig.StatementCacheCapacity = 0

			pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
			if err != nil {
				return nil, fmt.Errorf("connect db failed: %w", err)
			}

			return &CHHandler{
				cfg:         cfg,
				pool:        pool,
				userService: chcore.NewUserService(pool, cfg.PasswordPepper, cfg.SessionTTLHours),
				rankService: chcore.NewRankService(pool),
			}, nil
		}
	}

	return nil, nil
}

func (h *CHHandler) RegisterRoutes(r *gin.Engine) {
	if h == nil {
		return
	}

	r.Use(h.corsMiddleware())

	r.GET("/CH/healthz", h.Healthz)

	v1 := r.Group("/CH/api/v1")
	{
		v1.OPTIONS("/*any", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		auth := v1.Group("/auth")
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)

		v1.GET("/me", h.AuthRequired(), h.Me)
		v1.POST("/history", h.AuthRequired(), h.UploadHistory)
		v1.GET("/history", h.AuthRequired(), h.ListHistory)
		v1.POST("/rank", h.AuthRequired(), h.CreateRank)
		v1.GET("/rank", h.AuthRequired(), h.ListRank)
		v1.GET("/rank/latest", h.AuthRequired(), h.LatestRank)
		v1.POST("/sync", h.AuthRequired(), h.Sync)
	}
}

func (h *CHHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *CHHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.userService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrInvalidUsernameLength), errors.Is(err, chcore.ErrPasswordTooShort):
			writeError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, chcore.ErrUsernameExists):
			writeError(c, http.StatusConflict, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "create user failed")
		}
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		Token:     result.Token,
		UserID:    result.UserID,
		Username:  result.Username,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *CHHandler) Login(c *gin.Context) {
	var req registerRequest
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrUsernameAndPasswordRequired):
			writeError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, chcore.ErrInvalidUsernameOrPassword):
			writeError(c, http.StatusUnauthorized, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "query user failed")
		}
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token:     result.Token,
		UserID:    result.UserID,
		Username:  result.Username,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *CHHandler) Me(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.userService.Me(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query user failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   result.UserID,
		"username":  result.Username,
		"logged_in": true,
	})
}

func (h *CHHandler) UploadHistory(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Items []historyUploadItem `json:"items"`
	}
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	items := make([]chcore.HistoryUploadItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, chcore.HistoryUploadItem{
			AnimeID: item.AnimeID,
			Name:    item.Name,
			Cover:   item.Cover,
			AddedAt: item.AddedAt,
		})
	}

	uploaded, err := h.userService.UploadHistory(c.Request.Context(), userID, items)
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrHistoryItemsEmpty), errors.Is(err, chcore.ErrHistoryItemFieldsRequired):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "insert history failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"uploaded": uploaded})
}

func (h *CHHandler) ListHistory(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseLimit(c.Query("limit"), 50, 1, 200)
	items, err := h.userService.ListHistory(c.Request.Context(), userID, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query history failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *CHHandler) CreateRank(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req rankUploadRequest
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.rankService.CreateRank(c.Request.Context(), userID, chcore.RankInput{
		Title:         req.Title,
		TierBoardName: req.TierBoardName,
		GridBoardName: req.GridBoardName,
		Payload:       req.Payload,
	})
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrRankFieldsRequired), errors.Is(err, chcore.ErrPayloadRequired):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "insert rank failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *CHHandler) ListRank(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseLimit(c.Query("limit"), 20, 1, 100)
	items, err := h.rankService.ListRank(c.Request.Context(), userID, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query rank failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *CHHandler) LatestRank(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.rankService.LatestRank(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query latest rank failed")
		return
	}
	if item == nil {
		c.JSON(http.StatusOK, gin.H{"item": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": item,
	})
}

func (h *CHHandler) Sync(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		History []historyUploadItem `json:"history"`
		Rank    *rankUploadRequest  `json:"rank"`
	}
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	historyItems := make([]chcore.HistoryUploadItem, 0, len(req.History))
	for _, item := range req.History {
		historyItems = append(historyItems, chcore.HistoryUploadItem{
			AnimeID: item.AnimeID,
			Name:    item.Name,
			Cover:   item.Cover,
			AddedAt: item.AddedAt,
		})
	}

	var rankInput *chcore.RankInput
	if req.Rank != nil {
		rankInput = &chcore.RankInput{
			Title:         req.Rank.Title,
			TierBoardName: req.Rank.TierBoardName,
			GridBoardName: req.Rank.GridBoardName,
			Payload:       req.Rank.Payload,
		}
	}

	result, err := h.rankService.Sync(c.Request.Context(), userID, chcore.SyncInput{
		History: historyItems,
		Rank:    rankInput,
	})
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrHistoryOrRank), errors.Is(err, chcore.ErrSyncHistoryItemFieldsRequired), errors.Is(err, chcore.ErrInvalidRankPayload):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "commit tx failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"history_uploaded": result.HistoryUploaded,
		"rank_id":          result.RankID,
	})
}

func (h *CHHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := parseBearerToken(c.GetHeader("Authorization"))
		userID, err := h.userService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, chcore.ErrMissingToken), errors.Is(err, chcore.ErrInvalidToken), errors.Is(err, chcore.ErrTokenExpired):
				writeError(c, http.StatusUnauthorized, err.Error())
			default:
				writeError(c, http.StatusUnauthorized, "invalid token")
			}
			c.Abort()
			return
		}

		c.Set("ch_user_id", userID)
		c.Next()
	}
}

func (h *CHHandler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(h.cfg.AllowedOrigin)
		if origin != "" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			origin = "http://" + origin
		}
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		c.Next()
	}
}
