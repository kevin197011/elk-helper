// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"time"

	"gorm.io/gorm"
)

// ESConfig represents Elasticsearch data source configuration
type ESConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name            string `gorm:"not null;uniqueIndex" json:"name"`        // 配置名称
	URL             string `gorm:"not null" json:"url"`                     // ES 地址
	Username        string `json:"username,omitempty"`                      // 用户名（可选）
	Password        string `gorm:"type:text" json:"-"`                      // 密码（不返回）
	PasswordEnc     string `gorm:"type:text" json:"-"`                      // 加密后的密码（未来扩展）
	UseSSL          bool   `gorm:"default:false" json:"use_ssl"`            // 是否使用 SSL/TLS
	SkipVerify      bool   `gorm:"default:false" json:"skip_verify"`        // 是否跳过证书验证（仅用于开发/测试）
	CACertificate   string `gorm:"type:text" json:"-"`                      // CA 证书内容（不返回）
	IsDefault       bool   `gorm:"default:false" json:"is_default"`         // 是否为默认配置
	Description     string `json:"description,omitempty"`                   // 描述
	Enabled         bool   `gorm:"default:true" json:"enabled"`             // 是否启用
	LastTestAt      *time.Time `json:"last_test_at,omitempty"`              // 最后测试时间
	TestStatus      string `gorm:"default:unknown" json:"test_status"`      // 测试状态：unknown, success, failed
	TestError       string `gorm:"type:text" json:"test_error,omitempty"`   // 测试错误信息
}

// TableName specifies the table name for ESConfig
func (ESConfig) TableName() string {
	return "es_configs"
}

