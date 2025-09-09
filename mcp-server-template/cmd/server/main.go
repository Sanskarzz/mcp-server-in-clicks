package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-server-template/internal/config"
	"mcp-server-template/internal/server"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "config.json", "Path to configuration file")
		port       = flag.Int("port", 8080, "Server port")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		envFile    = flag.String("env", ".env", "Environment file path")
	)
	flag.Parse()

	// Setup logging
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Invalid log level")
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Load environment variables
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			logrus.WithError(err).Debug("Could not load env file (this is optional)")
		}
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		logrus.WithError(err).Fatal("Configuration validation failed")
	}

	logrus.WithFields(logrus.Fields{
		"server_name":     cfg.Server.Name,
		"server_version":  cfg.Server.Version,
		"tools_count":     len(cfg.Tools),
		"prompts_count":   len(cfg.Prompts),
		"resources_count": len(cfg.Resources),
	}).Info("Configuration loaded successfully")

	// Create and configure MCP server
	mcpServer, err := server.New(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create MCP server")
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logrus.WithField("signal", sig).Info("Received shutdown signal")

		// Give the server time to finish processing
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := mcpServer.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("Error during server shutdown")
		}

		cancel()
	}()

	// Start the server
	logrus.WithFields(logrus.Fields{
		"port":        *port,
		"config_path": *configPath,
	}).Info("Starting MCP server")

	if err := mcpServer.Start(ctx, *port); err != nil {
		logrus.WithError(err).Fatal("Server failed to start")
	}

	logrus.Info("MCP server stopped gracefully")
}
