// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"time"

	"gorm.io/gorm"
)

// LarkConfig represents Lark webhook configuration
type LarkConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string     `gorm:"not null;uniqueIndex" json:"name"`         // 配置名称
	WebhookURL  string     `gorm:"not null;type:text" json:"webhook_url"`    // Webhook URL
	IsDefault   bool       `gorm:"default:false" json:"is_default"`          // 是否为默认配置
	Description string     `json:"description,omitempty"`                    // 描述
	Enabled     bool       `gorm:"default:true" json:"enabled"`              // 是否启用
	LastTestAt  *time.Time `json:"last_test_at,omitempty"`                   // 最后测试时间
	TestStatus  string     `gorm:"default:unknown" json:"test_status"`       // 测试状态：unknown, success, failed
	TestError   string     `gorm:"type:text" json:"test_error,omitempty"`    // 测试错误信息
}

// TableName specifies the table name for LarkConfig
func (LarkConfig) TableName() string {
	return "lark_configs"
}

