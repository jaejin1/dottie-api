# Dottie API — Memory

## Project Overview
- Go 1.25 REST API for Dottie Flutter app
- Module: `github.com/jaejin1/dottie-api`
- Working dir: `/Users/jaejin/dev/jaejin1/dottie/dottie-api`

## Tech Stack
- Echo v4, PostgreSQL/PostGIS, sqlc, golang-migrate
- Firebase Admin SDK (auth), Cloudflare R2 (media), Mapbox (geocoding)
- zap (logging), viper (config), pgx/v5 (DB driver)

## Phase Status

### ✅ Phase 1 — 프로젝트 초기화 & 인프라 (완료)
- Echo 서버, 미들웨어(Auth/CORS/Logger/RateLimit), Firebase 클라이언트
- DB 스키마 (users, day_logs, dots, rooms, room_members, shared_day_logs)
- docker-compose, Makefile, config, health endpoint

### 🔄 Phase 2 — 인증 & 사용자 (진행 중)

**완료:**
- [x] `POST /v1/auth/login` — 카카오 토큰 검증 → Firebase Custom Token 발급
- [x] `GET /v1/users/me` — 내 프로필 조회
- [x] `PUT /v1/users/me` — 닉네임/프로필 이미지 수정
- [x] `PUT /v1/users/me/character` — 캐릭터 설정 변경
- [x] sqlc 코드 생성 (`internal/db/` 완성)
- [x] DB pool 연결 (`pgxpool`, main.go)

**미완료 (dottie-be-spec.md Phase 2 잔여):**
- [ ] Apple JWT 검증 구현 (`/auth/login` provider: "apple")
- [ ] Google 로그인 (`/auth/login` provider: "google")

### ⬜ Phase 3 — 기록 핵심
- POST /recordings/start, POST /recordings/end
- POST /dots, POST /dots/batch, GET /dots?date=
- GET /daylogs, GET /daylogs/:id
- Mapbox 역지오코딩 연동
- Cloudflare R2 presigned URL (POST /media/upload)

### ⬜ Phase 4 — 소셜
- POST /rooms, GET /rooms, GET /rooms/:id
- POST /rooms/:id/invite, POST /rooms/join
- POST /rooms/:id/share
- GET /rooms/:id/shared-map?date= (만남 감지 포함)

### ⬜ Phase 5 — 배포 & 안정화
- Fly.io 배포, Neon PostgreSQL, Cloudflare R2 프로덕션

## Key File Paths
- Entry: `cmd/server/main.go`
- Config: `internal/config/config.go` (viper, .env)
- Server/Routes: `internal/server/server.go`
- Middleware: `internal/middleware/` (auth, cors, logger, ratelimit)
- Handlers: `internal/handler/` (health, auth, user)
- Services: `internal/service/` (auth_service, user_service)
- Repository: `internal/repository/user_repo.go`
- DTOs: `internal/model/dto/` (request.go, response.go)
- Firebase client: `internal/pkg/firebase/client.go`
- Kakao client: `internal/pkg/kakao/client.go`
- Errors: `internal/pkg/errors/errors.go`
- Migrations: `db/migrations/000001_init.{up,down}.sql`
- sqlc config: `sqlc.yaml` (루트) — generates to `internal/db/`
- sqlc queries: `db/queries/`
- Generated DB code: `internal/db/` (편집 금지)

## DB Local (docker-compose)
- Image: postgis/postgis:17-3.5 (platform: linux/amd64 — Apple Silicon 에뮬레이션)
- DB: dottie / user: dottie / pass: dottie_local / port: 5432
- `make migrate-local` — runs migrations against local DB

## API Conventions
- Base: `/v1`
- Success: `{"data": {...}}`
- Error: `{"error": {"code": "...", "message": "..."}}`
- Auth: `Authorization: Bearer <firebase_id_token>`
- Health: `GET /health`

## Auth 구현 패턴
- `firebase_uid` = `"{provider}:{provider_id}"` (예: `"kakao:12345678"`)
- `/v1/auth/*` — 인증 미들웨어 없음 (public)
- `/v1/users/*` — `middleware.Auth(fbClient)` 적용
- fbClient nil 시 auth 미들웨어 미적용 (로컬 개발 편의, 경고 로그 출력)
- pgx.ErrNoRows → 신규 유저 생성 (get-or-create 패턴)

## sqlc 타입 매핑 주의사항
- `db.User.ID` → `pgtype.UUID` → `.String()` 으로 변환
- `db.User.ProfileImage` → `pgtype.Text` → `.Valid` 체크 후 `*string`
- `db.User.CharacterConfig` → `[]byte` (JSONB) → `json.Marshal/Unmarshal`
- `db.User.CreatedAt` → `pgtype.Timestamptz` → `.Time`
- not found 체크: `errors.Is(err, pgx.ErrNoRows)` (pgx/v5 sentinel)

## Notes
- `sqlc.yaml`은 루트(`/`)에 위치 — `db/sqlc.yaml`은 구버전 (무시)
- `make sqlc` → 루트의 sqlc.yaml 사용, `internal/db/` 재생성
- docker-compose의 `version:` 필드는 obsolete — 경고 무시해도 됨