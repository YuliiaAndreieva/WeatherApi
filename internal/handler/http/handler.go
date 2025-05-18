package http

import (
	"log"
	"net/http"
	"weather-api/internal/core/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	weatherSvc      *service.WeatherService
	subscriptionSvc *service.SubscriptionService
}

func NewHandler(weatherSvc *service.WeatherService, subscriptionSvc *service.SubscriptionService) *Handler {
	return &Handler{
		weatherSvc:      weatherSvc,
		subscriptionSvc: subscriptionSvc,
	}
}

func (h *Handler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	log.Printf("Received weather request for city: %s", city)

	if city == "" {
		log.Printf("Bad request: city parameter is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
		return
	}

	weather, err := h.weatherSvc.GetWeather(city)
	if err != nil {
		log.Printf("Failed to get weather for city %s: %v", city, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Successfully retrieved weather for city: %s", city)
	c.JSON(http.StatusOK, weather)
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Email     string `json:"email"`
		City      string `json:"city"`
		Frequency string `json:"frequency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid subscription request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received subscription request for email: %s, city: %s, frequency: %s", req.Email, req.City, req.Frequency)

	if req.Frequency != "daily" && req.Frequency != "hourly" {
		log.Printf("Invalid frequency in subscription request: %s", req.Frequency)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Frequency must be 'daily' or 'hourly'"})
		return
	}

	token, err := h.subscriptionSvc.Subscribe(c, req.Email, req.City, req.Frequency)
	if err != nil {
		log.Printf("Failed to process subscription for email %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Successfully processed subscription request for email: %s", req.Email)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Confirm(c *gin.Context) {
	token := c.Param("token")
	log.Printf("Received confirmation request for token: %s", token)

	if err := h.subscriptionSvc.Confirm(c, token); err != nil {
		log.Printf("Failed to confirm subscription for token %s: %v", token, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Successfully confirmed subscription for token: %s", token)
	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed"})
}

func (h *Handler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	log.Printf("Received unsubscribe request for token: %s", token)

	if err := h.subscriptionSvc.Unsubscribe(c, token); err != nil {
		log.Printf("Failed to unsubscribe for token %s: %v", token, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Successfully processed unsubscribe request for token: %s", token)
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed"})
}
