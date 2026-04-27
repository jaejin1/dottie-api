# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> Also read **[AGENTS.md](AGENTS.md)** for intent → skill mapping, orchestration rules, and lifecycle conventions (DEFINE → PLAN → BUILD → VERIFY → REVIEW → SHIP).

## Commands

```bash
# 로컬 개발 (DB 자동 시작 포함)
make dev

# 빌드
make build                     # bin/server 생성
go build ./...                 # 전체 패키지 컴파일 확인

# 테스트
go test ./...
go test ./internal/service/... -v -run TestFoo   # 단일 테스트

# DB
make migrate-local             # 로컬 docker-compose DB에 마이그레이션
make migrate-up DB_URL=...     # 임의 DB에 마이그레이션
make migrate-down DB_URL=...   # 마이그레이션 1단계 롤백
make migrate-create name=xxx   # 새 마이그레이션 파일 생성
make db-shell                  # psql 접속

# sqlc 코드 생성 (db/queries/*.sql 변경 후 반드시 실행)
make sqlc                      # → internal/db/ 에 Go 코드 생성

# Docker
docker-compose up -d db        # DB만 실행
docker-compose up              # 전체 실행
make docker-build
```

## Architecture

### Request Flow
```
HTTP Request
  → Echo (middleware chain: Recover → RequestID → Logger → CORS → RateLimit → BodyLimit)
  → Handler (internal/handler/)
  → Service (internal/service/)       ← 비즈니스 로직
  → Repository (internal/repository/) ← sqlc 생성 코드 래핑
  → PostgreSQL/PostGIS
```

### Key Layers

- **`internal/server/server.go`** — Echo 인스턴스 생성, 전체 미들웨어 등록, 라우트 그룹(`/v1`) 설정. 새 핸들러는 여기서 등록.
- **`internal/config/config.go`** — viper로 `.env` + 환경변수 로드. 모든 설정값은 `Config` struct 통해 접근.
- **`internal/middleware/auth.go`** — Firebase ID Token 검증. 검증 후 `c.Set("firebase_uid", uid)` 저장. 보호된 라우트에만 적용 (`/auth` 제외).
- **`internal/db/`** — `make sqlc`로 자동 생성 (직접 편집 금지). 쿼리 변경은 `db/queries/*.sql` 수정 후 재생성.

### DB 접근 패턴
`db/queries/*.sql` 작성 → `make sqlc` → `internal/db/` 에 타입세이프 Go 코드 생성 → `internal/repository/` 에서 래핑 → Service에서 사용.

### 위치 데이터
Dot의 `location` 컬럼은 `GEOMETRY(Point, 4326)`. 삽입 시 `ST_SetSRID(ST_MakePoint(lng, lat), 4326)`. 공간 쿼리는 `::geography` 캐스팅 후 `ST_DWithin`, `ST_Distance` 사용.

### 인증 플로우
소셜 로그인 토큰(카카오/Apple/Google) → `/auth/login` → 카카오 API 또는 JWT 검증 → Firebase Custom Token 생성 → 클라이언트가 Firebase ID Token 획득 → 이후 모든 요청 `Authorization: Bearer <id_token>` → `middleware/auth.go` 검증.

### 응답 형식
```go
// 성공
c.JSON(200, map[string]any{"data": ...})

// 에러
echo.NewHTTPError(404, map[string]any{"error": map[string]string{"code": "ROOM_NOT_FOUND", "message": "..."}})
```

## External Services

| 서비스 | 용도 | 설정 키 |
|--------|------|---------|
| Firebase Admin SDK | ID Token 검증, Custom Token 발급 | `FIREBASE_CREDENTIALS` (JSON string) |
| Kakao API | OAuth 토큰으로 사용자 정보 조회 | `KAKAO_REST_API_KEY` |
| Mapbox Geocoding | Dot 저장 시 place_name 자동 채움 | `MAPBOX_ACCESS_TOKEN` |
| Cloudflare R2 | 미디어 presigned URL (S3 호환) | `R2_*` |

`FIREBASE_CREDENTIALS` 미설정 시 Firebase 클라이언트는 nil로 초기화되고 경고 로그 출력 (로컬 개발 편의).

## Local DB

- postgis/postgis:17-3.5 (PostgreSQL 17 + PostGIS 3.5)
- `postgres://dottie:dottie_local@localhost:5432/dottie?sslmode=disable`

## Agent Toolkit (`.claude/`)

이 프로젝트는 `.claude/` 아래에 에이전트 워크플로우 도구를 포함합니다.

### Slash Commands (`.claude/commands/`)

| 커맨드 | 설명 |
|--------|------|
| `/spec` | 스펙 작성 — 코드 작성 전 구조화된 SPEC.md 생성 |
| `/plan` | 작업 분해 — SPEC.md 기반으로 tasks/plan.md + tasks/todo.md 생성 |
| `/build` | 증분 구현 — todo에서 다음 태스크를 TDD로 구현 후 커밋 |
| `/test` | TDD 워크플로우 — 실패 테스트 작성 → 구현 → 검증 (버그는 Prove-It 패턴) |
| `/review` | 5축 코드 리뷰 — correctness, readability, architecture, security, performance |
| `/ship` | 출시 체크리스트 — 3개 persona 병렬 fan-out 후 GO/NO-GO 결정 |
| `/code-simplify` | 동작 변경 없이 코드 단순화 |

### Agent Personas (`.claude/agents/`)

`/ship` 및 직접 호출로 사용 가능한 전문가 페르소나:

- **`code-reviewer`** — 5축 코드 리뷰 (merge 전 사용)
- **`security-auditor`** — 취약점 탐지, OWASP 기반 감사
- **`test-engineer`** — 테스트 전략, 커버리지 분석, Prove-It 패턴

페르소나 orchestration 규칙: **페르소나는 다른 페르소나를 호출하지 않는다.** 합성은 슬래시 커맨드 또는 사용자가 담당. 자세한 내용은 [`.claude/agents/README.md`](.claude/agents/README.md) 참고.

### Skills (`.claude/skills/`)

의도 → 스킬 매핑 (AGENTS.md 참고):

| 의도 | 스킬 |
|------|------|
| 새 기능 | `spec-driven-development` → `incremental-implementation` + `test-driven-development` |
| 계획/분해 | `planning-and-task-breakdown` |
| 버그/오류 | `debugging-and-error-recovery` |
| 코드 리뷰 | `code-review-and-quality` |
| 리팩토링 | `code-simplification` |
| API 설계 | `api-and-interface-design` |
| 보안 강화 | `security-and-hardening` |
| 출시 준비 | `shipping-and-launch` |

### References (`.claude/references/`)

- `security-checklist.md` — OWASP 기반 보안 체크리스트
- `testing-patterns.md` — 테스트 패턴 가이드
- `performance-checklist.md` — 성능 최적화 체크리스트
- `orchestration-patterns.md` — 에이전트 orchestration 패턴 카탈로그

## Boundaries

- **Always:** 새 엔드포인트는 Handler → Service → Repository 레이어 순서 준수. `internal/db/` 직접 편집 금지.
- **Always:** 쿼리 변경 후 `make sqlc` 실행.
- **Never:** 스킬이 적용되는 작업을 스킬 없이 바로 구현 (AGENTS.md의 Anti-Rationalization 참고).
