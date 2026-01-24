// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/service/alert"
)

type AlertHandler struct {
	service *alert.Service
}

func NewAlertHandler() *AlertHandler {
	return &AlertHandler{
		service: alert.NewService(),
	}
}

// GetAlerts returns alerts with pagination
// @Summary Get alerts
// @Tags alerts
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/alerts [get]
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	alerts, total, err := h.service.GetAll(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

// GetAlert returns an alert by ID
// @Summary Get alert by ID
// @Tags alerts
// @Param id path int true "Alert ID"
// @Produce json
// @Success 200 {object} models.Alert
// @Router /api/v1/alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert ID"})
		return
	}

	alert, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alert})
}

// DeleteAlert deletes an alert
// @Summary Delete an alert
// @Tags alerts
// @Param id path int true "Alert ID"
// @Success 204
// @Router /api/v1/alerts/{id} [delete]
func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetStats returns alert statistics
// @Summary Get alert statistics
// @Tags alerts
// @Param duration query string false "Duration (e.g., 1h, 24h, 7d)" default(24h)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/alerts/stats [get]
func (h *AlertHandler) GetStats(c *gin.Context) {
	durationStr := c.DefaultQuery("duration", "24h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 24 * time.Hour
	}

	stats, err := h.service.GetStats(duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// BatchDeleteAlerts deletes multiple alerts
// @Summary Batch delete alerts
// @Tags alerts
// @Accept json
// @Produce json
// @Param ids body []uint true "Alert IDs to delete"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/alerts/batch-delete [post]
func (h *AlertHandler) BatchDeleteAlerts(c *gin.Context) {
	var ids struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BatchDelete(ids.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"deleted_count": len(ids.IDs),
	})
}

// GetRuleAlertStats returns alert statistics grouped by rule
// @Summary Get alert statistics by rule
// @Tags alerts
// @Param duration query string false "Duration (e.g., 1h, 24h, 7d)" default(24h)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/alerts/rule-stats [get]
func (h *AlertHandler) GetRuleAlertStats(c *gin.Context) {
	durationStr := c.DefaultQuery("duration", "24h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 24 * time.Hour
	}

	stats, err := h.service.GetRuleAlertStats(duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetRuleTimeSeriesStats returns time series alert statistics for top rules
// @Summary Get time series alert statistics by rule
// @Tags alerts
// @Param duration query string false "Duration (e.g., 1h, 24h, 7d)" default(24h)
// @Param interval query int false "Interval in minutes" default(60)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/alerts/rule-timeseries [get]
func (h *AlertHandler) GetRuleTimeSeriesStats(c *gin.Context) {
	durationStr := c.DefaultQuery("duration", "24h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = 24 * time.Hour
	}

	intervalMinutes, _ := strconv.Atoi(c.DefaultQuery("interval", "60"))
	if intervalMinutes < 10 {
		intervalMinutes = 10
	}
	if intervalMinutes > 360 {
		intervalMinutes = 360
	}

	stats, err := h.service.GetRuleTimeSeriesStats(duration, intervalMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
