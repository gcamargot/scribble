package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database represents the database connection and provides methods for queries
type Database struct {
	conn *gorm.DB
}

// NewDatabase initializes a new database connection from environment variables
// Connection string format: postgres://user:password@host:port/dbname
func NewDatabase() (*Database, error) {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Fallback to constructing from individual env vars
		databaseURL = constructDatabaseURL()
	}

	// Open connection to PostgreSQL
	conn, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database for connection pool configuration
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	// Set max open connections to prevent resource exhaustion
	sqlDB.SetMaxOpenConns(25)
	// Set max idle connections to clean up resources
	sqlDB.SetMaxIdleConns(5)

	return &Database{
		conn: conn,
	}, nil
}

// constructDatabaseURL builds a PostgreSQL connection string from environment variables
// Supports DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME
func constructDatabaseURL() string {
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		dbName = "scribble_dev"
	}

	// Build DSN: postgres://user:password@host:port/dbname
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	)
}

// getLogLevel returns the appropriate GORM logger level based on LOG_LEVEL env var
func getLogLevel() logger.LogLevel {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		return logger.Info
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Warn
	}
}

// GetConnection returns the underlying GORM database connection
// Use this to perform queries: db.GetConnection().Find(&users)
func (db *Database) GetConnection() *gorm.DB {
	return db.conn
}

// AutoMigrate runs all pending migrations using GORM's auto migration
// This is useful for development but should be controlled in production
func (db *Database) AutoMigrate(models ...interface{}) error {
	return db.conn.AutoMigrate(models...)
}

// Close closes the database connection
func (db *Database) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}

// Ping checks if the database connection is alive
func (db *Database) Ping() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Ping()
}
