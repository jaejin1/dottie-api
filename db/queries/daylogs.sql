-- name: CreateDayLog :one
INSERT INTO day_logs (user_id, date, started_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDayLogByID :one
SELECT * FROM day_logs WHERE id = $1;

-- name: GetDayLogByUserAndDate :one
SELECT * FROM day_logs WHERE user_id = $1 AND date = $2;

-- name: ListDayLogsByUser :many
SELECT * FROM day_logs
WHERE user_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: EndDayLog :one
UPDATE day_logs
SET ended_at = $2, is_recording = false,
    total_distance_m = $3, place_count = $4, total_duration_sec = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
