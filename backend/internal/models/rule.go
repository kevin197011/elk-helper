// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// QueryCondition represents a flexible query condition
type QueryCondition struct {
	Field    string      `json:"field" yaml:"field"`
	Type     string      `json:"type,omitempty" yaml:"type,omitempty"`
	Value    interface{} `json:"value" yaml:"value"`
	Operator string      `json:"operator,omitempty" yaml:"operator,omitempty"` // 主要字段名
	Op       string      `json:"op,omitempty" yaml:"op,omitempty"`             // 兼容旧字段名
	Logic    string      `json:"logic,omitempty" yaml:"logic,omitempty"`
}

// QueryConditions is a slice of QueryCondition for JSON storage
type QueryConditions []QueryCondition

// Value implements driver.Valuer
func (qc QueryConditions) Value() (driver.Value, error) {
	return json.Marshal(qc)
}

// Scan implements sql.Scanner
func (qc *QueryConditions) Scan(value interface{}) error {
	if value == nil {
		*qc = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	// Handle empty string or empty array
	if len(bytes) == 0 || string(bytes) == "null" {
		*qc = nil
		return nil
	}

	return json.Unmarshal(bytes, qc)
}

// Rule represents an alert rule
type Rule struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string          `gorm:"not null;uniqueIndex" json:"name"`
	IndexPattern string          `gorm:"not null" json:"index_pattern"`
	Queries      QueryConditions `gorm:"type:text" json:"queries"`
	Enabled      bool            `gorm:"default:true" json:"enabled"`
	Interval     int             `gorm:"default:60" json:"interval"` // seconds
	ESConfigID   *uint           `gorm:"index" json:"es_config_id,omitempty"`           // ES 数据源配置 ID
	ESConfig     *ESConfig       `gorm:"foreignKey:ESConfigID" json:"es_config,omitempty"` // ES 数据源配置关联
	LarkWebhook  string          `json:"lark_webhook"`                // 保留用于向后兼容，如果设置了 LarkConfigID 则优先使用配置
	LarkConfigID *uint           `gorm:"index" json:"lark_config_id,omitempty"` // Lark 配置 ID
	LarkConfig   *LarkConfig     `gorm:"foreignKey:LarkConfigID" json:"lark_config,omitempty"` // Lark 配置关联
	Description  string          `json:"description,omitempty"`

	// Statistics
	LastRunTime *time.Time `json:"last_run_time,omitempty"`
	RunCount    int64      `gorm:"default:0" json:"run_count"`
	AlertCount  int64      `gorm:"default:0" json:"alert_count"`
}

// TableName specifies the table name for Rule
func (Rule) TableName() string {
	return "rules"
}
