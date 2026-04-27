package dto

// ── Recording ──────────────────────────────────────────────

type StartRecordingRequest struct {
	Date string `json:"date"` // "YYYY-MM-DD"
}

type EndRecordingRequest struct {
	DayLogID string `json:"day_log_id"`
}

// ── Dot ────────────────────────────────────────────────────

type CreateDotRequest struct {
	DayLogID      string  `json:"day_log_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timestamp     string  `json:"timestamp"` // RFC3339
	PlaceName     *string `json:"place_name"`
	PlaceCategory *string `json:"place_category"`
	Memo          *string `json:"memo"`
	Emotion       *string `json:"emotion"`
}

type BatchDotItem struct {
	ClientID      string  `json:"client_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timestamp     string  `json:"timestamp"`
	PlaceName     *string `json:"place_name"`
	PlaceCategory *string `json:"place_category"`
	Memo          *string `json:"memo"`
	Emotion       *string `json:"emotion"`
}

type CreateDotsBatchRequest struct {
	DayLogID string         `json:"day_log_id"`
	Dots     []BatchDotItem `json:"dots"`
}

// ── Media ──────────────────────────────────────────────────

type MediaUploadRequest struct {
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

// ── Auth / User (기존) ─────────────────────────────────────

type LoginRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
}

type UpdateUserRequest struct {
	Nickname     string  `json:"nickname"`
	ProfileImage *string `json:"profile_image"`
}

type UpdateCharacterRequest struct {
	Color      string `json:"color"`
	Accessory  string `json:"accessory"`
	Expression string `json:"expression"`
}
