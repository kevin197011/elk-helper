// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/esconfig"
	"github.com/kk/elk-helper/backend/internal/service/larkconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
)

type RuleHandler struct {
	service         *rule.Service
	queryService    *query.Service
	esConfigService *es_config.Service
	larkConfigService *lark_config.Service
}

func NewRuleHandler() *RuleHandler {
	queryService, _ := query.NewService()
	return &RuleHandler{
		service:          rule.NewService(),
		queryService:     queryService,
		esConfigService:  es_config.NewService(),
		larkConfigService: lark_config.NewService(),
	}
}

// GetRules returns all rules
// @Summary Get all rules
// @Tags rules
// @Produce json
// @Success 200 {array} models.Rule
// @Router /api/v1/rules [get]
func (h *RuleHandler) GetRules(c *gin.Context) {
	rules, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// GetRule returns a rule by ID
// @Summary Get rule by ID
// @Tags rules
// @Param id path int true "Rule ID"
// @Produce json
// @Success 200 {object} models.Rule
// @Router /api/v1/rules/{id} [get]
func (h *RuleHandler) GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// CreateRule creates a new rule
// @Summary Create a new rule
// @Tags rules
// @Accept json
// @Produce json
// @Param rule body models.Rule true "Rule data"
// @Success 201 {object} models.Rule
// @Router /api/v1/rules [post]
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

// UpdateRule updates an existing rule
// @Summary Update a rule
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body models.Rule true "Rule data"
// @Success 200 {object} models.Rule
// @Router /api/v1/rules/{id} [put]
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Update(uint(id), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedRule, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updatedRule})
}

// DeleteRule deletes a rule
// @Summary Delete a rule
// @Tags rules
// @Param id path int true "Rule ID"
// @Success 204
// @Router /api/v1/rules/{id} [delete]
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ToggleRuleEnabled toggles the enabled status of a rule
// @Summary Toggle rule enabled status
// @Tags rules
// @Param id path int true "Rule ID"
// @Produce json
// @Success 200 {object} models.Rule
// @Router /api/v1/rules/{id}/toggle [post]
func (h *RuleHandler) ToggleRuleEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	if err := h.service.ToggleEnabled(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// TestRule tests a rule's query without saving
// @Summary Test rule query
// @Tags rules
// @Accept json
// @Produce json
// @Param rule body models.Rule true "Rule to test"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rules/test [post]
func (h *RuleHandler) TestRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get query service based on rule's ES config
	var queryService *query.Service
	var err error

	if rule.ESConfigID != nil {
		// Load ES config if not already loaded
		var esConfig *models.ESConfig
		if rule.ESConfig != nil {
			esConfig = rule.ESConfig
			// If ESConfig was loaded from rule (likely from frontend), password might be empty
			// Need to reload from database to get actual password
			if esConfig.Password == "" {
				esConfig, err = h.esConfigService.GetByID(*rule.ESConfigID)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "ES config not found",
						"success": false,
					})
					return
				}
			}
		} else {
			esConfig, err = h.esConfigService.GetByID(*rule.ESConfigID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "ES config not found",
					"success": false,
				})
				return
			}
		}

		if !esConfig.Enabled {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "ES config is disabled",
				"success": false,
			})
			return
		}

		// Check if credentials are provided when ES requires authentication
		if esConfig.Username == "" || esConfig.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("ES 数据源配置缺少认证信息。请检查数据源配置（ID: %d）中的用户名和密码是否正确填写。", *rule.ESConfigID),
				"success": false,
			})
			return
		}

		queryService, err = query.NewServiceFromConfig(esConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	} else {
		// Use default query service (environment variables)
		queryService = h.queryService
		if queryService == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "未配置 Elasticsearch 数据源。请先在规则中选择一个 ES 数据源配置。",
				"success": false,
			})
			return
		}
	}

	// Test query with last 10 minutes
	toTime := time.Now()
	fromTime := toTime.Add(-10 * time.Minute)

	logs, err := queryService.QueryLogs(c.Request.Context(), &rule, fromTime, toTime, 100)
	if err != nil {
		errorMsg := err.Error()
		// Provide more helpful error message for 401 errors
		if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "Unauthorized") || strings.Contains(errorMsg, "missing authentication credentials") {
			if rule.ESConfigID == nil {
				errorMsg = "认证失败: 请先为规则选择一个 Elasticsearch 数据源配置，并确保该配置中已填写用户名和密码。"
			} else {
				errorMsg = "认证失败: Elasticsearch 需要认证。请检查数据源配置中的用户名和密码是否正确。"
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count": len(logs),
			"logs": logs,
			"time_range": gin.H{
				"from": fromTime.Format(time.RFC3339),
				"to":   toTime.Format(time.RFC3339),
			},
		},
	})
}

