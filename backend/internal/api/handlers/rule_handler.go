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
	es_config "github.com/kk/elk-helper/backend/internal/service/esconfig"
	lark_config "github.com/kk/elk-helper/backend/internal/service/larkconfig"
	"github.com/kk/elk-helper/backend/internal/service/query"
	"github.com/kk/elk-helper/backend/internal/service/rule"
	"github.com/kk/elk-helper/backend/internal/worker/scheduler"
)

type RuleHandler struct {
	service           *rule.Service
	queryService      *query.Service
	esConfigService   *es_config.Service
	larkConfigService *lark_config.Service
}

func NewRuleHandler() *RuleHandler {
	queryService, _ := query.NewService()
	return &RuleHandler{
		service:           rule.NewService(),
		queryService:      queryService,
		esConfigService:   es_config.NewService(),
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
	// Backward compatible behavior:
	// - If page/page_size not provided, return all rules as before.
	// - If provided, return paginated response with pagination metadata.
	_, hasPage := c.GetQuery("page")
	_, hasPageSize := c.GetQuery("page_size")

	if !hasPage && !hasPageSize {
		rules, err := h.service.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": rules})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	rules, total, err := h.service.GetAllPaged(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rules,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (int(total) + pageSize - 1) / pageSize,
		},
	})
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

	// Trigger immediate execution if rule is enabled
	if rule.Enabled {
		if sched := scheduler.GetGlobalScheduler(); sched != nil {
			sched.TriggerRule(rule.ID)
		}
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

	// Trigger immediate execution if rule is enabled
	if updatedRule.Enabled {
		if sched := scheduler.GetGlobalScheduler(); sched != nil {
			sched.TriggerRule(uint(id))
		}
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

	// Trigger immediate execution if rule was just enabled
	if rule.Enabled {
		if sched := scheduler.GetGlobalScheduler(); sched != nil {
			sched.TriggerRule(uint(id))
		}
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
			"logs":  logs,
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
	CreatedCount int      `json:"created_count"`
	UpdatedCount int      `json:"updated_count"`
	SkippedCount int      `json:"skipped_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
	Details      []string `json:"details"` // Detailed info about each rule processed
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

// ImportRules imports rules from JSON with automatic deduplication
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
	var details []string
	var createdCount, updatedCount, skippedCount int

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

		// Check if rule with same name already exists (deduplication by name)
		existingRule, err := h.service.GetByName(rule.Name)
		if err == nil && existingRule != nil {
			// Rule exists - check if there are actual changes
			hasChanges := existingRule.IndexPattern != rule.IndexPattern ||
				existingRule.Interval != rule.Interval ||
				existingRule.Description != rule.Description ||
				existingRule.Enabled != rule.Enabled ||
				!compareQueryConditions(existingRule.Queries, rule.Queries) ||
				!compareOptionalUint(existingRule.ESConfigID, rule.ESConfigID) ||
				!compareOptionalUint(existingRule.LarkConfigID, rule.LarkConfigID) ||
				existingRule.LarkWebhook != rule.LarkWebhook

			if hasChanges {
				// Update existing rule with new data
				if err := h.service.Update(existingRule.ID, &rule); err != nil {
					errors = append(errors, fmt.Sprintf("Rule '%s': update failed - %v", rule.Name, err))
					continue
				}
				updatedCount++
				details = append(details, fmt.Sprintf("Rule '%s': updated (ID: %d)", rule.Name, existingRule.ID))

				// Trigger immediate execution if rule is enabled
				if rule.Enabled {
					if sched := scheduler.GetGlobalScheduler(); sched != nil {
						sched.TriggerRule(existingRule.ID)
					}
				}
			} else {
				// No changes, skip
				skippedCount++
				details = append(details, fmt.Sprintf("Rule '%s': skipped (no changes)", rule.Name))
			}
		} else {
			// Rule does not exist - create new
			if err := h.service.Create(&rule); err != nil {
				// Check for unique constraint violation
				if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
					errors = append(errors, fmt.Sprintf("Rule '%s': name already exists", rule.Name))
				} else {
					errors = append(errors, fmt.Sprintf("Rule '%s': create failed - %v", rule.Name, err))
				}
				continue
			}
			createdCount++
			details = append(details, fmt.Sprintf("Rule '%s': created (ID: %d)", rule.Name, rule.ID))

			// Trigger immediate execution if rule is enabled
			if rule.Enabled {
				if sched := scheduler.GetGlobalScheduler(); sched != nil {
					sched.TriggerRule(rule.ID)
				}
			}
		}
	}

	c.JSON(http.StatusOK, ImportRulesResponse{
		CreatedCount: createdCount,
		UpdatedCount: updatedCount,
		SkippedCount: skippedCount,
		ErrorCount:   len(errors),
		Errors:       errors,
		Details:      details,
	})
}

// compareOptionalUint compares two optional uint pointers
func compareOptionalUint(a, b *uint) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// compareQueryConditions compares two QueryConditions slices
func compareQueryConditions(a, b models.QueryConditions) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Field != b[i].Field ||
			a[i].Type != b[i].Type ||
			a[i].Operator != b[i].Operator ||
			a[i].Op != b[i].Op ||
			a[i].Logic != b[i].Logic {
			return false
		}
		// Compare Value - convert to string for comparison
		aVal := fmt.Sprintf("%v", a[i].Value)
		bVal := fmt.Sprintf("%v", b[i].Value)
		if aVal != bVal {
			return false
		}
	}
	return true
}
