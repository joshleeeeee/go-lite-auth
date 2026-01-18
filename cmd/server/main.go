package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joshleeeeee/go-lite-auth/internal/config"
	"github.com/joshleeeeee/go-lite-auth/internal/database"
	"github.com/joshleeeeee/go-lite-auth/internal/router"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded from %s", *configPath)

	// Initialize MySQL
	if err := database.InitMySQL(&cfg.MySQL); err != nil {
		log.Fatalf("Failed to initialize MySQL: %v", err)
	}
	defer database.Close()

	// Auto migrate database tables
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer database.CloseRedis()

	// Setup router
	r := router.Setup(cfg.Server.Mode)

	// Start server in a goroutine
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	go func() {
		log.Printf("🚀 Lite-Auth server starting on http://localhost%s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
