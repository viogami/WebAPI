package sdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"auto-memories-doll/ch/internal/config"
	chserver "auto-memories-doll/ch/internal/server"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DatabaseURL     string
	SessionTTLHours int
	PasswordPepper  string
	AllowedOrigin   string
}

func NewHTTPHandler(cfg Config) (http.Handler, func(), error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, nil, fmt.Errorf("databaseURL is required")
	}

	if cfg.SessionTTLHours <= 0 {
		cfg.SessionTTLHours = 168
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse db config failed: %w", err)
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	poolCfg.ConnConfig.StatementCacheCapacity = 0

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect db failed: %w", err)
	}

	chCfg := config.Config{
		DatabaseURL:     cfg.DatabaseURL,
		SessionTTLHours: cfg.SessionTTLHours,
		PasswordPepper:  cfg.PasswordPepper,
		AllowedOrigin:   cfg.AllowedOrigin,
	}

	srv := chserver.New(chCfg, pool)
	return srv.Routes(), pool.Close, nil
}