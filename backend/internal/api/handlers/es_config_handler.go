// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
)

type ESConfigHandler struct {
	service      *es_config.Service
	queryService *query.Service
}

func NewESConfigHandler() *ESConfigHandler {
	queryService, _ := query.NewService()
	return &ESConfigHandler{
		service:      es_config.NewService(),
		queryService: queryService,
	}
}

// GetESConfigs returns all ES configurations
// @Summary Get all ES configurations
// @Tags es-configs
// @Produce json
// @Success 200 {array} models.ESConfig
// @Router /api/v1/es-configs [get]
func (h *ESConfigHandler) GetESConfigs(c *gin.Context) {
	configs, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove password from response
	for i := range configs {
		configs[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// GetESConfig returns an ES config by ID
// @Summary Get ES config by ID
// @Tags es-configs
// @Param id path int true "Config ID"
// @Produce json
// @Success 200 {object} models.ESConfig
// @Router /api/v1/es-configs/{id} [get]
func (h *ESConfigHandler) GetESConfig(c *gin.Context) {
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

	// Remove password from response
	config.Password = ""

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// CreateESConfig creates a new ES configuration
// @Summary Create a new ES configuration
// @Tags es-configs
// @Accept json
// @Produce json
// @Param config body models.ESConfig true "ES Config data"
// @Success 201 {object} models.ESConfig
// @Router /api/v1/es-configs [post]
func (h *ESConfigHandler) CreateESConfig(c *gin.Context) {
	// Read request body as map to properly handle password field
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Manually map fields from requestBody to config struct
	// This ensures password field is properly handled even if it's empty string
	config := models.ESConfig{}
	if name, ok := requestBody["name"].(string); ok {
		config.Name = name
	}
	if url, ok := requestBody["url"].(string); ok {
		config.URL = url
	}
	if username, ok := requestBody["username"].(string); ok {
		config.Username = username
	}
	// Always set password, even if empty string (to allow empty passwords)
	if pwd, ok := requestBody["password"].(string); ok {
		config.Password = pwd
	} else {
		// If password field is not provided, set to empty string
		config.Password = ""
	}
	if useSSL, ok := requestBody["use_ssl"].(bool); ok {
		config.UseSSL = useSSL
	}
	if skipVerify, ok := requestBody["skip_verify"].(bool); ok {
		config.SkipVerify = skipVerify
	}
	if caCert, ok := requestBody["ca_certificate"].(string); ok {
		config.CACertificate = caCert
	}
	if isDefault, ok := requestBody["is_default"].(bool); ok {
		config.IsDefault = isDefault
	}
	if description, ok := requestBody["description"].(string); ok {
		config.Description = description
	}
	if enabled, ok := requestBody["enabled"].(bool); ok {
		config.Enabled = enabled
	}

	if err := h.service.Create(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove password from response
	config.Password = ""

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdateESConfig updates an existing ES configuration
// @Summary Update an ES configuration
// @Tags es-configs
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param config body models.ESConfig true "ES Config data"
// @Success 200 {object} models.ESConfig
// @Router /api/v1/es-configs/{id} [put]
func (h *ESConfigHandler) UpdateESConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config ID"})
		return
	}

	// Read request body as map to check if password field exists
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing config to preserve password if not provided
	existingConfig, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	// Check if password field was provided in request
	_, passwordProvided := requestBody["password"]

	// Manually map fields from requestBody to config struct
	config := models.ESConfig{}
	if name, ok := requestBody["name"].(string); ok {
		config.Name = name
	}
	if url, ok := requestBody["url"].(string); ok {
		config.URL = url
	}
	if username, ok := requestBody["username"].(string); ok {
		config.Username = username
	}
	if passwordProvided {
		if pwd, ok := requestBody["password"].(string); ok {
			config.Password = pwd
		}
	} else {
		// Password field not provided, preserve existing
		config.Password = existingConfig.Password
	}
	if useSSL, ok := requestBody["use_ssl"].(bool); ok {
		config.UseSSL = useSSL
	}
	if skipVerify, ok := requestBody["skip_verify"].(bool); ok {
		config.SkipVerify = skipVerify
	}
	if caCert, ok := requestBody["ca_certificate"].(string); ok {
		config.CACertificate = caCert
	}
	if isDefault, ok := requestBody["is_default"].(bool); ok {
		config.IsDefault = isDefault
	}
	if description, ok := requestBody["description"].(string); ok {
		config.Description = description
	}
	if enabled, ok := requestBody["enabled"].(bool); ok {
		config.Enabled = enabled
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

	// Remove password from response
	updatedConfig.Password = ""

	c.JSON(http.StatusOK, gin.H{"data": updatedConfig})
}

// DeleteESConfig deletes an ES configuration
// @Summary Delete an ES configuration
// @Tags es-configs
// @Param id path int true "Config ID"
// @Success 204
// @Router /api/v1/es-configs/{id} [delete]
func (h *ESConfigHandler) DeleteESConfig(c *gin.Context) {
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

// TestESConfig tests the ES connection
// @Summary Test ES configuration connection
// @Tags es-configs
// @Param id path int true "Config ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/es-configs/{id}/test [post]
func (h *ESConfigHandler) TestESConfig(c *gin.Context) {
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

	// Don't validate credentials upfront - let the connection test determine if auth is needed
	// This allows testing ES instances with or without security enabled

	// Create query service using the config (this will handle SSL/TLS configuration)
	queryService, err := query.NewServiceFromConfig(config)
	if err != nil {
		h.service.UpdateTestResult(uint(id), "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Test connection with ping
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = queryService.TestConnection(ctx)
	if err != nil {
		errorMsg := err.Error()
		// Provide more helpful error message for 401 errors
		if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "Unauthorized") {
			if config.Username == "" || config.Password == "" {
				if config.Username == "" && config.Password == "" {
					errorMsg = "认证失败: 请配置用户名和密码。Elasticsearch 已启用安全认证，需要在数据源配置中输入用户名和密码。"
				} else if config.Username == "" {
					errorMsg = "认证失败: 请配置用户名。Elasticsearch 已启用安全认证，需要在数据源配置中输入用户名。"
				} else {
					errorMsg = "认证失败: 请配置密码。Elasticsearch 已启用安全认证，需要在数据源配置中输入密码（即使密码为空，也需要在前端明确输入）。"
				}
			} else {
				errorMsg = fmt.Sprintf("认证失败: 用户名或密码错误。请验证您的凭据是否正确。当前用户名: %s", config.Username)
			}
		}
		h.service.UpdateTestResult(uint(id), "failed", errorMsg)
		// 返回 200 状态码，在响应体中标记失败，这样前端可以统一处理
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   errorMsg,
		})
		return
	}

	// Success
	h.service.UpdateTestResult(uint(id), "success", "")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection test successful",
	})
}

// SetDefaultESConfig sets a configuration as default
// @Summary Set ES configuration as default
// @Tags es-configs
// @Param id path int true "Config ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/es-configs/{id}/set-default [post]
func (h *ESConfigHandler) SetDefaultESConfig(c *gin.Context) {
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
