package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/config"
	"github.com/Jaxetly/pull-request-service/internal/database"
	"github.com/Jaxetly/pull-request-service/internal/handler"
	"github.com/Jaxetly/pull-request-service/internal/service"
	"github.com/go-chi/chi/v5"
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

	teamSvc := service.NewTeamService(db.Pool)
	userSvc := service.NewUserService(db.Pool)
	prSvc := service.NewPRService(db.Pool)

	myServer := handler.NewServer(teamSvc, userSvc, prSvc)

	r := chi.NewRouter()
	r.Mount("/", api.Handler(myServer))

	log.Printf("Start listening to http://127.0.0.1:%v \n", cfg.Server.Port)

	http.ListenAndServe(":8080", r)
}
