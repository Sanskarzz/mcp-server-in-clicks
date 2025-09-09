package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"mcp-backend/internal/api"
	"mcp-backend/internal/config"
	"mcp-backend/internal/helm"
	"mcp-backend/internal/storage"
)

func main() {
	_ = godotenv.Load()
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)

	cfg := config.Load()

	// Mongo connection
	mongo, err := storage.NewMongoStore(context.Background(), cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.WithError(err).Warn("mongo not available, continuing (dev mode)")
	}
	if mongo != nil {
		defer mongo.Close(context.Background())
	}

	// Helm service
	helmSvc := helm.NewService(cfg)

	// API server
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// JWT middleware (HMAC shared secret)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}
	r.Use(api.AuthMiddleware(secret))

	api.AttachRoutes(r, log, mongo, helmSvc)

	srv := &http.Server{
		Addr:         ":6000",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.WithField("addr", srv.Addr).Info("backend listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
