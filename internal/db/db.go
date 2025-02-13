// Package db provides database connection and management functionality
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DB represents a database instance
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
