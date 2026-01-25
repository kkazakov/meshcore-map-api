CREATE TABLE IF NOT EXISTS repeaters
(
    public_key FixedString(64) CODEC(ZSTD(1)),
    name LowCardinality(String) CODEC(ZSTD(1)),
    lat Nullable(Float64) CODEC(Gorilla, ZSTD(1)),
    lon Nullable(Float64) CODEC(Gorilla, ZSTD(1)),
    
    created_date DateTime DEFAULT now() CODEC(Delta, ZSTD(1)),
    updated_at DateTime DEFAULT now() CODEC(Delta, ZSTD(1))
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_date)
ORDER BY (public_key)
SETTINGS index_granularity = 8192;
