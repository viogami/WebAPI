package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"webapi/conf"
	"webapi/server"
)

func main() {
	cfg, err := conf.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	s, err := server.NewServer(cfg)
	if err != nil {
		slog.Error("create server", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s.RegisterRoutes()

	if err := s.Run(ctx); err != nil {
		slog.Error("run server", "error", err)
		os.Exit(1)
	}
}
