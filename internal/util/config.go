package util

import (
	"os"
	"strconv"
)

type Config struct {
	DBConnStr     string
	WeatherAPIKey string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	Port          int
	BaseUrl       string
}

func LoadConfig() (*Config, error) {
	return &Config{
		DBConnStr:     os.Getenv("DB_CONN_STR"),
		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
		BaseUrl:       os.Getenv("BASE_URL"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      getEnvInt("SMTP_PORT", 587),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPass:      os.Getenv("SMTP_PASS"),
		Port:          getEnvInt("PORT", 8080),
	}, nil
}

func getEnvInt(key string, defaultValue int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func GetBaseURL() string {
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:8080"
	}
	return baseUrl
}
