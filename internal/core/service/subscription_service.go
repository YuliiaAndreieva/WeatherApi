package service

import (
	"context"
	"log"
	"weather-api/internal/core/domain"
	"weather-api/internal/core/port"
)

type SubscriptionService struct {
	repo       port.SubscriptionRepository
	weatherSvc port.WeatherService
	emailSvc   port.EmailService
	tokenSvc   port.TokenService
}

func NewSubscriptionService(repo port.SubscriptionRepository, weatherSvc port.WeatherService, emailSvc port.EmailService, tokenSvc port.TokenService) *SubscriptionService {
	return &SubscriptionService{
		repo:       repo,
		weatherSvc: weatherSvc,
		emailSvc:   emailSvc,
		tokenSvc:   tokenSvc,
	}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, email, city, frequency string) (string, error) {
	log.Printf("Attempting to create subscription for email: %s, city: %s, frequency: %s", email, city, frequency)

	token, err := s.tokenSvc.GenerateToken()
	if err != nil {
		log.Printf("Failed to generate token for email %s: %v", email, err)
		return "", err
	}

	sub := domain.Subscription{
		Email:       email,
		City:        city,
		Frequency:   frequency,
		Token:       token,
		IsConfirmed: false,
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		log.Printf("Failed to create subscription in repository for email %s: %v", email, err)
		return "", err
	}

	confirmURL := "http://localhost:8080/api/confirm/" + token
	body := "Confirm your subscription: " + confirmURL
	if err := s.emailSvc.SendEmail(email, "Confirm Subscription", body); err != nil {
		log.Printf("Failed to send confirmation email to %s: %v", email, err)
		return "", err
	}

	log.Printf("Successfully created subscription for email: %s with token: %s", email, token)
	return token, nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	log.Printf("Attempting to confirm subscription with token: %s", token)

	sub, err := s.repo.GetSubscriptionByToken(ctx, token)
	if err != nil {
		log.Printf("Failed to get subscription for token %s: %v", token, err)
		return err
	}
	sub.IsConfirmed = true
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		log.Printf("Failed to update subscription confirmation for token %s: %v", token, err)
		return err
	}

	log.Printf("Successfully confirmed subscription for email: %s", sub.Email)
	return nil
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	log.Printf("Attempting to unsubscribe with token: %s", token)

	if err := s.repo.DeleteSubscription(ctx, token); err != nil {
		log.Printf("Failed to delete subscription for token %s: %v", token, err)
		return err
	}

	log.Printf("Successfully unsubscribed for token: %s", token)
	return nil
}
