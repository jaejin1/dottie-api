# SPEC.md — Phase 3: 기록 핵심 (dots, daylogs)

## 1. Objective

하루 이동 기록(DayLog)을 시작·종료하고, 위치 점(Dot)을 단건/일괄로 저장·조회하며,
사진 업로드용 presigned URL을 발급한다.
Mapbox 역지오코딩으로 Dot 저장 시 장소명을 자동 채운다.

---

## 2. Current State

**이미 완료:**
- DB 스키마: `day_logs`, `dots` 테이블 및 PostGIS 인덱스
- sqlc 쿼리: `CreateDayLog`, `EndDayLog`, `ListDayLogsByUser`, `GetDayLogByID`, `GetDayLogByUserAndDate`, `CreateDot`, `GetDotsByDayLog`, `GetDotsByUserAndDate`, `UpdateDotPhotoURL`
- Config: `MapboxAccessToken`, `R2AccountID/AccessKeyID/SecretAccessKey/BucketName/PublicURL` 모두 준비
- Auth 미들웨어, 레이어 패턴(repository/service/handler) 확립

**미완료 (Phase 3에서 구현):**
- `internal/pkg/mapbox/`, `internal/pkg/storage/` 없음
- daylog/dot/media 레이어 전무
- statistics 계산 로직 없음

---

## 3. Endpoints

### POST /v1/recordings/start — 기록 시작
```
인증 필요
```
**Request:**
```json
{ "date": "2026-04-27" }
```
**Response 201:**
```json
{ "data": { "id": "uuid", "date": "2026-04-27", "started_at": "...", "is_recording": true } }
```
**서버 로직:**
1. firebase_uid → user.id 조회
2. `day_logs`에서 `(user_id, date)` UNIQUE 제약 — 이미 있으면 `409 ALREADY_RECORDING`
3. `CreateDayLog` 실행

---

### POST /v1/recordings/end — 기록 종료
```
인증 필요
```
**Request:**
```json
{ "day_log_id": "uuid" }
```
**Response 200:**
```json
{
  "data": {
    "id": "uuid",
    "ended_at": "...",
    "is_recording": false,
    "total_distance_m": 12500.5,
    "place_count": 8,
    "total_duration_sec": 43200
  }
}
```
**서버 로직:**
1. day_log 소유자 확인 (user_id 불일치 → 403)
2. 이미 종료됐으면 (`is_recording = false`) → `409 ALREADY_ENDED`
3. 통계 계산 (Go 레벨):
   - `GetDotsByDayLog`로 전체 dot 조회
   - `total_distance_m`: 순차 dot 간 Haversine 거리 합산
   - `place_count`: `place_name != null` 고유 개수
   - `total_duration_sec`: max(timestamp) - min(timestamp) 초
4. `EndDayLog` 실행

---

### POST /v1/dots — 단일 Dot 저장
```
인증 필요
```
**Request:**
```json
{
  "day_log_id": "uuid",
  "latitude": 37.5665,
  "longitude": 126.9780,
  "timestamp": "2026-04-27T14:00:00+09:00",
  "place_name": null,
  "place_category": null,
  "memo": null,
  "emotion": null
}
```
**Response 201:**
```json
{ "data": { "id": "uuid", "photo_upload_url": null } }
```
**서버 로직:**
1. day_log 소유자 확인
2. `place_name == null` 이면 Mapbox 역지오코딩 호출 (실패 시 null로 저장, 에러 전파 안 함)
3. `CreateDot` 실행 (PostGIS Point는 SQL에서 처리)

---

### POST /v1/dots/batch — 일괄 Dot 저장 (오프라인 동기화)
```
인증 필요
```
**Request:**
```json
{
  "day_log_id": "uuid",
  "dots": [
    { "client_id": "local-uuid-1", "latitude": 37.5665, "longitude": 126.9780,
      "timestamp": "...", "place_name": null, "memo": null, "emotion": null }
  ]
}
```
최대 50개 제한.

**Response 200:**
```json
{
  "data": {
    "synced": [{ "client_id": "local-uuid-1", "server_id": "uuid-abc" }],
    "failed": []
  }
}
```
**서버 로직:**
- 각 dot을 순서대로 `CreateDot` 실행
- 개별 실패 시 `failed` 배열에 추가, 전체 롤백 없음
- Mapbox 호출 없음 (배치에선 생략 — 성능 우선)

---

