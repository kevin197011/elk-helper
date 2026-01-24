// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/auth"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token    string       `json:"token"`
	User     *models.User `json:"user"`
	ExpiresAt string      `json:"expires_at"`
}

// Login handles user login
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		if err == auth.ErrUserDisabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
		ExpiresAt: "24h", // Token expires in 24 hours
	})
}

// Logout handles user logout
// @Summary User logout
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// We could implement token blacklisting here if needed
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// UpdatePasswordRequest represents password update request
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdatePassword handles password update
// @Summary Update user password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "Password update request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update password
	if err := h.authService.UpdatePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "incorrect old password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
			return
		}
		if err.Error() == "new password must be at least 6 characters" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度至少为 6 个字符"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码更新成功"})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} models.User
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userModel := user.(*models.User)
	userModel.Password = "" // Remove password

	c.JSON(http.StatusOK, gin.H{"data": userModel})
}

