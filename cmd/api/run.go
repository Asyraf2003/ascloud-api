package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/platform/datastore/postgres"
	"example.com/your-api/internal/platform/google"
	jwtp "example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/transport/http/router"
	"example.com/your-api/internal/transport/http/server"
)

func run() {
	addr := env("HTTP_ADDR", ":"+env("HTTP_PORT", "8080"))
	service := env("SERVICE_NAME", "api")
	shutdownTimeout := envDuration("SHUTDOWN_TIMEOUT", 10*time.Second)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log = log.With("service", service)

	dsn := env("DB_DSN", env("DATABASE_URL", ""))
	if dsn == "" {
		log.Error("missing db dsn (set DB_DSN or DATABASE_URL)")
		os.Exit(1)
	}

	db, err := postgres.New(dsn)
	if err != nil {
		log.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	var _ *sql.DB = db

	authCfg := config.LoadAuth()
	if err := authCfg.Validate(); err != nil {
		log.Error("invalid auth config", "err", err)
		os.Exit(1)
	}

	// JWT verifier sesuai verifier.go kamu
	jwtv, err := jwtp.NewHMACVerifier(authCfg.JWT.Issuer, authCfg.JWT.Audience, authCfg.JWT.Secret)
	if err != nil {
		log.Error("jwt verifier init failed", "err", err)
		os.Exit(1)
	}

	if err := google.WireAuthGoogle(db, authCfg); err != nil {
		log.Error("wire auth google failed", "err", err)
		os.Exit(1)
	}

	e := server.New(log, db)
	router.Register(e, jwtv) // <- signature router diubah

	go func() {
		log.Info("starting", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("server start failed", "err", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	log.Info("shutting down")

	sdCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := e.Shutdown(sdCtx); err != nil {
		log.Error("shutdown failed", "err", err)
		return
	}
	log.Info("shutdown complete")
}
