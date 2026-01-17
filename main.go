package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RadioInfo struct {
	BW float64 `json:"bw" validate:"required,gt=0"`
	SF int     `json:"sf" validate:"required,gt=0"`
	CR int     `json:"cr" validate:"required,gt=0"`
	TX int     `json:"tx" validate:"required,gt=0"`
}

type Metadata struct {
	Name   string    `json:"name" validate:"required"`
	Pubkey string    `json:"pubkey" validate:"required"`
	Radio  RadioInfo `json:"radio" validate:"required"`
}

type DeviceData struct {
	DeviceID   string  `json:"deviceId" validate:"required"`
	DeviceName string  `json:"deviceName" validate:"required"`
	RSSI       int     `json:"rssi"`
	SNR        float64 `json:"snr"`
	Timestamp  string  `json:"timestamp" validate:"required,timestamp"`
	Latitude   float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude  float64 `json:"longitude" validate:"required,min=-180,max=180"`
	ScanSource string  `json:"scanSource" validate:"required"`
}

type ReportRequest struct {
	Metadata Metadata     `json:"metadata" validate:"required"`
	Data     []DeviceData `json:"data" validate:"required,min=1,dive"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("timestamp", validateTimestamp)
}

func validateTimestamp(fl validator.FieldLevel) bool {
	timestamp := fl.Field().String()
	validFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	}
	for _, format := range validFormats {
		if _, err := time.Parse(format, timestamp); err == nil {
			return true
		}
	}
	return false
}

func handleReport(c *gin.Context) {
	var report ReportRequest

	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	if err := validate.Struct(&report); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("Received valid report from: %s\n", report.Metadata.Name)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func main() {
	router := gin.Default()

	router.HandleMethodNotAllowed = true

	router.POST("/report", handleReport)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Route not found"})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
	})

	port := "8080"
	log.Printf("Server starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
