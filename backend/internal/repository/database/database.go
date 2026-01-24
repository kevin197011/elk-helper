// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/kk/elk-helper/backend/internal/config"
	"github.com/kk/elk-helper/backend/internal/migrations"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// WithTimeout returns a gorm DB session with a best-effort timeout context.
// This is a guardrail against slow SQL operations blocking the API/worker forever.
func WithTimeout(parent context.Context) (*gorm.DB, context.CancelFunc) {
	timeoutSeconds := 5
	if config.AppConfig != nil && config.AppConfig.Database.QueryTimeoutSeconds > 0 {
		timeoutSeconds = config.AppConfig.Database.QueryTimeoutSeconds
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(timeoutSeconds)*time.Second)
	return DB.WithContext(ctx), cancel
}

// Init initializes the database connection and runs migrations
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

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(pgdriver.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Configure connection pool for high concurrency (1000+ rules)
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings for 1000+ concurrent rules
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Run database migrations
	if err := runMigrations(sqlDB, dbConfig.DBName); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations executes database migrations using golang-migrate
func runMigrations(db *sql.DB, dbName string) error {
	slog.Info("Running database migrations...")

	// Create source driver from embedded files
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration db driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, dbName, dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Database schema is up to date, no migrations needed")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		slog.Warn("Failed to get migration version", "error", err)
	} else {
		slog.Info("Database migrations completed", "version", version, "dirty", dirty)
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
