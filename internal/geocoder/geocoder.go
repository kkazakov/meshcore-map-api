package geocoder

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
)

type city struct {
	name        string
	countryCode string
	lat         float32
	lon         float32
	admin1Code  string
}

type gridCell struct {
	cities []city
}

type Geocoder struct {
	grid     map[int]map[int]*gridCell
	gridSize float64
	mu       sync.RWMutex
}

var instance *Geocoder
var once sync.Once

func GetInstance() *Geocoder {
	once.Do(func() {
		instance = &Geocoder{
			grid:     make(map[int]map[int]*gridCell),
			gridSize: 1.0,
		}
		if err := instance.loadCities(); err != nil {
			panic(err)
		}
	})
	return instance
}

func (g *Geocoder) getGridKey(lat, lon float64) (int, int) {
	gridLat := int(math.Floor(lat / g.gridSize))
	gridLon := int(math.Floor(lon / g.gridSize))
	return gridLat, gridLon
}

func (g *Geocoder) loadCities() error {
	dataPath := "data/cities15000.txt"
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "../../data/cities15000.txt"
	}

	file, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	g.mu.Lock()
	defer g.mu.Unlock()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 11 {
			continue
		}

		lat, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			continue
		}

		lon, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			continue
		}

		c := city{
			name:        fields[1],
			countryCode: fields[8],
			lat:         float32(lat),
			lon:         float32(lon),
			admin1Code:  fields[10],
		}

		gridLat, gridLon := g.getGridKey(lat, lon)

		if g.grid[gridLat] == nil {
			g.grid[gridLat] = make(map[int]*gridCell)
		}
		if g.grid[gridLat][gridLon] == nil {
			g.grid[gridLat][gridLon] = &gridCell{
				cities: make([]city, 0, 10),
			}
		}

		g.grid[gridLat][gridLon].cities = append(g.grid[gridLat][gridLon].cities, c)
	}

	return scanner.Err()
}

func (g *Geocoder) ReverseGeocode(lat, lon float64) (cityCode, districtCode, countryCode string) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	gridLat, gridLon := g.getGridKey(lat, lon)

	minDist := math.MaxFloat64
	var nearestCity *city

	for dLat := -1; dLat <= 1; dLat++ {
		for dLon := -1; dLon <= 1; dLon++ {
			checkLat := gridLat + dLat
			checkLon := gridLon + dLon

			if g.grid[checkLat] == nil || g.grid[checkLat][checkLon] == nil {
				continue
			}

			cell := g.grid[checkLat][checkLon]
			for i := range cell.cities {
				dist := haversine(lat, lon, float64(cell.cities[i].lat), float64(cell.cities[i].lon))
				if dist < minDist {
					minDist = dist
					nearestCity = &cell.cities[i]
				}
			}
		}
	}

	if nearestCity == nil {
		return "", "", ""
	}

	cityName := nearestCity.name
	if len(cityName) > 3 {
		cityCode = cityName[:3]
	} else {
		cityCode = cityName
	}

	districtCode = nearestCity.admin1Code
	if len(districtCode) > 3 {
		districtCode = districtCode[:3]
	}

	countryCode = nearestCity.countryCode
	if len(countryCode) > 2 {
		countryCode = countryCode[:2]
	}

	return strings.ToUpper(cityCode), strings.ToUpper(districtCode), strings.ToUpper(countryCode)
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	lat1 = lat1 * math.Pi / 180.0
	lat2 = lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
