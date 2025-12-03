// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/api/routes"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	"github.com/kk/elk-helper/backend/internal/service/systemconfig"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"github.com/kk/elk-helper/backend/internal/worker/scheduler"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := config.AppConfig.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Initialize database
	if err := database.Init(config.AppConfig.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize services
	queryService, err := query.NewService()
	if err != nil {
		log.Fatalf("Failed to create query service: %v", err)
	}

	ruleService := rule.NewService()
	alertService := alert.NewService()
	esConfigService := es_config.NewService()
	systemConfigService := system_config.NewService()

	// Start worker scheduler if enabled
	var sched *scheduler.Scheduler
	if config.AppConfig.Worker.Enabled {
		log.Println("Starting worker scheduler...")
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
			log.Fatalf("Failed to start scheduler: %v", err)
		}
		log.Printf("Worker scheduler started (check interval: %d seconds)", config.AppConfig.Worker.CheckInterval)
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
		log.Printf("API server starting on %s", addr)
		if err := r.Run(addr); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case err := <-serverErrChan:
		log.Fatalf("Server error: %v", err)
	}

	// Shutdown gracefully
	log.Println("Shutting down...")
	if sched != nil {
		sched.Stop()
		log.Println("Worker scheduler stopped")
	}
	log.Println("Shutdown complete")
}
