# Meshcore Map API

Lightweight HTTP API server for processing and storing Meshcore repeater reports.

## Setup

### Prerequisites

- Go 1.25.5+
- ClickHouse server (see `docker/clickhouse/docker-compose.yaml`)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure ClickHouse connection
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
