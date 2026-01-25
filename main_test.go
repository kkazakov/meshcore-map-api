package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		valid     bool
	}{
		{"RFC3339", "2026-01-16T21:41:52Z", true},
		{"RFC3339Nano", "2026-01-16T21:41:52.615226Z", true},
		{"Custom format with microseconds", "2026-01-16T21:41:52.615226", true},
		{"Custom format without subseconds", "2026-01-16T21:41:52", true},
		{"Invalid format", "2026-01-16 21:41:52", false},
		{"Invalid timestamp", "invalid", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := DeviceData{
				DeviceID:   "test-device",
				DeviceName: "Test Device",
				Timestamp:  tt.timestamp,
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			}

			err := validate.Struct(&data)
			if tt.valid && err != nil {
				t.Errorf("Expected valid timestamp, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid timestamp, got no error")
			}
		})
	}
}

func TestValidateRadioInfo(t *testing.T) {
	tests := []struct {
		name  string
		radio RadioInfo
		valid bool
	}{
		{
			name:  "Valid radio info",
			radio: RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20},
			valid: true,
		},
		{
			name:  "Zero BW",
			radio: RadioInfo{Freq: 915.0, BW: 0, SF: 7, CR: 5, TX: 20},
			valid: false,
		},
		{
			name:  "Negative BW",
			radio: RadioInfo{Freq: 915.0, BW: -125.0, SF: 7, CR: 5, TX: 20},
			valid: false,
		},
		{
			name:  "Zero SF",
			radio: RadioInfo{Freq: 915.0, BW: 125.0, SF: 0, CR: 5, TX: 20},
			valid: false,
		},
		{
			name:  "Zero CR",
			radio: RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 0, TX: 20},
			valid: false,
		},
		{
			name:  "Zero TX",
			radio: RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 0},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.radio)
			if tt.valid && err != nil {
				t.Errorf("Expected valid radio info, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid radio info, got no error")
			}
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	validRadio := RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20}

	tests := []struct {
		name     string
		metadata Metadata
		valid    bool
	}{
		{
			name:     "Valid metadata",
			metadata: Metadata{Name: "test-node", Pubkey: "abc123", Radio: validRadio},
			valid:    true,
		},
		{
			name:     "Empty name",
			metadata: Metadata{Name: "", Pubkey: "abc123", Radio: validRadio},
			valid:    false,
		},
		{
			name:     "Empty pubkey",
			metadata: Metadata{Name: "test-node", Pubkey: "", Radio: validRadio},
			valid:    false,
		},
		{
			name:     "Invalid radio",
			metadata: Metadata{Name: "test-node", Pubkey: "abc123", Radio: RadioInfo{Freq: 915.0, BW: 0, SF: 7, CR: 5, TX: 20}},
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.metadata)
			if tt.valid && err != nil {
				t.Errorf("Expected valid metadata, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid metadata, got no error")
			}
		})
	}
}

func TestValidateDeviceData(t *testing.T) {
	tests := []struct {
		name  string
		data  DeviceData
		valid bool
	}{
		{
			name: "Valid device data",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				RSSI:       -50,
				SNR:        8.5,
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: true,
		},
		{
			name: "Empty device ID",
			data: DeviceData{
				DeviceID:   "",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Empty device name",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Invalid timestamp",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "invalid",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Latitude too low",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   -91,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Latitude too high",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   91,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Longitude too low",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  -181,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Longitude too high",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  181,
				ScanSource: "active_ping_response",
			},
			valid: false,
		},
		{
			name: "Empty scan source",
			data: DeviceData{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.data)
			if tt.valid && err != nil {
				t.Errorf("Expected valid device data, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid device data, got no error")
			}
		})
	}
}

func TestValidateReportRequest(t *testing.T) {
	validMetadata := Metadata{
		Name:   "test-node",
		Pubkey: "abc123",
		Radio:  RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20},
	}

	validDeviceData := DeviceData{
		DeviceID:   "device-123",
		DeviceName: "Test Device",
		Timestamp:  "2026-01-16T21:41:52.615226",
		Latitude:   42.6674757,
		Longitude:  23.2714001,
		ScanSource: "active_ping_response",
	}

	tests := []struct {
		name   string
		report ReportRequest
		valid  bool
	}{
		{
			name: "Valid report",
			report: ReportRequest{
				Metadata: validMetadata,
				Data:     []DeviceData{validDeviceData},
			},
			valid: true,
		},
		{
			name: "Empty data array",
			report: ReportRequest{
				Metadata: validMetadata,
				Data:     []DeviceData{},
			},
			valid: false,
		},
		{
			name: "Invalid metadata",
			report: ReportRequest{
				Metadata: Metadata{Name: "", Pubkey: "abc123", Radio: RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20}},
				Data:     []DeviceData{validDeviceData},
			},
			valid: false,
		},
		{
			name: "Invalid device data in array",
			report: ReportRequest{
				Metadata: validMetadata,
				Data: []DeviceData{
					{
						DeviceID:   "",
						DeviceName: "Test Device",
						Timestamp:  "2026-01-16T21:41:52.615226",
						Latitude:   42.6674757,
						Longitude:  23.2714001,
						ScanSource: "active_ping_response",
					},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.report)
			if tt.valid && err != nil {
				t.Errorf("Expected valid report, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid report, got no error")
			}
		})
	}
}

func TestHandleReport(t *testing.T) {
	router := gin.New()
	router.POST("/report", handleReport)

	validReport := ReportRequest{
		Metadata: Metadata{
			Name:   "test-node",
			Pubkey: "abc123",
			Radio:  RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20},
		},
		Data: []DeviceData{
			{
				DeviceID:   "device-123",
				DeviceName: "Test Device",
				RSSI:       -50,
				SNR:        8.5,
				Timestamp:  "2026-01-16T21:41:52.615226",
				Latitude:   42.6674757,
				Longitude:  23.2714001,
				ScanSource: "active_ping_response",
			},
		},
	}

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "Valid report",
			payload:        validReport,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid JSON",
			payload: map[string]interface{}{
				"metadata": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty data array",
			payload: ReportRequest{
				Metadata: Metadata{
					Name:   "test-node",
					Pubkey: "abc123",
					Radio:  RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20},
				},
				Data: []DeviceData{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid latitude",
			payload: ReportRequest{
				Metadata: Metadata{
					Name:   "test-node",
					Pubkey: "abc123",
					Radio:  RadioInfo{Freq: 915.0, BW: 125.0, SF: 7, CR: 5, TX: 20},
				},
				Data: []DeviceData{
					{
						DeviceID:   "device-123",
						DeviceName: "Test Device",
						Timestamp:  "2026-01-16T21:41:52.615226",
						Latitude:   91,
						Longitude:  23.2714001,
						ScanSource: "active_ping_response",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/report", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