// BatchDeleteRules deletes multiple rules
// @Summary Batch delete rules
// @Tags rules
// @Accept json
// @Produce json
// @Param ids body []uint true "Rule IDs to delete"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rules/batch-delete [post]
func (h *RuleHandler) BatchDeleteRules(c *gin.Context) {
	var ids struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var errors []string
	var successCount int

	for _, id := range ids.IDs {
		if err := h.service.Delete(id); err != nil {
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"error_count":   len(errors),
		"errors":        errors,
	})
}

// ExportRules exports all rules as JSON
// @Summary Export all rules
// @Tags rules
// @Produce application/json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rules/export [get]
func (h *RuleHandler) ExportRules(c *gin.Context) {
	rules, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clean rules for export (remove runtime data and test information)
	cleanRules := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		cleanRule := map[string]interface{}{
			"name":          rule.Name,
			"index_pattern": rule.IndexPattern,
			"queries":       rule.Queries,
			"enabled":       rule.Enabled,
			"interval":      rule.Interval,
			"description":   rule.Description,
		}

		// Add ES config reference (only ID and name, no sensitive/test data)
		if rule.ESConfigID != nil {
			cleanRule["es_config_id"] = *rule.ESConfigID
			if rule.ESConfig != nil {
				cleanRule["es_config"] = map[string]interface{}{
					"id":          rule.ESConfig.ID,
					"name":        rule.ESConfig.Name,
					"url":         rule.ESConfig.URL,
					"username":    rule.ESConfig.Username,
					"use_ssl":     rule.ESConfig.UseSSL,
					"skip_verify": rule.ESConfig.SkipVerify,
					"enabled":     rule.ESConfig.Enabled,
					"description": rule.ESConfig.Description,
					"is_default":  rule.ESConfig.IsDefault,
					// Exclude: Password, CACertificate, LastTestAt, TestStatus, TestError
				}
			}
		}

		// Add Lark config reference (only ID and name, no test data)
		if rule.LarkConfigID != nil {
			cleanRule["lark_config_id"] = *rule.LarkConfigID
			if rule.LarkConfig != nil {
				cleanRule["lark_config"] = map[string]interface{}{
					"id":          rule.LarkConfig.ID,
					"name":        rule.LarkConfig.Name,
					"webhook_url": rule.LarkConfig.WebhookURL,
					"enabled":     rule.LarkConfig.Enabled,
					"description": rule.LarkConfig.Description,
					"is_default":  rule.LarkConfig.IsDefault,
					// Exclude: LastTestAt, TestStatus, TestError
				}
			}
		}

		// Keep lark_webhook for backward compatibility
		if rule.LarkWebhook != "" {
			cleanRule["lark_webhook"] = rule.LarkWebhook
		}

		// Exclude: ID, CreatedAt, UpdatedAt, LastRunTime, RunCount, AlertCount (runtime statistics)

		cleanRules = append(cleanRules, cleanRule)
	}

	// Create export data structure
	exportData := map[string]interface{}{
		"version":     "1.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"rules":       cleanRules,
	}

	// Set headers for file download
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=rules_export_%s.json", time.Now().Format("20060102_150405")))

	c.JSON(http.StatusOK, exportData)
}

// ImportRulesRequest represents the import request
type ImportRulesRequest struct {
	Rules []models.Rule `json:"rules" binding:"required"`
}

// ImportRulesResponse represents the import response
type ImportRulesResponse struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
}

