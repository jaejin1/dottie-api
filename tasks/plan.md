# Phase 3 Plan — 기록 핵심 (dots, daylogs)

## 의존성 그래프

```
external clients
  internal/pkg/mapbox/geocoding.go     ← Mapbox API (best-effort)
  internal/pkg/storage/r2_client.go    ← Cloudflare R2 presigned URL

repositories (sqlc Querier 래핑)
  internal/repository/daylog_repo.go   ← CreateDayLog, GetByID, GetByUserAndDate, ListByUser, End
  internal/repository/dot_repo.go      ← Create, GetByDayLog, GetByUserAndDate

services
  internal/service/daylog_service.go   ← daylog_repo + dot_repo (통계 계산)
  internal/service/dot_service.go      ← dot_repo + daylog_repo + mapbox
  internal/service/media_service.go    ← r2_client + user_repo

handlers
  internal/handler/recording_handler.go ← daylog_service
  internal/handler/dot_handler.go       ← dot_service
  internal/handler/daylog_handler.go    ← daylog_service
  internal/handler/media_handler.go     ← media_service

wiring
  internal/server/server.go            ← 신규 핸들러 등록
  internal/model/dto/{request,response}.go ← Phase 3 DTO 추가
```

## sqlc 타입 핵심 (생성 코드에서 확인)

```go
// DayLog
CreateDayLogParams { UserID pgtype.UUID; Date pgtype.Date; StartedAt pgtype.Timestamptz }
EndDayLogParams    { ID pgtype.UUID; EndedAt pgtype.Timestamptz; TotalDistanceM pgtype.Float8; PlaceCount pgtype.Int4; TotalDurationSec pgtype.Int4 }
ListDayLogsByUserParams { UserID pgtype.UUID; Limit int32; Offset int32 }

// Dot
CreateDotParams { DayLogID, UserID pgtype.UUID; Latitude, Longitude float64; Timestamp pgtype.Timestamptz; PlaceName, PlaceCategory, Memo, Emotion pgtype.Text }
GetDotsByUserAndDateParams { UserID pgtype.UUID; Date pgtype.Date }
```

---

## Task 1 — Recording Start/End (DayLog 핵심 경로)

**목표:** `POST /v1/recordings/start`, `POST /v1/recordings/end` 동작

**새 파일:**
- `internal/repository/daylog_repo.go`
- `internal/repository/dot_repo.go` (end에서 dots 조회 필요)
- `internal/service/daylog_service.go` (StartRecording, EndRecording + Haversine 통계)
- `internal/handler/recording_handler.go`
- `internal/model/dto/` — StartRecordingRequest, EndRecordingRequest, DayLogResponse 추가

**Haversine 공식 (Go):**
```go
func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
    const R = 6371000.0
    φ1, φ2 := lat1*math.Pi/180, lat2*math.Pi/180
    Δφ := (lat2-lat1) * math.Pi / 180
    Δλ := (lng2-lng1) * math.Pi / 180
    a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
    return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
```

**에러 케이스:**
- `409 ALREADY_RECORDING` — (user_id, date) 중복 시 pgx unique violation 감지
- `403 FORBIDDEN` — day_log.UserID != req user
- `409 ALREADY_ENDED` — is_recording == false

**서버 라우트 추가:**
```go
recordings := v1.Group("/recordings", middleware.Auth(fbClient))
recordings.POST("/start", recordingHandler.Start)
recordings.POST("/end", recordingHandler.End)
```

**검증:**
```bash
curl -X POST /v1/recordings/start -d '{"date":"2026-04-27"}' -H "Authorization: Bearer <token>"
# → 201 { "data": { "id": "...", "is_recording": true } }
curl -X POST /v1/recordings/start -d '{"date":"2026-04-27"}' -H "..."
# → 409 ALREADY_RECORDING
curl -X POST /v1/recordings/end -d '{"day_log_id":"<id>"}' -H "..."
# → 200 { "data": { "is_recording": false, "total_distance_m": 0, ... } }
```

---

## Task 2 — Dot 단건 저장 + 날짜별 조회

**목표:** `POST /v1/dots`, `GET /v1/dots?date=`

**새 파일:**
- `internal/pkg/mapbox/geocoding.go` (ReverseGeocode — best-effort)
- `internal/service/dot_service.go` (CreateDot, GetDotsByDate)
- `internal/handler/dot_handler.go`

**Mapbox 클라이언트 인터페이스:**
```go
type GeocodingClient interface {
    ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodingResult, error)
}
type GeocodingResult struct { PlaceName string; Category string }
```
config.MapboxAccessToken 미설정 시 nil 허용 — dot 저장은 정상 진행.

