# Meshcore Map API

Lightweight HTTP API server for processing and storing Meshcore repeater reports.

## Setup

### Prerequisites

- Go 1.25.5+
- ClickHouse server (see `docker/clickhouse/docker-compose.yaml`)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure environment variables (see Configuration section)
3. Download geocoding data:
```bash
mkdir -p data
curl -L -o data/cities15000.zip https://download.geonames.org/export/dump/cities15000.zip
cd data && unzip cities15000.zip && cd ..
```

4. Install dependencies:
```bash
go mod download
```

5. Build the server:
```bash
go build -o server
```

### Running

```bash
./server
```

The server runs on port 8080 by default.

## Configuration

The application uses environment variables configured in the `.env` file:

### ClickHouse Connection

- `CLICKHOUSE_HOST` - ClickHouse server hostname (default: localhost)
- `CLICKHOUSE_PORT` - ClickHouse server port (default: 9000)
- `CLICKHOUSE_DATABASE` - Database name (default: meshcore)
- `CLICKHOUSE_USER` - Database username
- `CLICKHOUSE_PASSWORD` - Database password

### Privacy Settings

- `STORE_PRECISE_LOCATION` - Controls storage of precise coordinates (default: true)
  - `true` - Stores exact latitude and longitude values
  - `false` - Stores NULL for lat/lon, only geohash is saved (privacy mode)
  
Note: Geohash is always calculated and stored regardless of this setting, providing approximate location data with 8-character precision.

## Features

- Validates and stores repeater reports
- Automatic reverse geocoding (city, district, country codes)
- Geohash generation for efficient spatial queries
- Offline geocoding using GeoNames data (loaded once at startup)
- Optimized memory usage with spatial grid indexing (~13 MB for 33K cities)

## API Endpoints

### POST /report

Submit a repeater report with device data.

## Development

See `AGENTS.md` for detailed development guidelines.
