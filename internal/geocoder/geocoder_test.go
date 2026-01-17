package geocoder

import (
	"runtime"
	"testing"
)

func TestMemoryUsage(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	geo := GetInstance()

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocatedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024

	t.Logf("Memory used by geocoder: %.2f MB", allocatedMB)
	t.Logf("Grid cells: %d", countGridCells(geo))
}

func countGridCells(g *Geocoder) int {
	count := 0
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, lonMap := range g.grid {
		count += len(lonMap)
	}
	return count
}

func TestReverseGeocode(t *testing.T) {
	geo := GetInstance()

	tests := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Sofia, Bulgaria", 42.6977, 23.3219},
		{"New York, USA", 40.7128, -74.0060},
		{"Tokyo, Japan", 35.6762, 139.6503},
		{"London, UK", 51.5074, -0.1278},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			city, district, country := geo.ReverseGeocode(tt.lat, tt.lon)
			if country == "" {
				t.Errorf("Expected country code, got empty string")
			}
			t.Logf("%s -> City: %s, District: %s, Country: %s", tt.name, city, district, country)
		})
	}
}
