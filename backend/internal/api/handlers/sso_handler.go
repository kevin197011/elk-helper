// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/service/auth"
	"github.com/kk/elk-helper/backend/internal/service/sso"
)

// SSOHandler handles public SSO login flows.
type SSOHandler struct {
	ssoService  *sso.Service
	authService *auth.Service
}

// NewSSOHandler creates an SSO handler.
func NewSSOHandler(ssoService *sso.Service, authService *auth.Service) *SSOHandler {
	return &SSOHandler{
		ssoService:  ssoService,
		authService: authService,
	}
}

// ListProviders GET /api/v1/auth/sso/providers
func (h *SSOHandler) ListProviders(c *gin.Context) {
	providers, err := h.ssoService.ListProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": providers})
}

// OIDCLogin GET /api/v1/auth/sso/oidc/:id/login
func (h *SSOHandler) OIDCLogin(c *gin.Context) {
	id, ok := parseSSOProviderID(c)
	if !ok {
		return
	}
	state, err := sso.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	authURL, err := h.ssoService.BuildOIDCAuthURL(id, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetCookie(oidcStateCookieName(id), state, 600, "/", "", secure, true)
	c.Redirect(http.StatusFound, authURL)
}

// OIDCCallback GET /api/v1/auth/sso/oidc/:id/callback
func (h *SSOHandler) OIDCCallback(c *gin.Context) {
	id, ok := parseSSOProviderID(c)
	if !ok {
		return
	}
	frontendBase := config.AppConfig.Auth.SSOFrontendBaseURL

	if errStr := c.Query("error"); errStr != "" {
		desc := c.Query("error_description")
		h.redirectSSOError(c, frontendBase, errStr+": "+desc)
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		h.redirectSSOError(c, frontendBase, "missing code or state")
		return
	}

	expectedState, _ := c.Cookie(oidcStateCookieName(id))
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	if expectedState == "" || expectedState != state {
		h.redirectSSOError(c, frontendBase, "invalid OIDC state")
		return
	}
	c.SetCookie(oidcStateCookieName(id), "", -1, "/", "", secure, true)

	username, err := h.ssoService.ExchangeOIDC(c.Request.Context(), id, code)
	if err != nil {
		h.redirectSSOError(c, frontendBase, err.Error())
		return
	}

	source := fmt.Sprintf("oidc:%d", id)
	user, err := h.authService.FindOrCreateSSOUser(username, source)
	if err != nil {
		h.redirectSSOError(c, frontendBase, err.Error())
		return
	}

	token, err := h.authService.IssueSessionToken(user)
	if err != nil {
		slog.Error("SSO issue token failed", "error", err)
		h.redirectSSOError(c, frontendBase, "failed to issue session token")
		return
	}

	q := url.Values{}
	q.Set("token", token)
	q.Set("username", user.Username)
	q.Set("role", string(user.Role))
	c.Redirect(http.StatusFound, frontendBase+"/sso-callback?"+q.Encode())
}

func (h *SSOHandler) redirectSSOError(c *gin.Context, frontendBase, msg string) {
	c.Redirect(http.StatusFound, frontendBase+"/sso-callback?error="+url.QueryEscape(msg))
}

func parseSSOProviderID(c *gin.Context) (uint, bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return 0, false
	}
	return uint(id64), true
}

func oidcStateCookieName(id uint) string {
	return fmt.Sprintf("oidc_state_%d", id)
}
