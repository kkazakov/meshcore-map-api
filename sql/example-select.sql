-- Get latest repeater info by public_key using FINAL (simple approach)
-- FINAL forces deduplication at query time
SELECT *
FROM repeaters
FINAL
WHERE public_key = 'a1b2c3d4e5f67890abcdef1234567890abcdef1234567890abcdef1234567890';


-- Get all latest repeaters using FINAL (simple approach)
-- Good for small to medium datasets
SELECT *
FROM repeaters
FINAL
ORDER BY public_key;


-- Get latest repeater info using argMax (performant approach)
-- Recommended for production and large datasets
-- argMax returns the value of the first argument for the row with the maximum value of the second argument
SELECT 
    public_key,
    argMax(name, updated_at) AS name,
    argMax(lat, updated_at) AS lat,
    argMax(lon, updated_at) AS lon,
    min(created_date) AS created_date,
    max(updated_at) AS updated_at
FROM repeaters
WHERE public_key = 'a1b2c3d4e5f67890abcdef1234567890abcdef1234567890abcdef1234567890'
GROUP BY public_key;


-- Get all latest repeaters using argMax (performant approach)
-- Best for production queries with many rows
SELECT 
    public_key,
    argMax(name, updated_at) AS name,
    argMax(lat, updated_at) AS lat,
    argMax(lon, updated_at) AS lon,
    min(created_date) AS created_date,
    max(updated_at) AS updated_at
FROM repeaters
GROUP BY public_key
ORDER BY public_key;


-- Get repeaters within a geographic bounding box (latest version only)
SELECT 
    public_key,
    argMax(name, updated_at) AS name,
    argMax(lat, updated_at) AS lat,
    argMax(lon, updated_at) AS lon,
    argMax(updated_at, updated_at) AS updated_at
FROM repeaters
GROUP BY public_key
HAVING lat BETWEEN 40.0 AND 45.0 
   AND lon BETWEEN 20.0 AND 25.0
ORDER BY updated_at DESC;
