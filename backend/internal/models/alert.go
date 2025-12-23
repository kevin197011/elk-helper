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

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusSent   AlertStatus = "sent"
	AlertStatusFailed AlertStatus = "failed"
)

// LogData stores the matched log data
type LogData []map[string]interface{}

// Value implements driver.Valuer
func (ld LogData) Value() (driver.Value, error) {
	return json.Marshal(ld)
}

// Scan implements sql.Scanner
func (ld *LogData) Scan(value interface{}) error {
	if value == nil {
		*ld = nil
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

	// Handle empty string or null string
	if len(bytes) == 0 || string(bytes) == "null" {
		*ld = nil
		return nil
	}

	return json.Unmarshal(bytes, ld)
}

// Alert represents an alert record
type Alert struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	RuleID    uint        `gorm:"not null;index" json:"rule_id"`
	Rule      Rule        `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"rule,omitempty"`
	IndexName string      `gorm:"not null" json:"index_name"`
	LogCount  int         `json:"log_count"`
	Logs      LogData     `gorm:"type:text" json:"logs"`
	TimeRange string      `json:"time_range"` // e.g., "2025-11-28 10:00:00 ~ 10:01:00"
	Status    AlertStatus `gorm:"default:'sent'" json:"status"`
	ErrorMsg  string      `json:"error_msg,omitempty"`
}

// TableName specifies the table name for Alert
func (Alert) TableName() string {
	return "alerts"
}
