-- name: CreateDot :one
INSERT INTO dots (day_log_id, user_id, location, latitude, longitude, timestamp, place_name, place_category, memo, emotion)
VALUES (
    $1, $2,
    ST_SetSRID(ST_MakePoint($4, $3), 4326),  -- ST_MakePoint(lng, lat)
    $3, $4,
    $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetDotsByDayLog :many
SELECT * FROM dots
WHERE day_log_id = $1
ORDER BY timestamp ASC;

-- name: GetDotsByUserAndDate :many
SELECT d.* FROM dots d
JOIN day_logs dl ON dl.id = d.day_log_id
WHERE d.user_id = $1 AND dl.date = $2
ORDER BY d.timestamp ASC;

-- name: UpdateDotPhotoURL :one
UPDATE dots SET photo_url = $2 WHERE id = $1 RETURNING *;
