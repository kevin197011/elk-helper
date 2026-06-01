// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"time"

	"gorm.io/gorm"
)

const SSOTypeOIDC = "oidc"

// SSOProvider stores OIDC IdP configuration.
type SSOProvider struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Type       string `gorm:"not null;default:'oidc'" json:"type"`
	Name       string `gorm:"not null" json:"name"`
	Enabled    bool   `gorm:"default:true" json:"enabled"`
	ConfigJSON string `gorm:"column:config_json;type:text;not null" json:"config_json"`
}

// TableName specifies the table name for SSOProvider.
func (SSOProvider) TableName() string {
	return "sso_providers"
}
