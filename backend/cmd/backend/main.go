package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "backend",
		Short: "Backend service CLI",
		Long:  "Backend service CLI",
	}

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server",
		RunE:  runServer,
	}
)

func init() {
	// Flags
	serverCmd.Flags().Int("port", 6161, "HTTP port")
	_ = viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))

	// Config precedence: flags > env > file
	viper.SetEnvPrefix("BACKEND")
	viper.AutomaticEnv()

	rootCmd.AddCommand(serverCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Logger (JSON)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := viper.GetInt("port")
	addr := fmt.Sprintf(":%d", port)

	// Router
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           requestLoggerMiddleware(logger, r),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start
	go func() {
		logger.Info("server starting", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("server shutting down")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}
	logger.Info("server stopped")
	return nil
}

func requestLoggerMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rlw := &responseLogger{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rlw, r)
		dur := time.Since(start)
		logger.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rlw.statusCode),
			slog.Duration("duration_ms", dur),
		)
	})
}

type responseLogger struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *responseLogger) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
