// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/rule"
)

type StatusHandler struct {
	ruleService   *rule.Service
	alertService  *alert.Service
	esConfigService *es_config.Service
}

func NewStatusHandler(ruleService *rule.Service, alertService *alert.Service, esConfigService *es_config.Service) *StatusHandler {
	return &StatusHandler{
		ruleService:     ruleService,
		alertService:    alertService,
		esConfigService: esConfigService,
	}
}

// GetStatus returns system status
// @Summary Get system status
// @Tags status
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/status [get]
func (h *StatusHandler) GetStatus(c *gin.Context) {
	// Get rule statistics
	rules, err := h.ruleService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enabledCount := 0
	for _, r := range rules {
		if r.Enabled {
			enabledCount++
		}
	}

	// Get ES status based on data source configurations
	esStatusInfo := h.getESStatusInfo()

	// Get recent alert stats
	stats, _ := h.alertService.GetStats(24 * time.Hour)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"rules": gin.H{
				"total":   len(rules),
				"enabled": enabledCount,
			},
			"elasticsearch": esStatusInfo,
			"alerts_24h": stats,
		},
	})
}

// getESStatusInfo returns ES connection status information for all data sources
func (h *StatusHandler) getESStatusInfo() gin.H {
	// Get all ES configurations
	configs, err := h.esConfigService.GetAll()
	if err != nil {
		return gin.H{
			"status":        "unknown",
			"total":         0,
			"success_count": 0,
			"failed_count":  0,
			"unknown_count": 0,
			"details":       []gin.H{},
		}
	}

	// Filter enabled configurations
	var enabledConfigs []models.ESConfig
	for _, config := range configs {
		if config.Enabled {
			enabledConfigs = append(enabledConfigs, config)
		}
	}

	// If no enabled configurations
	if len(enabledConfigs) == 0 {
		return gin.H{
			"status":        "not_configured",
			"total":         0,
			"success_count": 0,
			"failed_count":  0,
			"unknown_count": 0,
			"details":       []gin.H{},
		}
	}

	// Count statuses and build details
	successCount := 0
	failedCount := 0
	unknownCount := 0
	var details []gin.H

	for _, config := range enabledConfigs {
		var status string
		switch config.TestStatus {
		case "success":
			successCount++
			status = "success"
		case "failed":
			failedCount++
			status = "failed"
		default:
			unknownCount++
			status = "unknown"
		}

		details = append(details, gin.H{
			"id":     config.ID,
			"name":   config.Name,
			"url":    config.URL,
			"status": status,
		})
	}

	// Determine overall status
	var overallStatus string
	if successCount > 0 {
		overallStatus = "connected"
	} else if failedCount > 0 {
		overallStatus = "disconnected"
	} else {
		overallStatus = "unknown"
	}

	return gin.H{
		"status":        overallStatus,
		"total":         len(enabledConfigs),
		"success_count": successCount,
		"failed_count":  failedCount,
		"unknown_count": unknownCount,
		"details":       details,
	}
}
