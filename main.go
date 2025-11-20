package main

import (
	"context"
	"log"
	"time"

	"github.com/Jaxetly/pull-request-service/internal/config"
	"github.com/Jaxetly/pull-request-service/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Hello Avito!")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration Error:\n%v", err)
	}

	log.Println("Configuration loaded successfully")
	log.Printf("Database: %v", cfg.Database.DatabaseURL())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Подключаемся к БД
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database!")
}
