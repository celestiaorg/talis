// Package db provides database connectivity and operations
package db

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/celestiaorg/talis/internal/db/models"
)

// Database configuration constants
const (
	// DefaultHost is the default database host
	DefaultHost = "localhost"
	// DefaultPort is the default database port
	DefaultPort = 5432
	// DefaultUser is the default database user
	DefaultUser = "postgres"
	// DefaultPassword is the default database password
	DefaultPassword = "postgres"
	// DefaultDBName is the default database name
	DefaultDBName     = "postgres"
	DefaultSSLEnabled = false
)

// Options represents database connection configuration options
type Options struct {
	Host       string
	User       string
	Password   string
	DBName     string
	Port       int
	SSLEnabled *bool
	LogLevel   logger.LogLevel
}

// New creates a new database connection with the given options
func New(opts Options) (*gorm.DB, error) {
	opts = setDefaults(opts)
	sslMode := "disable"
	if opts.SSLEnabled != nil && *opts.SSLEnabled {
		sslMode = "enable"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		opts.Host, opts.User, opts.Password, opts.DBName, opts.Port, sslMode)

	// Configure custom logger to ignore record not found errors
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:                  opts.LogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Configure GORM
	config := &gorm.Config{
		Logger: newLogger,
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	// --- Add Admin Entities Setup ---
	if err := setupAdminEntities(db); err != nil {
		// Log the error but don't necessarily block startup unless critical
		log.Printf("Warning: Failed to setup admin entities: %v", err)
		// Depending on requirements, you might want to return nil, err here
	}
	// --- End Admin Entities Setup ---

	return db, nil
}

// IsDuplicateKeyError checks if the given error is a PostgreSQL duplicate key error
func IsDuplicateKeyError(err error) bool {
	return errors.Is(postgres.Dialector{}.Translate(err), gorm.ErrDuplicatedKey)
}

func setDefaults(opts Options) Options {
	if opts.Host == "" {
		opts.Host = DefaultHost
	}
	if opts.User == "" {
		opts.User = DefaultUser
	}
	if opts.Password == "" {
		opts.Password = DefaultPassword
	}
	if opts.DBName == "" {
		opts.DBName = DefaultDBName
	}
	if opts.Port == 0 {
		opts.Port = DefaultPort
	}
	if opts.SSLEnabled == nil {
		sslMode := DefaultSSLEnabled
		opts.SSLEnabled = &sslMode
	}
	if opts.LogLevel == 0 {
		opts.LogLevel = logger.Warn
	}
	return opts
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Instance{},
		&models.Project{},
		&models.Task{},
		&models.User{},
	)
}

// setupAdminEntities ensures the admin user and project exist in the database.
func setupAdminEntities(db *gorm.DB) error {
	log.Println("Ensuring admin user and project exist...")

	adminUser := models.User{
		Model:    gorm.Model{ID: models.AdminID},
		Username: "talis-admin", // Or just "admin"
	}
	// Use Where condition based on ID for FirstOrCreate check
	result := db.Where("id = ?", models.AdminID).FirstOrCreate(&models.User{}, adminUser)
	if result.Error != nil {
		return fmt.Errorf("failed to ensure admin user exists (ID: %d): %w", models.AdminID, result.Error)
	}

	// --- Ensure Admin Project Exists ---
	adminProject := models.Project{
		Model:   gorm.Model{ID: models.AdminProjectID},
		Name:    "talis-admin-project", // Or just "admin-project"
		OwnerID: models.AdminID,        // Link it to the admin user
		// Add other non-nullable fields with default values if necessary
		// Example: Description: "Default admin project", if Description is non-nullable
	}
	result = db.Where("id = ?", models.AdminProjectID).FirstOrCreate(&models.Project{}, adminProject)
	if result.Error != nil {
		return fmt.Errorf("failed to ensure admin project exists (ID: %d): %w", models.AdminProjectID, result.Error)
	}

	log.Println("Admin entities check complete.")
	return nil
}
