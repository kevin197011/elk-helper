// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/kk/elk-helper/backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user is disabled")
)

// Service provides authentication services
type Service struct {
	db        *gorm.DB
	jwtSecret []byte
}

// NewService creates a new auth service
func NewService(db *gorm.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(username, password string) (string, *models.User, error) {
	var user models.User

	// Find user by username
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Check if user is enabled
	if !user.Enabled {
		return "", nil, ErrUserDisabled
	}

	// Verify password
	if !user.CheckPassword(password) {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(&user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	return token, &user, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(username, password, email string, role models.UserRole) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Role:     role,
		Enabled:  true,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// generateToken generates a JWT token for a user
func (s *Service) generateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "elk-helper",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// InitDefaultAdmin creates a default admin user if no users exist
func (s *Service) InitDefaultAdmin() error {
	var count int64
	s.db.Model(&models.User{}).Count(&count)

	if count == 0 {
		// Get admin credentials from environment variables, use defaults if not set
		adminUsername := getEnv("ADMIN_USERNAME", "admin")
		if adminUsername == "" {
			adminUsername = "admin"
		}

		adminPassword := getEnv("ADMIN_PASSWORD", "admin123")
		if adminPassword == "" {
			adminPassword = "admin123"
		}

		adminEmail := getEnv("ADMIN_EMAIL", "admin@example.com")
		if adminEmail == "" {
			adminEmail = "admin@example.com"
		}

		// Create default admin user
		adminUser := &models.User{
			Username: adminUsername,
			Email:    adminEmail,
			Role:     models.RoleAdmin,
			Enabled:  true,
		}

		if err := adminUser.HashPassword(adminPassword); err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		if err := s.db.Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create default admin: %w", err)
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// UpdatePassword updates user password
func (s *Service) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if !user.CheckPassword(oldPassword) {
		return fmt.Errorf("incorrect old password")
	}

	// Validate new password
	if len(newPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters")
	}

	// Update password
	if err := user.HashPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
