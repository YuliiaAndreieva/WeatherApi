package service

import (
	"weather-api/internal/core/domain"
	"weather-api/internal/core/port"
)

type WeatherService struct {
	weatherSvc port.WeatherService
}

func NewWeatherService(weatherSvc port.WeatherService) *WeatherService {
	return &WeatherService{weatherSvc: weatherSvc}
}

func (s *WeatherService) GetWeather(city string) (domain.Weather, error) {
	return s.weatherSvc.GetWeather(city)
}
