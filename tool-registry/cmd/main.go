package main

import (
	"encoding/json"
	"log"
	"net/http"
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

	http.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		tools, _ := registry.ListTools()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tools)
	})

	http.HandleFunc("/tools/register", func(w http.ResponseWriter, r *http.Request) {
		var tool models.Tool
		if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := registry.RegisterTool(&tool); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	log.Println("Listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func countTools(r *handler.ToolRegistry) int {
	tools, _ := r.ListTools()
	return len(tools)
}
