package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/joshleeeeee/go-lite-auth/internal/config"
	"github.com/joshleeeeee/go-lite-auth/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database based on the configuration driver
func InitDB(cfg *config.Config) error {
	var err error
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch cfg.Database.Driver {
	case "mysql":
		DB, err = gorm.Open(mysql.Open(cfg.MySQL.DSN()), gormConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL: %w", err)
		}
		// Configure connection pool for MySQL
		sqlDB, _ := DB.DB()
		sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
		log.Println("Database: MySQL connected successfully")

	case "postgres":
		DB, err = gorm.Open(postgres.Open(cfg.Postgres.DSN()), gormConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		// Configure connection pool for PostgreSQL
		sqlDB, _ := DB.DB()
		sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)
		log.Println("Database: PostgreSQL connected successfully")

	case "sqlite", "":
		// Ensure data directory exists
		dbDir := filepath.Dir(cfg.SQLite.Path)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("failed to create sqlite directory: %w", err)
		}

		DB, err = gorm.Open(sqlite.Open(cfg.SQLite.Path), gormConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite: %w", err)
		}
		log.Printf("Database: SQLite connected successfully at %s", cfg.SQLite.Path)

	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	return nil
}

// AutoMigrate creates/updates database tables
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&model.User{},
		&model.Client{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	log.Println("Database migration completed")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
