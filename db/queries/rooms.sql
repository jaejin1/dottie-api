-- name: CreateRoom :one
INSERT INTO rooms (name, owner_id) VALUES ($1, $2) RETURNING *;

-- name: GetRoomByID :one
SELECT * FROM rooms WHERE id = $1;

-- name: GetRoomByInviteCode :one
SELECT * FROM rooms WHERE invite_code = $1 AND invite_expires > NOW();

-- name: SetInviteCode :one
UPDATE rooms SET invite_code = $2, invite_expires = $3, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: ListRoomsByUser :many
SELECT r.* FROM rooms r
JOIN room_members rm ON rm.room_id = r.id
WHERE rm.user_id = $1
ORDER BY r.created_at DESC;

-- name: AddRoomMember :exec
INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetRoomMemberCount :one
SELECT COUNT(*) FROM room_members WHERE room_id = $1;

-- name: IsRoomMember :one
SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2);
