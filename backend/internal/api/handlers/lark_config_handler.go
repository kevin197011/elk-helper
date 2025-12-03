// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/larkconfig"
)

type LarkConfigHandler struct {
	service *lark_config.Service
}

func NewLarkConfigHandler() *LarkConfigHandler {
	return &LarkConfigHandler{
		service: lark_config.NewService(),
	}
}

// GetLarkConfigs returns all Lark configurations
// @Summary Get all Lark configurations
// @Tags lark-configs
// @Produce json
// @Success 200 {array} models.LarkConfig
// @Router /api/v1/lark-configs [get]
func (h *LarkConfigHandler) GetLarkConfigs(c *gin.Context) {
	configs, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// GetLarkConfig returns a Lark config by ID
// @Summary Get Lark config by ID
// @Tags lark-configs
// @Param id path int true "Config ID"
// @Produce json
// @Success 200 {object} models.LarkConfig
// @Router /api/v1/lark-configs/{id} [get]
func (h *LarkConfigHandler) GetLarkConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	config, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// CreateLarkConfig creates a new Lark configuration
// @Summary Create a new Lark configuration
// @Tags lark-configs
// @Accept json
// @Produce json
// @Param config body models.LarkConfig true "Lark Config data"
// @Success 201 {object} models.LarkConfig
// @Router /api/v1/lark-configs [post]
func (h *LarkConfigHandler) CreateLarkConfig(c *gin.Context) {
	var config models.LarkConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdateLarkConfig updates an existing Lark configuration
// @Summary Update a Lark configuration
// @Tags lark-configs
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param config body models.LarkConfig true "Lark Config data"
// @Success 200 {object} models.LarkConfig
// @Router /api/v1/lark-configs/{id} [put]
func (h *LarkConfigHandler) UpdateLarkConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	var config models.LarkConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Update(uint(id), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedConfig, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedConfig})
}

// DeleteLarkConfig deletes a Lark configuration
// @Summary Delete a Lark configuration
// @Tags lark-configs
// @Param id path int true "Config ID"
// @Success 204
// @Router /api/v1/lark-configs/{id} [delete]
func (h *LarkConfigHandler) DeleteLarkConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// TestLarkConfig tests the Lark webhook connection
// @Summary Test Lark configuration connection
// @Tags lark-configs
// @Param id path int true "Config ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/lark-configs/{id}/test [post]
func (h *LarkConfigHandler) TestLarkConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	config, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Test webhook by sending a test message
	testMessage := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": "测试消息：ELK Helper 连接测试",
		},
	}

	body, err := json.Marshal(testMessage)
	if err != nil {
		h.service.UpdateTestResult(uint(id), "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		h.service.UpdateTestResult(uint(id), "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.service.UpdateTestResult(uint(id), "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.service.UpdateTestResult(uint(id), "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if resp.StatusCode == http.StatusOK {
		if code, ok := result["code"].(float64); ok && code == 0 {
			h.service.UpdateTestResult(uint(id), "success", "")
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Webhook 测试成功",
			})
			return
		}
	}

	errMsg := "Lark API 返回错误"
	if msg, ok := result["msg"].(string); ok {
		errMsg = msg
	}
	h.service.UpdateTestResult(uint(id), "failed", errMsg)
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   errMsg,
	})
}

// SetDefaultLarkConfig sets a configuration as default
// @Summary Set Lark configuration as default
// @Tags lark-configs
// @Param id path int true "Config ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/lark-configs/{id}/set-default [post]
func (h *LarkConfigHandler) SetDefaultLarkConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	if err := h.service.SetDefault(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

