package db

import (
	"fmt"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a connection to the database and performs migrations
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migrate the schema (safe migration that handles existing tables)
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Table{},
		&models.Field{},
		&models.Relationship{},
		&models.CollaborationSession{},
	)
	if err != nil {
		// Check if the error is about tables already existing
		if !isTableExistsError(err.Error()) {
			return nil, err
		}
		// If it's just table exists errors, we can continue safely
	}

	return db, nil
}

// isTableExistsError checks if the error is related to tables already existing
func isTableExistsError(errorMsg string) bool {
	errorMsg = strings.ToLower(errorMsg)
	return strings.Contains(errorMsg, "already exists") ||
		   strings.Contains(errorMsg, "relation") && strings.Contains(errorMsg, "already exists")
}
