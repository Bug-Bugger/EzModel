package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Connect establishes a connection to the database and performs migrations
func Connect(cfg *config.Config) (*gorm.DB, error) {
	// Primary database connection
	primaryDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(primaryDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary database: %w", err)
	}

	// Configure read replica if enabled
	if cfg.DatabaseReplica.Enabled && cfg.DatabaseReplica.Host != "" {
		log.Printf("Configuring read replica: %s:%s", cfg.DatabaseReplica.Host, cfg.DatabaseReplica.Port)

		replicaDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DatabaseReplica.User, cfg.DatabaseReplica.Password, cfg.DatabaseReplica.Host,
			cfg.DatabaseReplica.Port, cfg.DatabaseReplica.DBName, cfg.DatabaseReplica.SSLMode)

		err = db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(replicaDSN)},
			Policy:   dbresolver.RandomPolicy{}, // Load balance across replicas
		}).
			SetConnMaxIdleTime(time.Hour).
			SetConnMaxLifetime(24 * time.Hour).
			SetMaxIdleConns(10).
			SetMaxOpenConns(100))

		if err != nil {
			log.Printf("WARNING: Failed to configure read replica: %v. Continuing with primary only.", err)
		} else {
			log.Println("Read replica configured successfully")
			// Start health check for replica
			go startReplicaHealthCheck(db)
		}
	} else {
		log.Println("Read replica not enabled, using primary database only")
	}

	// Ensure pgcrypto extension exists for UUID generation defaults
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, fmt.Errorf("failed to enable pgcrypto extension: %w", err)
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

// startReplicaHealthCheck monitors replica health and logs issues
func startReplicaHealthCheck(db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Use replica for health check
		var result int
		err := db.Clauses(dbresolver.Read).Raw("SELECT 1").Scan(&result).Error
		if err != nil {
			log.Printf("Replica health check failed: %v. Queries will use primary database.", err)
		}
	}
}

// isTableExistsError checks if the error is related to tables already existing
func isTableExistsError(errorMsg string) bool {
	errorMsg = strings.ToLower(errorMsg)
	return strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "relation") && strings.Contains(errorMsg, "already exists")
}
