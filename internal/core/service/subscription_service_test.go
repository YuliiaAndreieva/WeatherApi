package service

import (
	"context"
	"errors"
	"testing"
	"weather-api/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"weather-api/internal/core/domain"
	"weather-api/internal/util"
)

func TestSubscriptionService_Subscribe(t *testing.T) {
	ctx := context.Background()
	email := "user1@example.com"
	city := "Kyiv"
	frequency := domain.FrequencyDaily

	tests := []struct {
		name          string
		email         string
		city          string
		frequency     domain.Frequency
		setupMocks    func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		verifyMocks   func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		expectedToken string
		expectedError error
	}{
		{
			name:      "success",
			email:     email,
			city:      city,
			frequency: frequency,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				token := "token123"
				tokenSvc.On("GenerateToken").Return(token, nil)
				sub := domain.Subscription{
					Email:       email,
					City:        city,
					Frequency:   frequency,
					Token:       token,
					IsConfirmed: false,
				}
				repo.On("CreateSubscription", ctx, sub).Return(nil)
				subject, body := util.BuildConfirmationEmail(city, token)
				emailSvc.On("SendEmail", email, subject, body).Return(nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				tokenSvc.AssertExpectations(t)
				repo.AssertExpectations(t)
				emailSvc.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
			},
			expectedToken: "token123",
			expectedError: nil,
		},
		{
			name:      "token generation error",
			email:     email,
			city:      city,
			frequency: frequency,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				tokenSvc.On("GenerateToken").Return("", errors.New("token error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				tokenSvc.AssertExpectations(t)
				repo.AssertNotCalled(t, "CreateSubscription", mock.Anything, mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
			},
			expectedToken: "",
			expectedError: errors.New("token error"),
		},
		{
			name:      "repository error",
			email:     email,
			city:      city,
			frequency: frequency,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				token := "token123"
				tokenSvc.On("GenerateToken").Return(token, nil)
				sub := domain.Subscription{
					Email:       email,
					City:        city,
					Frequency:   frequency,
					Token:       token,
					IsConfirmed: false,
				}
				repo.On("CreateSubscription", ctx, sub).Return(errors.New("db error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				tokenSvc.AssertExpectations(t)
				repo.AssertExpectations(t)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
			},
			expectedToken: "",
			expectedError: errors.New("db error"),
		},
		{
			name:      "email service error",
			email:     email,
			city:      city,
			frequency: frequency,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				token := "token123"
				tokenSvc.On("GenerateToken").Return(token, nil)
				sub := domain.Subscription{
					Email:       email,
					City:        city,
					Frequency:   frequency,
					Token:       token,
					IsConfirmed: false,
				}
				repo.On("CreateSubscription", ctx, sub).Return(nil)
				subject, body := util.BuildConfirmationEmail(city, token)
				emailSvc.On("SendEmail", email, subject, body).Return(errors.New("SMTP error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				tokenSvc.AssertExpectations(t)
				repo.AssertExpectations(t)
				emailSvc.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
			},
			expectedToken: "",
			expectedError: errors.New("SMTP error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			weatherSvc := &mocks.MockWeatherService{}
			emailSvc := &mocks.MockEmailService{}
			tokenSvc := &mocks.MockTokenService{}
			service := NewSubscriptionService(repo, weatherSvc, emailSvc, tokenSvc)

			tt.setupMocks(repo, weatherSvc, emailSvc, tokenSvc)

			token, err := service.Subscribe(ctx, tt.email, tt.city, tt.frequency)

			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedError, err)
			tt.verifyMocks(t, repo, weatherSvc, emailSvc, tokenSvc)
		})
	}
}

func TestSubscriptionService_Confirm(t *testing.T) {
	ctx := context.Background()
	token := "token123"

	tests := []struct {
		name          string
		token         string
		setupMocks    func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		verifyMocks   func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		expectedError error
	}{
		{
			name:  "success",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				sub := domain.Subscription{
					Email:       "user1@example.com",
					City:        "Kyiv",
					Frequency:   domain.FrequencyDaily,
					Token:       token,
					IsConfirmed: false,
				}
				repo.On("GetSubscriptionByToken", ctx, token).Return(sub, nil)
				updatedSub := sub
				updatedSub.IsConfirmed = true
				repo.On("UpdateSubscription", ctx, updatedSub).Return(nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				tokenSvc.AssertNotCalled(t, "GenerateToken")
			},
			expectedError: nil,
		},
		{
			name:  "token not found",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.On("GetSubscriptionByToken", ctx, token).Return(domain.Subscription{}, errors.New("not found"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.AssertExpectations(t)
				repo.AssertNotCalled(t, "UpdateSubscription", mock.Anything, mock.Anything)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				tokenSvc.AssertNotCalled(t, "GenerateToken")
			},
			expectedError: errors.New("not found"),
		},
		{
			name:  "update subscription error",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				sub := domain.Subscription{
					Email:       "user1@example.com",
					City:        "Kyiv",
					Frequency:   domain.FrequencyDaily,
					Token:       token,
					IsConfirmed: false,
				}
				repo.On("GetSubscriptionByToken", ctx, token).Return(sub, nil)
				updatedSub := sub
				updatedSub.IsConfirmed = true
				repo.On("UpdateSubscription", ctx, updatedSub).Return(errors.New("db error"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				tokenSvc.AssertNotCalled(t, "GenerateToken")
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			weatherSvc := &mocks.MockWeatherService{}
			emailSvc := &mocks.MockEmailService{}
			tokenSvc := &mocks.MockTokenService{}
			service := NewSubscriptionService(repo, weatherSvc, emailSvc, tokenSvc)

			tt.setupMocks(repo, weatherSvc, emailSvc, tokenSvc)

			err := service.Confirm(ctx, tt.token)

			assert.Equal(t, tt.expectedError, err)
			tt.verifyMocks(t, repo, weatherSvc, emailSvc, tokenSvc)
		})
	}
}

func TestSubscriptionService_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	token := "token123"

	tests := []struct {
		name          string
		token         string
		setupMocks    func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		verifyMocks   func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService)
		expectedError error
	}{
		{
			name:  "success",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.On("DeleteSubscription", ctx, token).Return(nil)
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				tokenSvc.AssertNotCalled(t, "GenerateToken")
			},
			expectedError: nil,
		},
		{
			name:  "deletion error",
			token: token,
			setupMocks: func(repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.On("DeleteSubscription", ctx, token).Return(errors.New("not found"))
			},
			verifyMocks: func(t *testing.T, repo *mocks.MockSubscriptionRepository, weatherSvc *mocks.MockWeatherService, emailSvc *mocks.MockEmailService, tokenSvc *mocks.MockTokenService) {
				repo.AssertExpectations(t)
				weatherSvc.AssertNotCalled(t, "GetWeather", mock.Anything)
				emailSvc.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)
				tokenSvc.AssertNotCalled(t, "GenerateToken")
			},
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockSubscriptionRepository{}
			weatherSvc := &mocks.MockWeatherService{}
			emailSvc := &mocks.MockEmailService{}
			tokenSvc := &mocks.MockTokenService{}
			service := NewSubscriptionService(repo, weatherSvc, emailSvc, tokenSvc)

			tt.setupMocks(repo, weatherSvc, emailSvc, tokenSvc)

			err := service.Unsubscribe(ctx, tt.token)

			assert.Equal(t, tt.expectedError, err)
			tt.verifyMocks(t, repo, weatherSvc, emailSvc, tokenSvc)
		})
	}
}
