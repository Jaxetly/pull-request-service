package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Jaxetly/pull-request-service/internal/api"
	"github.com/Jaxetly/pull-request-service/internal/config"
	"github.com/Jaxetly/pull-request-service/internal/database"
	"github.com/Jaxetly/pull-request-service/internal/handler"
	"github.com/Jaxetly/pull-request-service/internal/service"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
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
	statsSvc := service.NewStatsService(db.Pool)

	myServer := handler.NewServer(teamSvc, userSvc, prSvc, statsSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", api.Handler(myServer))

	URL := fmt.Sprintf("http://127.0.0.1:%d", cfg.Server.Port)
	openapiURL := fmt.Sprintf("%s/openapi.yml", URL)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(openapiURL),
	))

	r.Get("/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./openapi.yml")
	})

	log.Printf("Start listening to %s \n", URL)

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), r)
}