### GET /v1/dots?date=2026-04-27 — 날짜별 Dot 조회
```
인증 필요
```
**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid", "latitude": 37.5665, "longitude": 126.9780,
      "timestamp": "...", "place_name": "광화문", "place_category": "cafe",
      "photo_url": null, "memo": null, "emotion": "happy"
    }
  ]
}
```
`date` 파라미터 없으면 오늘 날짜 기본.

---

### GET /v1/daylogs — DayLog 목록
```
인증 필요, 페이지네이션: ?limit=20&offset=0
```
**Response 200:**
```json
{
  "data": [
    { "id": "uuid", "date": "2026-04-27", "started_at": "...", "ended_at": "...",
      "is_recording": false, "total_distance_m": 12500, "place_count": 8 }
  ],
  "meta": { "limit": 20, "offset": 0 }
}
```

---

### GET /v1/daylogs/:id — DayLog 상세
```
인증 필요
```
**Response 200:** DayLog 단건 + dots 배열 포함
```json
{
  "data": {
    "id": "uuid", "date": "2026-04-27", "...",
    "dots": [ { "id": "uuid", "latitude": ..., ... } ]
  }
}
```
소유자 불일치 → 403.

---

### POST /v1/media/upload — Presigned URL 발급
```
인증 필요
```
**Request:**
```json
{ "content_type": "image/jpeg", "file_size": 2048576 }
```
**Response 200:**
```json
{
  "data": {
    "upload_url": "https://...",
    "public_url": "https://media.dottie.app/...",
    "expires_in": 3600
  }
}
```
**서버 로직:**
1. `file_size` > 10MB → `400 FILE_TOO_LARGE`
2. `content_type` ∉ {image/jpeg, image/png, image/heic} → `400 INVALID_FILE_TYPE`
3. 파일 경로: `users/<user_id>/dots/<YYYYMMDD>/<uuid>.<ext>`
4. R2 presigned PUT URL 생성 (1시간 유효)

---

## 4. Project Structure (신규 파일)

```
internal/
  pkg/
    mapbox/
      geocoding.go       ← ReverseGeocode(ctx, lat, lng) → place_name, category
    storage/
      r2_client.go       ← GenerateUploadURL, GetPublicURL
  repository/
    daylog_repo.go       ← Create, GetByID, GetByUserAndDate, ListByUser, End
    dot_repo.go          ← Create, GetByDayLog, GetByUserAndDate, UpdatePhotoURL
  service/
    daylog_service.go    ← StartRecording, EndRecording, GetDayLog, ListDayLogs, GetDayLogWithDots
    dot_service.go       ← CreateDot, CreateDotsBatch, GetDotsByDate
    media_service.go     ← GenerateUploadURL
  handler/
    recording_handler.go ← POST /recordings/start, POST /recordings/end
    dot_handler.go       ← POST /dots, POST /dots/batch, GET /dots
    daylog_handler.go    ← GET /daylogs, GET /daylogs/:id
    media_handler.go     ← POST /media/upload
  model/dto/
    request.go           ← 기존 파일에 Phase 3 DTO 추가
    response.go          ← 기존 파일에 Phase 3 DTO 추가
```

---

## 5. Statistics 계산 (Go 레벨 — Haversine)

`EndRecording` 서비스에서 직접 계산:

```go
// total_distance_m: 연속 dot 간 거리 합산
func haversineMeters(lat1, lng1, lat2, lng2 float64) float64

// place_count: place_name != "" 인 고유 place_name 수
// total_duration_sec: last.timestamp - first.timestamp (초)
```

PostGIS `ST_Distance` 쿼리 대신 Go 레벨 계산 — 쿼리 복잡도 없이 이미 조회된 dots 활용.

---

## 6. Code Style

- Phase 2 패턴 그대로 유지 (Repository interface → Service → Handler)
- Mapbox 실패는 silent: dot 저장은 성공, place_name만 nil
- R2 클라이언트는 config nil 체크 후 nil 허용 (로컬 개발 시 미설정 가능)
- 통계 계산은 service 레이어에서 pure function으로 분리 (테스트 용이)
- DayLog detail(`GET /daylogs/:id`)은 service에서 dots를 별도 조회 후 조합

---

## 7. Boundaries

**Always:**
- Dot 저장 시 day_log 소유자 확인 (firebase_uid → user.id 매칭)
- DayLog end 시 소유자 확인 + `is_recording` 상태 확인
- batch 최대 50개 제한 — 초과 시 `400 BATCH_LIMIT_EXCEEDED`
- R2 파일 크기 10MB, content_type 화이트리스트 검증

**Never:**
- Mapbox 실패로 Dot 저장이 실패해선 안 됨 (best-effort)
- batch에서 Mapbox 호출 (성능 이슈)
- `internal/db/` 직접 편집
