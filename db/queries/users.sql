-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByFirebaseUID :one
SELECT * FROM users WHERE firebase_uid = $1;

-- name: CreateUser :one
INSERT INTO users (firebase_uid, provider, nickname, profile_image, character_config)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET nickname = $2, profile_image = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterConfig :one
UPDATE users
SET character_config = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
