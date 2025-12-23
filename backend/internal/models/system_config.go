// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig represents system configuration
type SystemConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Key   string `gorm:"not null;uniqueIndex" json:"key"`   // 配置键
	Value string `gorm:"type:text" json:"value"`            // 配置值（JSON 格式）
	Description string `gorm:"type:text" json:"description,omitempty"` // 描述
}

// TableName specifies the table name for SystemConfig
func (SystemConfig) TableName() string {
	return "system_configs"
}

// CleanupConfig represents cleanup task configuration
type CleanupConfig struct {
	Enabled      bool   `json:"enabled"`       // 是否启用清理任务
	Hour         int    `json:"hour"`          // 执行时间：小时 (0-23)
	Minute       int    `json:"minute"`        // 执行时间：分钟 (0-59)
	RetentionDays int   `json:"retention_days"` // 保留天数
	LastExecutionStatus string `json:"last_execution_status,omitempty"` // 上次执行状态: "success", "failed", "never"
	LastExecutionTime   *time.Time `json:"last_execution_time,omitempty"` // 上次执行时间
	LastExecutionResult string `json:"last_execution_result,omitempty"` // 上次执行结果描述（如删除数量或错误信息）
}

