package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"auto-memories-doll/ch/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg  config.Config
	pool *pgxpool.Pool
}

func New(cfg config.Config, pool *pgxpool.Pool) *Server {
	return &Server{cfg: cfg, pool: pool}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/me", s.withAuth(s.handleMe))
	mux.HandleFunc("/api/v1/history", s.withAuth(s.handleHistory))
	mux.HandleFunc("/api/v1/rank", s.withAuth(s.handleRank))
	mux.HandleFunc("/api/v1/rank/latest", s.withAuth(s.handleLatestRank))
	mux.HandleFunc("/api/v1/sync", s.withAuth(s.handleSync))

	return s.withCORS(mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	username := strings.TrimSpace(req.Username)
	if len(username) < 3 || len(username) > 32 {
		writeError(w, http.StatusBadRequest, "username length must be 3-32")
		return
	}
	if len(req.Password) < 4 {
		writeError(w, http.StatusBadRequest, "password is too short")
		return
	}

	salt, err := randomHex(16)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate salt")
		return
	}
	hash := hashPassword(req.Password, salt, s.cfg.PasswordPepper)

	var userID int64
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO users (username, password_salt, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, username, salt, hash).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(w, http.StatusConflict, "username already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "create user failed")
		return
	}

	token, expiresAt, err := s.createSession(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	var userID int64
	var salt, storedHash string
	err := s.pool.QueryRow(r.Context(), `
		SELECT id, password_salt, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&userID, &salt, &storedHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "query user failed")
		return
	}

	if hashPassword(req.Password, salt, s.cfg.PasswordPepper) != storedHash {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, expiresAt, err := s.createSession(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var username string
	err := s.pool.QueryRow(r.Context(), `SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query user failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":   userID,
		"username":  username,
		"logged_in": true,
	})
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

type authedHandler func(http.ResponseWriter, *http.Request, int64)

func (s *Server) withAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		token := parseBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}

		var userID int64
		var expiresAt time.Time
		err := s.pool.QueryRow(r.Context(), `
			SELECT user_id, expires_at
			FROM user_sessions
			WHERE token = $1
		`, token).Scan(&userID, &expiresAt)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if time.Now().After(expiresAt) {
			_, _ = s.pool.Exec(context.Background(), `DELETE FROM user_sessions WHERE token = $1`, token)
			writeError(w, http.StatusUnauthorized, "token expired")
			return
		}

		next(w, r, userID)
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) createSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := randomHex(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(time.Duration(s.cfg.SessionTTLHours) * time.Hour)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO user_sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}
