// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/alert"
	"github.com/kk/elk-helper/backend/internal/service/systemconfig"
)

type SystemConfigHandler struct {
	service      *system_config.Service
	alertService *alert.Service
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{
		service:      system_config.NewService(),
		alertService: alert.NewService(),
	}
}

// GetCleanupConfig returns cleanup task configuration
// @Summary Get cleanup task configuration
// @Tags system-config
// @Produce json
// @Success 200 {object} models.CleanupConfig
// @Router /api/v1/system-config/cleanup [get]
func (h *SystemConfigHandler) GetCleanupConfig(c *gin.Context) {
	config, err := h.service.GetCleanupConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": config})
}

// UpdateCleanupConfig updates cleanup task configuration
// @Summary Update cleanup task configuration
// @Tags system-config
// @Accept json
// @Produce json
// @Param config body models.CleanupConfig true "Cleanup configuration"
// @Success 200 {object} models.CleanupConfig
// @Router /api/v1/system-config/cleanup [put]
func (h *SystemConfigHandler) UpdateCleanupConfig(c *gin.Context) {
	var config models.CleanupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateCleanupConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// ManualCleanup manually triggers cleanup of old alerts
// @Summary Manually trigger cleanup of old alerts
// @Tags system-config
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system-config/cleanup/manual [post]
func (h *SystemConfigHandler) ManualCleanup(c *gin.Context) {
	// Get cleanup configuration
	config, err := h.service.GetCleanupConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取清理配置失败: %v", err)})
		return
	}

	// Execute cleanup
	retentionDuration := time.Duration(config.RetentionDays) * 24 * time.Hour
	rowsAffected, err := h.alertService.CleanupOldData(retentionDuration)
	if err != nil {
		// Update execution status to failed
		_ = h.service.UpdateCleanupExecutionStatus("failed", fmt.Sprintf("清理失败: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("清理失败: %v", err)})
		return
	}

	// Update execution status to success
	resultMsg := fmt.Sprintf("成功删除 %d 条告警数据", rowsAffected)
	if rowsAffected == 0 {
		resultMsg = "没有需要清理的数据"
	}
	_ = h.service.UpdateCleanupExecutionStatus("success", resultMsg)

	c.JSON(http.StatusOK, gin.H{
		"message":        "清理完成",
		"deleted_count":  rowsAffected,
		"retention_days": config.RetentionDays,
	})
}

