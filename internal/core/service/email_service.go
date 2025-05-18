package service

import (
	"context"
	"fmt"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/port"
)

type EmailService struct {
	repo       port.SubscriptionRepository
	weatherSvc port.WeatherService
	emailSvc   port.EmailService
}

func NewEmailService(repo port.SubscriptionRepository, weatherSvc port.WeatherService, emailSvc port.EmailService) *EmailService {
	return &EmailService{
		repo:       repo,
		weatherSvc: weatherSvc,
		emailSvc:   emailSvc,
	}
}

func (s *EmailService) SendUpdates(ctx context.Context, frequency domain.Frequency) {
	subs, err := s.repo.GetSubscriptionsByFrequency(ctx, string(frequency))
	if err != nil {
		log.Printf("Failed to get %s subscriptions: %v", frequency, err)
		return
	}
	s.sendUpdates(subs)
}

func (s *EmailService) sendUpdates(subs []domain.Subscription) {
	for _, sub := range subs {
		if !sub.IsConfirmed {
			continue
		}
		weather, err := s.weatherSvc.GetWeather(sub.City)
		if err != nil {
			log.Printf("Failed to get weather for %s: %v", sub.City, err)
			continue
		}
		body := fmt.Sprintf("Weather in %s: Temp %.2f°C, Humidity %d%%, %s\nUnsubscribe: http://localhost:8080/api/unsubscribe/%s", sub.City, weather.Temperature, weather.Humidity, weather.Description, sub.Token)
		if err := s.emailSvc.SendEmail(sub.Email, "Weather Update", body); err != nil {
			log.Printf("Failed to send email to %s: %v", sub.Email, err)
		}
	}
}
