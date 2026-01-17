# Meshcore Map API

Lightweight HTTP API server for processing and storing Meshcore repeater reports.

## Setup

### Prerequisites

- Go 1.25.5+
- Docker and Docker daemon running
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

### Running

#### Development Mode

```bash
go run main.go
```

Or build and run locally:

```bash
go build -o server
./server
```

The server runs on port 8080 by default.

#### Production Deployment (Docker)

Deploy using the automated deployment script:

```bash
./deploy.sh
```

This script will:
1. Verify `.env` file exists (required)
2. Build the Go binary
3. Build a Docker image
4. Stop and remove any existing container
5. Start a new container with `--restart unless-stopped` policy

**Prerequisites for deployment:**
- Docker installed and running
- `.env` file configured in project root
- Port 8080 available

**Manage the deployed container:**

```bash
docker logs -f meshcore-map-api
docker stop meshcore-map-api
docker restart meshcore-map-api
docker rm -f meshcore-map-api
```

To redeploy after code changes, simply run `./deploy.sh` again.

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
