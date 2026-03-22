package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"auto-memories-doll/ch/internal/config"
	"auto-memories-doll/ch/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}
	defer pool.Close()

	srv := server.New(cfg, pool)

	addr := ":" + cfg.Port
	log.Printf("C.H backend listening on %s", addr)
	
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
