package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/mmcloughlin/geohash"
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
var db driver.Conn

func init() {
	validate = validator.New()
	validate.RegisterValidation("timestamp", validateTimestamp)

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	var err error
	db, err = initClickHouse()
	if err != nil {
		log.Fatalf("Failed to initialize ClickHouse: %v", err)
	}
}

func initClickHouse() (driver.Conn, error) {
	host := os.Getenv("CLICKHOUSE_HOST")
	port := os.Getenv("CLICKHOUSE_PORT")
	database := os.Getenv("CLICKHOUSE_DATABASE")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      time.Second * 30,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to ClickHouse")
	return conn, nil
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

	if err := insertReportData(report); err != nil {
		log.Printf("Error inserting report data: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to store report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func insertReportData(report ReportRequest) error {
	ctx := context.Background()

	batch, err := db.PrepareBatch(ctx, "INSERT INTO repeater_reports")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, device := range report.Data {
		timestamp, err := parseTimestamp(device.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}

		geoHash := geohash.Encode(device.Latitude, device.Longitude)

		err = batch.Append(
			timestamp,
			report.Metadata.Name,
			report.Metadata.Pubkey,
			report.Metadata.Radio.BW,
			report.Metadata.Radio.SF,
			report.Metadata.Radio.CR,
			report.Metadata.Radio.TX,
			device.DeviceID,
			device.DeviceName,
			device.RSSI,
			device.SNR,
			device.Latitude,
			device.Longitude,
			geoHash,
			"",
			"",
			"",
			device.ScanSource,
			time.Now(),
		)

		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func parseTimestamp(ts string) (time.Time, error) {
	validFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	}
	for _, format := range validFormats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
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