**dot_service.CreateDot 로직:**
1. daylog_repo.GetByID → 소유자 확인
2. mapbox.ReverseGeocode (place_name == "" 일 때만, 실패 시 무시)
3. dot_repo.Create

**서버 라우트:**
```go
dots := v1.Group("/dots", middleware.Auth(fbClient))
dots.POST("", dotHandler.CreateDot)
dots.POST("/batch", dotHandler.CreateDotsBatch)
dots.GET("", dotHandler.GetDotsByDate)
```

**검증:**
```bash
curl -X POST /v1/dots -d '{"day_log_id":"<id>","latitude":37.5665,"longitude":126.978,"timestamp":"..."}'
# → 201 { "data": { "id": "...", "photo_upload_url": null } }
curl /v1/dots?date=2026-04-27
# → 200 { "data": [...] }
```

---

## Task 3 — Dot 일괄 저장 (batch)

**목표:** `POST /v1/dots/batch`

**기존 파일 수정:**
- `dot_service.go` — CreateDotsBatch 메서드 추가
- `dot_handler.go` — CreateDotsBatch 핸들러 추가

**로직:**
- 50개 초과 → `400 BATCH_LIMIT_EXCEEDED`
- 각 dot 개별 Create — 실패 시 failed[] 추가, 전체 롤백 없음
- Mapbox 호출 없음

**검증:**
```bash
curl -X POST /v1/dots/batch -d '{"day_log_id":"<id>","dots":[{"client_id":"c1","latitude":...},...]}'
# → 200 { "data": { "synced": [{"client_id":"c1","server_id":"..."}], "failed": [] } }
curl -X POST /v1/dots/batch -d '{"day_log_id":"<id>","dots":[...51개...]}'
# → 400 BATCH_LIMIT_EXCEEDED
```

---

## Task 4 — DayLog 목록 + 상세

**목표:** `GET /v1/daylogs`, `GET /v1/daylogs/:id`

**새 파일:**
- `internal/handler/daylog_handler.go`
- `internal/service/daylog_service.go` — ListDayLogs, GetDayLogWithDots 추가

**DayLogWithDots response:**
```go
type DayLogDetailResponse struct {
    DayLogResponse
    Dots []DotResponse `json:"dots"`
}
```

**서버 라우트:**
```go
daylogs := v1.Group("/daylogs", middleware.Auth(fbClient))
daylogs.GET("", daylogHandler.List)
daylogs.GET("/:id", daylogHandler.GetDetail)
```

**검증:**
```bash
curl /v1/daylogs?limit=20&offset=0
# → 200 { "data": [...], "meta": {"limit":20,"offset":0} }
curl /v1/daylogs/<id>
# → 200 { "data": { "id":"...", "dots": [...] } }
curl /v1/daylogs/<타인id>
# → 403 FORBIDDEN
```

---

## Task 5 — Media Upload (R2 Presigned URL)

**목표:** `POST /v1/media/upload`

**새 파일:**
- `internal/pkg/storage/r2_client.go`
- `internal/service/media_service.go`
- `internal/handler/media_handler.go`

**R2 클라이언트:**
```go
type StorageClient interface {
    GenerateUploadURL(ctx context.Context, key, contentType string) (string, error)
    GetPublicURL(key string) string
}
```
R2 설정 미완료 시 nil → `503 STORAGE_NOT_CONFIGURED`

**파일 경로:** `users/<user_id>/dots/<YYYYMMDD>/<uuid>.<ext>`
content_type → ext 매핑: jpeg→jpg, png→png, heic→heic

**서버 라우트:**
```go
media := v1.Group("/media", middleware.Auth(fbClient))
media.POST("/upload", mediaHandler.Upload)
```

**검증:**
```bash
curl -X POST /v1/media/upload -d '{"content_type":"image/jpeg","file_size":1024}'
# → 200 { "data": { "upload_url":"...", "public_url":"...", "expires_in":3600 } }
curl -X POST /v1/media/upload -d '{"content_type":"image/gif","file_size":1024}'
# → 400 INVALID_FILE_TYPE
curl -X POST /v1/media/upload -d '{"content_type":"image/jpeg","file_size":11000000}'
# → 400 FILE_TOO_LARGE
```

---

## Checkpoint — 전체 빌드 검증

각 Task 완료 후:
```bash
go build ./...   # 컴파일 에러 없음
```

전체 완료 후:
```bash
make dev         # 서버 정상 기동
curl /health     # 200 ok
```
