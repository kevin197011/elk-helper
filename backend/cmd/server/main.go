// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/api/routes"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	es_config "github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	system_config "github.com/kk/elk-helper/backend/internal/service/systemconfig"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"github.com/kk/elk-helper/backend/internal/worker/scheduler"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	if err := config.AppConfig.Validate(); err != nil {
		slog.Error("Invalid config", "error", err)
		os.Exit(1)
	}

	// Initialize database
	if err := database.Init(config.AppConfig.Database); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services
	queryService, err := query.NewService()
	if err != nil {
		slog.Error("Failed to create query service", "error", err)
		os.Exit(1)
	}

	ruleService := rule.NewService()
	alertService := alert.NewService()
	esConfigService := es_config.NewService()
	systemConfigService := system_config.NewService()

	// Start worker scheduler if enabled
	var sched *scheduler.Scheduler
	if config.AppConfig.Worker.Enabled {
		slog.Info("Starting worker scheduler...")
		sched = scheduler.NewScheduler(
			ruleService,
			queryService,
			esConfigService,
			alertService,
			systemConfigService,
			time.Duration(config.AppConfig.Worker.CheckInterval)*time.Second,
			config.AppConfig.Worker.RetryTimes,
			config.AppConfig.Worker.BatchSize,
		)

		if err := sched.Start(); err != nil {
			slog.Error("Failed to start scheduler", "error", err)
			os.Exit(1)
		}
		slog.Info("Worker scheduler started", "check_interval", config.AppConfig.Worker.CheckInterval)
	}

	// Set Gin mode
	gin.SetMode(config.AppConfig.Server.Mode)

	// Create Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	// Start server in goroutine
	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)

	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("API server starting", "address", addr)
		if err := r.Run(addr); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("Received signal", "signal", sig)
	case err := <-serverErrChan:
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	// Shutdown gracefully
	slog.Info("Shutting down...")
	if sched != nil {
		sched.Stop()
		slog.Info("Worker scheduler stopped")
	}
	slog.Info("Shutdown complete")
}
