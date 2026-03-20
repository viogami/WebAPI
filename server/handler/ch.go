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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CHHandler struct {
	cfg  conf.CHApiConfig
	pool *pgxpool.Pool
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

			pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
			if err != nil {
				return nil, fmt.Errorf("connect db failed: %w", err)
			}

			return &CHHandler{cfg: cfg, pool: pool}, nil
		}
	}

	return nil, nil
}

func (h *CHHandler) RegisterRoutes(r *gin.Engine) {
	if h == nil {
		return
	}

	r.GET("/CH/healthz", h.Healthz)

	v1 := r.Group("/CH/api/v1")
	v1.Use(h.corsMiddleware())
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

	username := strings.TrimSpace(req.Username)
	if len(username) < 3 || len(username) > 32 {
		writeError(c, http.StatusBadRequest, "username length must be 3-32")
		return
	}
	if len(req.Password) < 4 {
		writeError(c, http.StatusBadRequest, "password is too short")
		return
	}

	salt, err := randomHex(16)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to generate salt")
		return
	}
	hash := hashPassword(req.Password, salt, h.cfg.PasswordPepper)

	var userID int64
	err = h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO users (username, password_salt, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, username, salt, hash).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(c, http.StatusConflict, "username already exists")
			return
		}
		writeError(c, http.StatusInternalServerError, "create user failed")
		return
	}

	token, expiresAt, err := h.createSession(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "create session failed")
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (h *CHHandler) Login(c *gin.Context) {
	var req registerRequest
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		writeError(c, http.StatusBadRequest, "username and password are required")
		return
	}

	var userID int64
	var salt, storedHash string
	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT id, password_salt, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&userID, &salt, &storedHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(c, http.StatusUnauthorized, "invalid username or password")
			return
		}
		writeError(c, http.StatusInternalServerError, "query user failed")
		return
	}

	if hashPassword(req.Password, salt, h.cfg.PasswordPepper) != storedHash {
		writeError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, expiresAt, err := h.createSession(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "create session failed")
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (h *CHHandler) Me(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var username string
	err := h.pool.QueryRow(c.Request.Context(), `SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query user failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"username":  username,
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
	if len(req.Items) == 0 {
		writeError(c, http.StatusBadRequest, "items cannot be empty")
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "begin tx failed")
		return
	}
	defer tx.Rollback(c.Request.Context())

	for _, item := range req.Items {
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Cover) == "" {
			writeError(c, http.StatusBadRequest, "anime_id, name, cover are required")
			return
		}
		clientAddedAt := parseRFC3339Nullable(item.AddedAt)
		_, err := tx.Exec(c.Request.Context(), `
			INSERT INTO anime_history_records (user_id, anime_id, anime_name, cover, client_added_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, item.AnimeID, item.Name, item.Cover, clientAddedAt)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "insert history failed")
			return
		}
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeError(c, http.StatusInternalServerError, "commit failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"uploaded": len(req.Items)})
}

func (h *CHHandler) ListHistory(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseLimit(c.Query("limit"), 50, 1, 200)
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, anime_id, anime_name, cover,
		       COALESCE(client_added_at::text, '') AS client_added_at,
		       created_at::text
		FROM anime_history_records
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query history failed")
		return
	}
	defer rows.Close()

	items := make([]historyQueryItem, 0, limit)
	for rows.Next() {
		var item historyQueryItem
		if err := rows.Scan(&item.ID, &item.AnimeID, &item.Name, &item.Cover, &item.AddedAt, &item.CreatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "scan history failed")
			return
		}
		items = append(items, item)
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

	req.Title = strings.TrimSpace(req.Title)
	req.TierBoardName = strings.TrimSpace(req.TierBoardName)
	req.GridBoardName = strings.TrimSpace(req.GridBoardName)
	if req.Title == "" || req.TierBoardName == "" || req.GridBoardName == "" {
		writeError(c, http.StatusBadRequest, "title, tier_board_name, grid_board_name are required")
		return
	}
	if len(req.Payload) == 0 {
		writeError(c, http.StatusBadRequest, "payload is required")
		return
	}

	var id int64
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.Title, req.TierBoardName, req.GridBoardName, req.Payload).Scan(&id)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "insert rank failed")
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
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query rank failed")
		return
	}
	defer rows.Close()

	type rankItem struct {
		ID            int64           `json:"id"`
		Title         string          `json:"title"`
		TierBoardName string          `json:"tier_board_name"`
		GridBoardName string          `json:"grid_board_name"`
		Payload       json.RawMessage `json:"payload"`
		CreatedAt     string          `json:"created_at"`
	}

	items := make([]rankItem, 0, limit)
	for rows.Next() {
		var item rankItem
		if err := rows.Scan(&item.ID, &item.Title, &item.TierBoardName, &item.GridBoardName, &item.Payload, &item.CreatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "scan rank failed")
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *CHHandler) LatestRank(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var id int64
	var title, tierBoardName, gridBoardName string
	var payload json.RawMessage
	var createdAt string
	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&id, &title, &tierBoardName, &gridBoardName, &payload, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{"item": nil})
			return
		}
		writeError(c, http.StatusInternalServerError, "query latest rank failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": gin.H{
			"id":              id,
			"title":           title,
			"tier_board_name": tierBoardName,
			"grid_board_name": gridBoardName,
			"payload":         payload,
			"created_at":      createdAt,
		},
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
	if len(req.History) == 0 && req.Rank == nil {
		writeError(c, http.StatusBadRequest, "history or rank is required")
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "begin tx failed")
		return
	}
	defer tx.Rollback(c.Request.Context())

	historyCount := 0
	for _, item := range req.History {
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Cover) == "" {
			writeError(c, http.StatusBadRequest, "history item fields are required")
			return
		}
		clientAddedAt := parseRFC3339Nullable(item.AddedAt)
		_, err := tx.Exec(c.Request.Context(), `
			INSERT INTO anime_history_records (user_id, anime_id, anime_name, cover, client_added_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, item.AnimeID, item.Name, item.Cover, clientAddedAt)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "insert history failed")
			return
		}
		historyCount++
	}

	var rankID *int64
	if req.Rank != nil {
		req.Rank.Title = strings.TrimSpace(req.Rank.Title)
		req.Rank.TierBoardName = strings.TrimSpace(req.Rank.TierBoardName)
		req.Rank.GridBoardName = strings.TrimSpace(req.Rank.GridBoardName)
		if req.Rank.Title == "" || req.Rank.TierBoardName == "" || req.Rank.GridBoardName == "" || len(req.Rank.Payload) == 0 {
			writeError(c, http.StatusBadRequest, "invalid rank payload")
			return
		}

		var id int64
		err := tx.QueryRow(c.Request.Context(), `
			INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, userID, req.Rank.Title, req.Rank.TierBoardName, req.Rank.GridBoardName, req.Rank.Payload).Scan(&id)
		if err != nil {
			writeError(c, http.StatusInternalServerError, "insert rank failed")
			return
		}
		rankID = &id
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeError(c, http.StatusInternalServerError, "commit tx failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"history_uploaded": historyCount,
		"rank_id":          rankID,
	})
}

func (h *CHHandler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := parseBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			writeError(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}

		var userID int64
		var expiresAt time.Time
		err := h.pool.QueryRow(c.Request.Context(), `
			SELECT user_id, expires_at
			FROM user_sessions
			WHERE token = $1
		`, token).Scan(&userID, &expiresAt)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		if time.Now().After(expiresAt) {
			_, _ = h.pool.Exec(context.Background(), `DELETE FROM user_sessions WHERE token = $1`, token)
			writeError(c, http.StatusUnauthorized, "token expired")
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

func (h *CHHandler) createSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := randomHex(32)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.SessionTTLHours) * time.Hour)
	_, err = h.pool.Exec(ctx, `
		INSERT INTO user_sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}
