package main

import (
	"log"
	"os"

	"github.com/blcvn/backend/services/ba-tool-registry/internal/handler"
	"github.com/blcvn/backend/services/ba-tool-registry/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=ba_agent port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate schemas
	if err := db.AutoMigrate(&models.Tool{}, &models.ToolPermission{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	registry := handler.NewToolRegistry(db)

	log.Println("Tool Registry Service started")
	log.Printf("Registry initialized with %d tools", countTools(registry))

	// Keep service running
	select {}
}

func countTools(r *handler.ToolRegistry) int {
	tools, _ := r.ListTools()
	return len(tools)
}
