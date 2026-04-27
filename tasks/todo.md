# Phase 3 Todo

## Task 1 — Recording Start/End
- [ ] `internal/repository/daylog_repo.go`
- [ ] `internal/repository/dot_repo.go`
- [ ] `internal/model/dto/` — DayLog/Recording DTO 추가
- [ ] `internal/service/daylog_service.go` (StartRecording, EndRecording + Haversine)
- [ ] `internal/handler/recording_handler.go`
- [ ] `internal/server/server.go` — /recordings 라우트 등록
- [ ] 빌드 확인: `go build ./...`

## Task 2 — Dot 단건 + GET
- [ ] `internal/pkg/mapbox/geocoding.go`
- [ ] `internal/model/dto/` — Dot DTO 추가
- [ ] `internal/service/dot_service.go` (CreateDot, GetDotsByDate)
- [ ] `internal/handler/dot_handler.go` (CreateDot, GetDotsByDate)
- [ ] `internal/server/server.go` — /dots 라우트 등록
- [ ] 빌드 확인: `go build ./...`

## Task 3 — Dot batch
- [ ] `dot_service.go` — CreateDotsBatch 추가
- [ ] `dot_handler.go` — CreateDotsBatch 추가
- [ ] 빌드 확인: `go build ./...`

## Task 4 — DayLog 목록 + 상세
- [ ] `internal/service/daylog_service.go` — ListDayLogs, GetDayLogWithDots 추가
- [ ] `internal/handler/daylog_handler.go`
- [ ] `internal/server/server.go` — /daylogs 라우트 등록
- [ ] 빌드 확인: `go build ./...`

## Task 5 — Media Upload
- [ ] `internal/pkg/storage/r2_client.go`
- [ ] `internal/service/media_service.go`
- [ ] `internal/handler/media_handler.go`
- [ ] `internal/server/server.go` — /media 라우트 등록
- [ ] 빌드 확인: `go build ./...`

## Final Checkpoint
- [ ] `go build ./...` 전체 빌드 성공
- [ ] `make dev` 서버 기동 성공
- [ ] 라우트별 smoke test (curl)
