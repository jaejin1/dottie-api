package dto

import "time"

// ── DayLog ─────────────────────────────────────────────────

type DayLogResponse struct {
	ID               string     `json:"id"`
	Date             string     `json:"date"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at"`
	IsRecording      bool       `json:"is_recording"`
	TotalDistanceM   *float64   `json:"total_distance_m"`
	PlaceCount       *int32     `json:"place_count"`
	TotalDurationSec *int32     `json:"total_duration_sec"`
}

type DayLogDetailResponse struct {
	DayLogResponse
	Dots []DotResponse `json:"dots"`
}

// ── Dot ────────────────────────────────────────────────────

type DotResponse struct {
	ID            string     `json:"id"`
	DayLogID      string     `json:"day_log_id"`
	Latitude      float64    `json:"latitude"`
	Longitude     float64    `json:"longitude"`
	Timestamp     time.Time  `json:"timestamp"`
	PlaceName     *string    `json:"place_name"`
	PlaceCategory *string    `json:"place_category"`
	PhotoURL      *string    `json:"photo_url"`
	Memo          *string    `json:"memo"`
	Emotion       *string    `json:"emotion"`
}

type CreateDotResponse struct {
	ID            string  `json:"id"`
	PhotoUploadURL *string `json:"photo_upload_url"`
}

type BatchSyncedItem struct {
	ClientID string `json:"client_id"`
	ServerID string `json:"server_id"`
}

type BatchFailedItem struct {
	ClientID string `json:"client_id"`
	Reason   string `json:"reason"`
}

type CreateDotsBatchResponse struct {
	Synced []BatchSyncedItem `json:"synced"`
	Failed []BatchFailedItem `json:"failed"`
}

// ── Media ──────────────────────────────────────────────────

type MediaUploadResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	ExpiresIn int    `json:"expires_in"`
}

type CharacterConfig struct {
	Color      string `json:"color"`
	Accessory  string `json:"accessory"`
	Expression string `json:"expression"`
}

type UserResponse struct {
	ID              string          `json:"id"`
	Nickname        string          `json:"nickname"`
	ProfileImage    *string         `json:"profile_image"`
	CharacterConfig CharacterConfig `json:"character_config"`
	Provider        string          `json:"provider"`
	CreatedAt       time.Time       `json:"created_at"`
}

type LoginResponse struct {
	FirebaseCustomToken string       `json:"firebase_custom_token"`
	User                UserResponse `json:"user"`
	IsNew               bool         `json:"is_new"`
}
