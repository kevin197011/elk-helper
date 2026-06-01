// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/sso"
)

// SSOAdminHandler manages SSO provider configuration (admin only).
type SSOAdminHandler struct {
	ssoService *sso.Service
}

// NewSSOAdminHandler creates an SSO admin handler.
func NewSSOAdminHandler(ssoService *sso.Service) *SSOAdminHandler {
	return &SSOAdminHandler{ssoService: ssoService}
}

// SSOProviderRequest is the create/update payload for SSO providers.
type SSOProviderRequest struct {
	Type    string                 `json:"type"`
	Name    string                 `json:"name" binding:"required"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config" binding:"required"`
}

// List GET /api/v1/sso/providers
func (h *SSOAdminHandler) List(c *gin.Context) {
	list, err := h.ssoService.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// Create POST /api/v1/sso/providers
func (h *SSOAdminHandler) Create(c *gin.Context) {
	var req SSOProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	providerType := req.Type
	if providerType == "" {
		providerType = models.SSOTypeOIDC
	}
	if providerType != models.SSOTypeOIDC {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be oidc"})
		return
	}
	if msg := sso.ValidateProviderConfig(providerType, req.Config); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	raw, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to serialize config"})
		return
	}
	p := &models.SSOProvider{
		Type:       providerType,
		Name:       req.Name,
		Enabled:    req.Enabled,
		ConfigJSON: string(raw),
	}
	created, err := h.ssoService.CreateProvider(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": created.ID}})
}

// Update PUT /api/v1/sso/providers/:id
func (h *SSOAdminHandler) Update(c *gin.Context) {
	id, ok := parseAdminSSOID(c)
	if !ok {
		return
	}
	var req SSOProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exist, err := h.ssoService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if msg := sso.ValidateProviderConfig(exist.Type, req.Config); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	raw, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to serialize config"})
		return
	}
	exist.Name = req.Name
	exist.Enabled = req.Enabled
	exist.ConfigJSON = string(raw)
	if err := h.ssoService.UpdateProvider(exist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

// Toggle POST /api/v1/sso/providers/:id/toggle
func (h *SSOAdminHandler) Toggle(c *gin.Context) {
	id, ok := parseAdminSSOID(c)
	if !ok {
		return
	}
	exist, err := h.ssoService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	enabled := !exist.Enabled
	if err := h.ssoService.SetEnabled(id, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": enabled}})
}

// Delete DELETE /api/v1/sso/providers/:id
func (h *SSOAdminHandler) Delete(c *gin.Context) {
	id, ok := parseAdminSSOID(c)
	if !ok {
		return
	}
	if err := h.ssoService.DeleteProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func parseAdminSSOID(c *gin.Context) (uint, bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return 0, false
	}
	return uint(id64), true
}
