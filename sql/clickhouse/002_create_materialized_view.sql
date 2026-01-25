CREATE MATERIALIZED VIEW IF NOT EXISTS repeater_reports_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (repeater_pubkey, hour, device_id)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    repeater_pubkey,
    repeater_name,
    device_id,
    device_name,
    geohash,
    region_code,
    district_code,
    country_code,
    count() AS report_count,
    avg(rssi) AS avg_rssi,
    avg(snr) AS avg_snr,
    min(rssi) AS min_rssi,
    max(rssi) AS max_rssi
FROM repeater_reports
GROUP BY hour, repeater_pubkey, repeater_name, device_id, device_name, 
          geohash, region_code, district_code, country_code;
