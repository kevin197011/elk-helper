// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package routes

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/api/handlers"
	"github.com/kk/elk-helper/backend/internal/api/middleware"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/repository/database"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	"github.com/kk/elk-helper/backend/internal/service/auth"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	"golang.org/x/time/rate"
)

// SetupRoutes configures all routes
func SetupRoutes(r *gin.Engine) {
	// Middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	// Health check (support both GET and HEAD for Docker health checks)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.Status(200)
	})

	// Initialize auth service
	authService := auth.NewService(database.DB, config.AppConfig.Auth.JWTSecret)

	// Initialize default admin user if no users exist
	if err := authService.InitDefaultAdmin(); err != nil {
		// Log error but don't fail startup
		slog.Warn("Failed to initialize default admin", "error", err)
	}

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		authHandler := handlers.NewAuthHandler(authService)
		authRoutes := v1.Group("/auth")
		{
			limiter := middleware.NewIPRateLimiter(
				rate.Limit(float64(config.AppConfig.Auth.LoginRateLimitPerMinute)/60.0),
				config.AppConfig.Auth.LoginRateLimitBurst,
				10*time.Minute,
			)
			authRoutes.POST("/login", middleware.RateLimitMiddleware(limiter), authHandler.Login)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Auth routes that require authentication
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.PUT("/auth/password", authHandler.UpdatePassword)

			// Services for status handler
			ruleService := rule.NewService()
			alertService := alert.NewService()

			// Rule routes
			ruleHandler := handlers.NewRuleHandler()
			rules := protected.Group("/rules")
			{
				rules.GET("", ruleHandler.GetRules)
				rules.GET("/:id", ruleHandler.GetRule)
				rules.POST("", ruleHandler.CreateRule)
				rules.PUT("/:id", ruleHandler.UpdateRule)
				rules.DELETE("/:id", ruleHandler.DeleteRule)
				rules.POST("/:id/toggle", ruleHandler.ToggleRuleEnabled)
				rules.POST("/:id/clone", ruleHandler.CloneRule)
				rules.POST("/test", ruleHandler.TestRule)
				rules.POST("/batch-delete", ruleHandler.BatchDeleteRules)
				rules.GET("/export", ruleHandler.ExportRules)
				rules.POST("/import", ruleHandler.ImportRules)
			}

			// Alert routes
			alertHandler := handlers.NewAlertHandler()
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", alertHandler.GetAlerts)
				alerts.GET("/stats", alertHandler.GetStats)
				alerts.GET("/rule-stats", alertHandler.GetRuleAlertStats)
				alerts.GET("/rule-timeseries", alertHandler.GetRuleTimeSeriesStats)
				alerts.GET("/:id", alertHandler.GetAlert)
				alerts.DELETE("/:id", alertHandler.DeleteAlert)
				alerts.POST("/batch-delete", alertHandler.BatchDeleteAlerts)
			}

			// Status routes
			esConfigService := es_config.NewService()
			statusHandler := handlers.NewStatusHandler(ruleService, alertService, esConfigService)
			status := protected.Group("/status")
			{
				status.GET("", statusHandler.GetStatus)
			}

			// ES Config routes
			esConfigHandler := handlers.NewESConfigHandler()
			esConfigs := protected.Group("/es-configs")
			{
				esConfigs.GET("", esConfigHandler.GetESConfigs)
				esConfigs.GET("/:id", esConfigHandler.GetESConfig)
				esConfigs.POST("", esConfigHandler.CreateESConfig)
				esConfigs.PUT("/:id", esConfigHandler.UpdateESConfig)
				esConfigs.DELETE("/:id", esConfigHandler.DeleteESConfig)
				esConfigs.POST("/:id/test", esConfigHandler.TestESConfig)
				esConfigs.POST("/:id/set-default", esConfigHandler.SetDefaultESConfig)
			}

			// Lark Config routes
			larkConfigHandler := handlers.NewLarkConfigHandler()
			larkConfigs := protected.Group("/lark-configs")
			{
				larkConfigs.GET("", larkConfigHandler.GetLarkConfigs)
				larkConfigs.GET("/:id", larkConfigHandler.GetLarkConfig)
				larkConfigs.POST("", larkConfigHandler.CreateLarkConfig)
				larkConfigs.PUT("/:id", larkConfigHandler.UpdateLarkConfig)
				larkConfigs.DELETE("/:id", larkConfigHandler.DeleteLarkConfig)
				larkConfigs.POST("/:id/test", larkConfigHandler.TestLarkConfig)
				larkConfigs.POST("/:id/set-default", larkConfigHandler.SetDefaultLarkConfig)
			}

			// System Config routes
			systemConfigHandler := handlers.NewSystemConfigHandler()
			systemConfigs := protected.Group("/system-config")
			{
				systemConfigs.GET("/cleanup", systemConfigHandler.GetCleanupConfig)
				systemConfigs.PUT("/cleanup", systemConfigHandler.UpdateCleanupConfig)
				systemConfigs.POST("/cleanup/manual", systemConfigHandler.ManualCleanup)
			}
		}
	}
}
