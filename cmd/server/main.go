package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"time"
	"weather-api/internal/core/domain"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"weather-api/internal/adapter/email"
	"weather-api/internal/adapter/repository/postgres"
	"weather-api/internal/adapter/weather"
	"weather-api/internal/core/service"
	httphandler "weather-api/internal/handler/http"
	"weather-api/internal/util"
)

func main() {
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	m, err := migrate.New("file://migrations", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize migration: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	emailAdapter := email.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	weatherAdapter := weather.NewWeatherService(cfg.WeatherAPIKey)
	repo := postgres.NewSubscriptionRepo(db)

	weatherService := service.NewWeatherService(weatherAdapter)
	tokenService := service.NewTokenService()
	subscriptionService := service.NewSubscriptionService(repo, weatherService, emailAdapter, tokenService)
	emailService := service.NewEmailService(repo, weatherAdapter, emailAdapter)

	weatherHandler := httphandler.NewWeatherHandler(weatherService)
	subscriptionHandler := httphandler.NewSubscriptionHandler(subscriptionService)

	r := gin.Default()
	r.Static("/web", "./web")
	r.GET("/api/weather", weatherHandler.GetWeather)
	r.POST("/api/subscribe", subscriptionHandler.Subscribe)
	r.GET("/api/confirm/:token", subscriptionHandler.Confirm)
	r.GET("/api/unsubscribe/:token", subscriptionHandler.Unsubscribe)

	cron := cron.New()
	cron.AddFunc("* * * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyHourly) })
	cron.AddFunc("0 0 8 * * *", func() { emailService.SendUpdates(context.Background(), domain.FrequencyDaily) })
	cron.Start()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server running on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
