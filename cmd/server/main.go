package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaejin1/dottie-api/internal/config"
	"github.com/jaejin1/dottie-api/internal/pkg/firebase"
	"github.com/jaejin1/dottie-api/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	logger, err := buildLogger(cfg.Env)
	if err != nil {
		log.Fatalf("logger init error: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	var fbClient *firebase.Client
	if cfg.FirebaseCredentials != "" {
		fbClient, err = firebase.NewClient([]byte(cfg.FirebaseCredentials))
		if err != nil {
			logger.Fatal("firebase init error", zap.Error(err))
		}
		logger.Info("firebase client initialized")
	} else {
		logger.Warn("FIREBASE_CREDENTIALS not set — auth middleware disabled")
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		logger.Fatal("db pool init error", zap.Error(err))
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Fatal("db ping failed", zap.Error(err))
	}
	logger.Info("database connected")

	s := server.New(cfg, logger, fbClient, dbPool)
	if err := s.Start(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}

func buildLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
