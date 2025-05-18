package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"time"

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
	log.Printf("pass %s", cfg.SMTPPass)

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

	handler := httphandler.NewHandler(weatherService, subscriptionService)

	r := gin.Default()
	r.Static("/web", "./web")
	r.GET("/api/weather", handler.GetWeather)
	r.POST("/api/subscribe", handler.Subscribe)
	r.GET("/api/confirm/:token", handler.Confirm)
	r.GET("/api/unsubscribe/:token", handler.Unsubscribe)

	cron := cron.New()
	cron.AddFunc("* * * * *", func() { emailService.SendHourlyUpdates(context.Background()) })
	cron.AddFunc("0 0 8 * * *", func() { emailService.SendDailyUpdates(context.Background()) })
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
