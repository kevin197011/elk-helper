// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package database

import (
	"fmt"
	"time"

	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection
func Init(dbConfig config.DatabaseConfig) error {
	var err error

	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.Port,
		dbConfig.SSLMode,
	)

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Configure connection pool for high concurrency (1000+ rules)
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings for 1000+ concurrent rules
	// MaxOpenConns: maximum number of open connections to the database
	// MaxIdleConns: maximum number of connections in the idle connection pool
	// ConnMaxLifetime: maximum amount of time a connection may be reused
	sqlDB.SetMaxOpenConns(200)                // Allow up to 200 open connections
	sqlDB.SetMaxIdleConns(50)                 // Keep 50 idle connections ready
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Recycle connections after 5 minutes

	// Auto migrate
	if err := DB.AutoMigrate(&models.Rule{}, &models.Alert{}, &models.ESConfig{}, &models.LarkConfig{}, &models.User{}, &models.SystemConfig{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
