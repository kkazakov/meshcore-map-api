CREATE TABLE IF NOT EXISTS repeater_reports
(
    timestamp DateTime64(6, 'UTC') CODEC(Delta, ZSTD(1)),
    
    repeater_name LowCardinality(String) CODEC(ZSTD(1)),
    repeater_pubkey FixedString(64) CODEC(ZSTD(1)),
    
    reporter_name LowCardinality(String) CODEC(ZSTD(1)),
    reporter_pubkey FixedString(64) CODEC(ZSTD(1)),
    
    radio_freq Float32 CODEC(ZSTD(1)),
    radio_bw Float32 CODEC(ZSTD(1)),
    radio_sf UInt8 CODEC(ZSTD(1)),
    radio_cr UInt8 CODEC(ZSTD(1)),
    radio_tx UInt8 CODEC(ZSTD(1)),
    
    device_id LowCardinality(String) CODEC(ZSTD(1)),
    device_name LowCardinality(String) CODEC(ZSTD(1)),
    
    rssi Int16 CODEC(ZSTD(1)),
    snr Float32 CODEC(ZSTD(1)),
    
    latitude Nullable(Float64) CODEC(Gorilla, ZSTD(1)),
    longitude Nullable(Float64) CODEC(Gorilla, ZSTD(1)),
    
    geohash String CODEC(ZSTD(1)),
    
    city_code LowCardinality(FixedString(3)) CODEC(ZSTD(1)),
    district_code LowCardinality(FixedString(3)) CODEC(ZSTD(1)),
    country_code LowCardinality(FixedString(2)) CODEC(ZSTD(1)),
    
    scan_source LowCardinality(String) CODEC(ZSTD(1)),
    
    ingested_at DateTime DEFAULT now() CODEC(Delta, ZSTD(1))
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (repeater_pubkey, toStartOfHour(timestamp), geohash, device_id)
TTL timestamp + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

ALTER TABLE repeater_reports 
    ADD INDEX idx_geohash geohash TYPE bloom_filter GRANULARITY 4;
