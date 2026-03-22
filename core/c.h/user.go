package chcore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidUsernameLength       = errors.New("username length must be 3-32")
	ErrPasswordTooShort            = errors.New("password is too short")
	ErrUsernameAndPasswordRequired = errors.New("username and password are required")
	ErrUsernameExists              = errors.New("username already exists")
	ErrInvalidUsernameOrPassword   = errors.New("invalid username or password")
	ErrMissingToken                = errors.New("missing token")
	ErrInvalidToken                = errors.New("invalid token")
	ErrTokenExpired                = errors.New("token expired")
)

type UserService struct {
	pool            *pgxpool.Pool
	passwordPepper  string
	sessionTTLHours int
}

type AuthResult struct {
	Token     string
	UserID    int64
	Username  string
	ExpiresAt time.Time
}

type MeResult struct {
	UserID   int64
	Username string
}

func NewUserService(pool *pgxpool.Pool, passwordPepper string, sessionTTLHours int) *UserService {
	return &UserService{
		pool:            pool,
		passwordPepper:  passwordPepper,
		sessionTTLHours: sessionTTLHours,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return nil, ErrInvalidUsernameLength
	}
	if len(password) < 4 {
		return nil, ErrPasswordTooShort
	}

	salt, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	hash := hashPassword(password, salt, s.passwordPepper)

	var userID int64
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_salt, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, username, salt, hash).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	token, expiresAt, err := s.createSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrUsernameAndPasswordRequired
	}

	var userID int64
	var salt, storedHash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, password_salt, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&userID, &salt, &storedHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidUsernameOrPassword
		}
		return nil, err
	}

	if hashPassword(password, salt, s.passwordPepper) != storedHash {
		return nil, ErrInvalidUsernameOrPassword
	}

	token, expiresAt, err := s.createSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:     token,
		UserID:    userID,
		Username:  username,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *UserService) Me(ctx context.Context, userID int64) (*MeResult, error) {
	var username string
	err := s.pool.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		return nil, err
	}

	return &MeResult{UserID: userID, Username: username}, nil
}

func (s *UserService) ValidateSession(ctx context.Context, token string) (int64, error) {
	if strings.TrimSpace(token) == "" {
		return 0, ErrMissingToken
	}

	var userID int64
	var expiresAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT user_id, expires_at
		FROM user_sessions
		WHERE token = $1
	`, token).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	if time.Now().After(expiresAt) {
		_, _ = s.pool.Exec(context.Background(), `DELETE FROM user_sessions WHERE token = $1`, token)
		return 0, ErrTokenExpired
	}

	return userID, nil
}

func (s *UserService) createSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := randomHex(32)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(time.Duration(s.sessionTTLHours) * time.Hour)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO user_sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func hashPassword(password, salt, pepper string) string {
	sum := sha256.Sum256([]byte(password + ":" + salt + ":" + pepper))
	return hex.EncodeToString(sum[:])
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func parseRFC3339Nullable(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return t
}
