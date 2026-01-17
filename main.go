package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type RadioInfo struct {
	BW float64 `json:"bw"`
	SF int     `json:"sf"`
	CR int     `json:"cr"`
	TX int     `json:"tx"`
}

type Metadata struct {
	Name   string    `json:"name"`
	Pubkey string    `json:"pubkey"`
	Radio  RadioInfo `json:"radio"`
}

type DeviceData struct {
	DeviceID   string  `json:"deviceId"`
	DeviceName string  `json:"deviceName"`
	RSSI       int     `json:"rssi"`
	SNR        float64 `json:"snr"`
	Timestamp  string  `json:"timestamp"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	ScanSource string  `json:"scanSource"`
}

type ReportRequest struct {
	Metadata Metadata     `json:"metadata"`
	Data     []DeviceData `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func validateRadioInfo(radio RadioInfo) error {
	if radio.BW <= 0 {
		return fmt.Errorf("radio.bw must be greater than 0")
	}
	if radio.SF <= 0 {
		return fmt.Errorf("radio.sf must be greater than 0")
	}
	if radio.CR <= 0 {
		return fmt.Errorf("radio.cr must be greater than 0")
	}
	if radio.TX <= 0 {
		return fmt.Errorf("radio.tx must be greater than 0")
	}
	return nil
}

func validateMetadata(metadata Metadata) error {
	if strings.TrimSpace(metadata.Name) == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if strings.TrimSpace(metadata.Pubkey) == "" {
		return fmt.Errorf("metadata.pubkey is required")
	}
	if err := validateRadioInfo(metadata.Radio); err != nil {
		return err
	}
	return nil
}

func validateDeviceData(data DeviceData, index int) error {
	if strings.TrimSpace(data.DeviceID) == "" {
		return fmt.Errorf("data[%d].deviceId is required", index)
	}
	if strings.TrimSpace(data.DeviceName) == "" {
		return fmt.Errorf("data[%d].deviceName is required", index)
	}
	if strings.TrimSpace(data.Timestamp) == "" {
		return fmt.Errorf("data[%d].timestamp is required", index)
	}
	// Validate timestamp format - try multiple common formats
	validFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	}
	validTimestamp := false
	for _, format := range validFormats {
		if _, err := time.Parse(format, data.Timestamp); err == nil {
			validTimestamp = true
			break
		}
	}
	if !validTimestamp {
		return fmt.Errorf("data[%d].timestamp must be in a valid ISO8601/RFC3339 format", index)
	}
	// Validate latitude range
	if data.Latitude < -90 || data.Latitude > 90 {
		return fmt.Errorf("data[%d].latitude must be between -90 and 90", index)
	}
	// Validate longitude range
	if data.Longitude < -180 || data.Longitude > 180 {
		return fmt.Errorf("data[%d].longitude must be between -180 and 180", index)
	}
	if strings.TrimSpace(data.ScanSource) == "" {
		return fmt.Errorf("data[%d].scanSource is required", index)
	}
	return nil
}

func validateReport(report ReportRequest) error {
	if err := validateMetadata(report.Metadata); err != nil {
		return err
	}
	if len(report.Data) == 0 {
		return fmt.Errorf("data array cannot be empty")
	}
	for i, device := range report.Data {
		if err := validateDeviceData(device, i); err != nil {
			return err
		}
	}
	return nil
}

func handleReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var report ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}
	defer r.Body.Close()

	// Validate the report
	if err := validateReport(report); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	log.Printf("Received valid report from: %s\n", report.Metadata.Name)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
	http.HandleFunc("/report", handleReport)

	port := "8080"
	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
