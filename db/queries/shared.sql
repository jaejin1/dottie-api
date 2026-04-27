-- name: ShareDayLog :one
INSERT INTO shared_day_logs (room_id, day_log_id, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (room_id, day_log_id) DO NOTHING
RETURNING *;

-- name: GetSharedDayLogsByRoomAndDate :many
SELECT sdl.*, dl.date, dl.total_distance_m, dl.place_count, dl.total_duration_sec
FROM shared_day_logs sdl
JOIN day_logs dl ON dl.id = sdl.day_log_id
WHERE sdl.room_id = $1 AND dl.date = $2
ORDER BY sdl.shared_at DESC;

-- name: FindEncounters :many
SELECT
    a.id AS dot_a_id,
    b.id AS dot_b_id,
    a.user_id AS user_a,
    b.user_id AS user_b,
    a.timestamp,
    ST_Distance(a.location::geography, b.location::geography) AS distance_m,
    a.latitude AS lat,
    a.longitude AS lng,
    a.place_name
FROM dots a
JOIN dots b ON a.user_id != b.user_id
    AND a.day_log_id IN (
        SELECT sdl.day_log_id FROM shared_day_logs sdl
        JOIN day_logs dl ON dl.id = sdl.day_log_id
        WHERE sdl.room_id = $1 AND dl.date = $2
    )
    AND b.day_log_id IN (
        SELECT sdl.day_log_id FROM shared_day_logs sdl
        JOIN day_logs dl ON dl.id = sdl.day_log_id
        WHERE sdl.room_id = $1 AND dl.date = $2
    )
    AND ABS(EXTRACT(EPOCH FROM (a.timestamp - b.timestamp))) < 1800
    AND ST_DWithin(a.location::geography, b.location::geography, 100)
WHERE a.user_id < b.user_id  -- 중복 방지
ORDER BY a.timestamp;
