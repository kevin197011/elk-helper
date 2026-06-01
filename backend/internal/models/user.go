// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole represents user role
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"

	// AuthSourceLocal marks users that sign in with username/password.
	AuthSourceLocal = "local"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username    string     `gorm:"not null;uniqueIndex" json:"username"`         // 用户名
	Password    string     `json:"-"`                                            // 密码（SSO 用户可为空）
	Email       string     `gorm:"uniqueIndex" json:"email,omitempty"`           // 邮箱
	Role        UserRole   `gorm:"default:'user'" json:"role"`                   // 角色：admin, user
	Enabled     bool       `gorm:"default:true" json:"enabled"`                  // 是否启用
	AuthSource  string     `gorm:"default:'local'" json:"auth_source,omitempty"` // local 或 oidc:<id>
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`                      // 最后登录时间
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// HashPassword hashes the password using bcrypt
func (u *User) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// CheckPassword checks if the provided password matches the hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
