// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kk/elk-helper/backend/internal/models"
	"gorm.io/gorm"
)

// ErrSSOOnlyLogin is returned when a user must sign in via SSO.
var ErrSSOOnlyLogin = errors.New("this account uses SSO login; password login is not available")

// FindOrCreateSSOUser finds an existing user or creates one with the default user role.
func (s *Service) FindOrCreateSSOUser(username, source string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("SSO username is empty")
	}
	if source == "" {
		source = "sso"
	}

	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err == nil {
		if !user.Enabled {
			return nil, ErrUserDisabled
		}
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	user = models.User{
		Username:   username,
		Role:       models.RoleUser,
		Enabled:    true,
		AuthSource: source,
	}
	placeholder, err := randomPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate placeholder password: %w", err)
	}
	if err := user.HashPassword(placeholder); err != nil {
		return nil, fmt.Errorf("failed to hash placeholder password: %w", err)
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create SSO user: %w", err)
	}
	return &user, nil
}

// IssueSessionToken updates last login and returns a JWT for the user.
func (s *Service) IssueSessionToken(user *models.User) (string, error) {
	if !user.Enabled {
		return "", ErrUserDisabled
	}
	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}
	now := time.Now()
	if err := s.db.Model(user).Update("last_login_at", now).Error; err != nil {
		return "", fmt.Errorf("failed to update last login: %w", err)
	}
	user.LastLoginAt = &now
	return token, nil
}

func randomPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isLocalAuthSource(source string) bool {
	return source == "" || source == models.AuthSourceLocal
}
