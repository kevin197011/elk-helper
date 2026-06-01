// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kk/elk-helper/backend/internal/models"
	"github.com/kk/elk-helper/backend/internal/service/auth"
)

// UserHandler handles user management (admin only).
type UserHandler struct {
	authService *auth.Service
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(authService *auth.Service) *UserHandler {
	return &UserHandler{authService: authService}
}

type createUserRequest struct {
	Username string          `json:"username" binding:"required,min=2,max=64"`
	Password string          `json:"password" binding:"required,min=6"`
	Email    string          `json:"email"`
	Role     models.UserRole `json:"role" binding:"required"`
	Enabled  *bool           `json:"enabled"`
}

type updateUserRequest struct {
	Email   *string          `json:"email"`
	Role    *models.UserRole `json:"role"`
	Enabled *bool            `json:"enabled"`
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func sanitizeUser(user *models.User) {
	if user == nil {
		return
	}
	user.Password = ""
}

func mapUserServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, auth.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, auth.ErrUsernameExists), errors.Is(err, auth.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, auth.ErrCannotDeleteSelf), errors.Is(err, auth.ErrLastAdmin), errors.Is(err, auth.ErrInvalidRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return true
}

// ListUsers returns all users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range users {
		sanitizeUser(&users[i])
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// GetUser returns one user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	user, err := h.authService.GetUserByID(id)
	if mapUserServiceError(c, err) {
		return
	}
	sanitizeUser(user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	user, err := h.authService.CreateUser(req.Username, req.Password, req.Email, req.Role)
	if mapUserServiceError(c, err) {
		return
	}

	if !enabled {
		enabledVal := false
		user, err = h.authService.UpdateUser(user.ID, auth.UpdateUserInput{Enabled: &enabledVal})
		if mapUserServiceError(c, err) {
			return
		}
	}

	sanitizeUser(user)
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// UpdateUser updates a user.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.UpdateUser(id, auth.UpdateUserInput{
		Email:   req.Email,
		Role:    req.Role,
		Enabled: req.Enabled,
	})
	if mapUserServiceError(c, err) {
		return
	}
	sanitizeUser(user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser deletes a user.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	actorID := c.GetUint("user_id")
	if err := h.authService.DeleteUser(actorID, id); mapUserServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ResetPassword resets a user's password (admin).
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResetPassword(id, req.NewPassword); mapUserServiceError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, err
	}
	return uint(id), nil
}