// CloneRule clones an existing rule with a new name
// @Summary Clone a rule
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID to clone"
// @Param data body map[string]string true "New rule name"
// @Success 201 {object} models.Rule
// @Router /api/v1/rules/{id}/clone [post]
func (h *RuleHandler) CloneRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Clone the rule
	clonedRule, err := h.service.Clone(uint(id), req.Name)
	if err != nil {
		// Check if it's a duplicate name error
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusConflict, gin.H{"error": "规则名称已存在，请使用其他名称"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": clonedRule})
}

// ImportRules imports rules from JSON
// @Summary Import rules from JSON
// @Tags rules
// @Accept json
// @Produce json
// @Param data body ImportRulesRequest true "Rules to import"
// @Success 200 {object} ImportRulesResponse
// @Router /api/v1/rules/import [post]
func (h *RuleHandler) ImportRules(c *gin.Context) {
	var req ImportRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no rules to import"})
		return
	}

	var errors []string
	var successCount int

	for i, rule := range req.Rules {
		// Reset ID and timestamps for import (will be auto-generated/updated)
		rule.ID = 0
		rule.CreatedAt = time.Time{}
		rule.UpdatedAt = time.Time{}

		// Validate required fields
		if rule.Name == "" {
			errors = append(errors, fmt.Sprintf("Rule #%d: name is required", i+1))
			continue
		}
		if rule.IndexPattern == "" {
			errors = append(errors, fmt.Sprintf("Rule '%s': index_pattern is required", rule.Name))
			continue
		}

		// Resolve ES config by name if es_config is provided, otherwise use es_config_id
		if rule.ESConfig != nil && rule.ESConfig.Name != "" {
			// Try to find ES config by name
			esConfig, err := h.esConfigService.GetByName(rule.ESConfig.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Rule '%s': ES config '%s' not found", rule.Name, rule.ESConfig.Name))
				continue
			}
			// Reset ESConfigID to the found config's ID
			rule.ESConfigID = &esConfig.ID
			rule.ESConfig = nil // Clear to avoid foreign key issues
		} else if rule.ESConfigID != nil {
			// Verify ES config ID exists
			_, err := h.esConfigService.GetByID(*rule.ESConfigID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Rule '%s': ES config ID %d not found", rule.Name, *rule.ESConfigID))
				continue
			}
		}

		// Resolve Lark config by name if lark_config is provided, otherwise use lark_config_id
		if rule.LarkConfig != nil && rule.LarkConfig.Name != "" {
			// Try to find Lark config by name
			larkConfig, err := h.larkConfigService.GetByName(rule.LarkConfig.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Rule '%s': Lark config '%s' not found", rule.Name, rule.LarkConfig.Name))
				continue
			}
			// Reset LarkConfigID to the found config's ID
			rule.LarkConfigID = &larkConfig.ID
			rule.LarkConfig = nil // Clear to avoid foreign key issues
		} else if rule.LarkConfigID != nil {
			// Verify Lark config ID exists
			_, err := h.larkConfigService.GetByID(*rule.LarkConfigID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Rule '%s': Lark config ID %d not found", rule.Name, *rule.LarkConfigID))
				continue
			}
		}

		// Create or update rule (check if name exists)
		existingRules, err := h.service.GetAll()
		if err == nil {
			// Check if rule with same name exists
			var existingRule *models.Rule
			for _, r := range existingRules {
				if r.Name == rule.Name {
					existingRule = &r
					break
				}
			}

			if existingRule != nil {
				// Update existing rule
				rule.ID = existingRule.ID
				if err := h.service.Update(existingRule.ID, &rule); err != nil {
					errors = append(errors, fmt.Sprintf("Rule '%s': update failed - %v", rule.Name, err))
					continue
				}
				successCount++
			} else {
				// Create new rule
				if err := h.service.Create(&rule); err != nil {
					errors = append(errors, fmt.Sprintf("Rule '%s': create failed - %v", rule.Name, err))
					continue
				}
				successCount++
			}
		} else {
			// If GetAll fails, try to create directly
			if err := h.service.Create(&rule); err != nil {
				errors = append(errors, fmt.Sprintf("Rule '%s': create failed - %v", rule.Name, err))
				continue
			}
			successCount++
		}
	}

	c.JSON(http.StatusOK, ImportRulesResponse{
		SuccessCount: successCount,
		ErrorCount:   len(errors),
		Errors:       errors,
	})
}

