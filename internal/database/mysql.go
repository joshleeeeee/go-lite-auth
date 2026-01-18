package database

import (
	"fmt"
	"log"

	"github.com/joshleeeeee/LiteSSO/internal/config"
	"github.com/joshleeeeee/LiteSSO/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMySQL initializes the MySQL connection
func InitMySQL(cfg *config.MySQLConfig) error {
	var err error

	// Configure GORM logger based on mode
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	log.Println("MySQL connected successfully")
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
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
