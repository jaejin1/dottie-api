# Dottie Backend Specification

## Go API Server — Project Setup & MVP Development Guide

---

## 1. 프로젝트 개요

### 1.1 역할

Dottie 모바일 앱(Flutter)의 백엔드 API 서버. 위치 데이터(dot) 저장·조회, 사용자 관리, 방(room) 관리, 미디어 업로드, 공유 지도 데이터 제공.

### 1.2 아키텍처 요약

```
[Flutter App] ←→ [Cloudflare CDN/R2] ←→ [Go API Server] ←→ [PostgreSQL]
                                              ↕
                                      [Firebase Admin SDK]
                                      [Mapbox Geocoding API]
```

### 1.3 개발 환경

- **1인 개발자** (풀스택)
- Claude agent로 개발 보조

---

## 2. 기술 스택

### 2.1 언어 & 프레임워크

- **Go 1.22+**
- **Echo v4** — HTTP 프레임워크 (경량, 미들웨어 풍부, 성능 우수)
  - 대안: Fiber, Gin — Echo 선택 이유: 미들웨어 체인, 구조적 에러 핸들링

### 2.2 데이터베이스

- **PostgreSQL 17+**
  - PostGIS 확장 — 위치 데이터 저장 & 공간 쿼리 (반경 내 dot 찾기, 만남 감지)
  - 타임스탬프 + 좌표 인덱싱
- **ORM/Query Builder:**
  - **sqlc** — SQL 파일에서 타입세이프 Go 코드 생성 (직관적, 성능 좋음)
  - 마이그레이션: **golang-migrate/migrate**

### 2.3 인증

- **Firebase Admin SDK (Go)**
  - `firebase.google.com/go/v4`
  - Firebase ID Token 검증 (Flutter에서 받은 토큰)
  - Custom Token 발급 (소셜 로그인용)
- **카카오 로그인 검증:**
  - 카카오 OAuth access token → 카카오 API로 사용자 정보 조회 → Firebase Custom Token 생성

### 2.4 인프라 & 호스팅

#### 옵션 A: Cloudflare 중심 (권장 — 1인 개발자에 적합)

```
Cloudflare Workers/Pages  — 정적 에셋, CDN
Cloudflare R2             — 이미지/미디어 스토리지 (S3 호환, 이그레스 무료!)
Fly.io 또는 Railway       — Go 서버 호스팅 (Docker 기반)
Neon 또는 Supabase DB     — Managed PostgreSQL (PostGIS 지원)
```

#### 옵션 B: AWS 중심

```
AWS ECS Fargate 또는 EC2  — Go 서버
AWS RDS PostgreSQL        — DB (PostGIS 활성화)
AWS S3 + CloudFront       — 미디어 + CDN
```

#### 권장 조합 (비용·복잡도 최적화)

```
Go 서버:     Fly.io (무료 티어: 3 shared VMs, 충분)
DB:         Neon PostgreSQL (무료 티어: 0.5GB, MVP 충분) + PostGIS
스토리지:    Cloudflare R2 (무료 티어: 10GB, 이그레스 무료)
CDN:        Cloudflare (무료)
DNS:        Cloudflare
도메인:      dottie.app (구매)
```

이유:

- Cloudflare R2는 이그레스 비용 0원 → 사진 서빙 비용 걱정 없음
- Fly.io는 Go 바이너리 Docker 배포 간단
- Neon은 서버리스 PostgreSQL, PostGIS 지원, 자동 스케일
- 전체 MVP 운영비: 월 $0~$10

### 2.5 기타 의존성

- **go-redis** — 세션 캐싱, 초대 코드 TTL (선택, MVP에서는 PostgreSQL로 충분)
- **zap** (uber-go/zap) — 구조화된 로깅
- **viper** — 설정 관리 (환경변수, config 파일)
- **validator/v10** — 요청 바디 검증
- **uuid** (google/uuid) — ID 생성
- **aws-sdk-go-v2** (S3 호환) — R2 presigned URL 생성
- **mapbox geocoding** — 역지오코딩 HTTP 호출 (별도 SDK 불필요)

---

## 3. 프로젝트 구조

```
dottie-server/
├── cmd/
│   └── server/
│       └── main.go              # 진입점
├── internal/
│   ├── config/
│   │   └── config.go            # Viper 설정 로드
│   ├── server/
│   │   └── server.go            # Echo 서버 초기화, 미들웨어, 라우트 등록
│   ├── middleware/
│   │   ├── auth.go              # Firebase ID Token 검증 미들웨어
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── ratelimit.go
│   ├── handler/                 # HTTP 핸들러 (Controller 역할)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── dot_handler.go
│   │   ├── daylog_handler.go
│   │   ├── room_handler.go
│   │   ├── shared_map_handler.go
│   │   └── media_handler.go
│   ├── service/                 # 비즈니스 로직
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── dot_service.go
│   │   ├── daylog_service.go
│   │   ├── room_service.go
│   │   ├── shared_map_service.go
│   │   ├── media_service.go
│   │   └── geocoding_service.go # Mapbox 역지오코딩
│   ├── repository/              # DB 접근 (sqlc 생성 코드 래핑)
│   │   ├── user_repo.go
│   │   ├── dot_repo.go
│   │   ├── daylog_repo.go
│   │   ├── room_repo.go
│   │   └── shared_repo.go
│   ├── model/                   # 도메인 모델 & DTO
│   │   ├── user.go
│   │   ├── dot.go
│   │   ├── daylog.go
│   │   ├── room.go
│   │   ├── character.go
│   │   └── dto/
│   │       ├── request.go       # API 요청 DTO
│   │       └── response.go      # API 응답 DTO
│   └── pkg/
│       ├── firebase/
│       │   └── client.go        # Firebase Admin SDK 초기화
│       ├── storage/
│       │   └── r2_client.go     # Cloudflare R2 (S3 호환) 클라이언트
│       ├── mapbox/
│       │   └── geocoding.go     # Mapbox Geocoding API 호출
│       └── errors/
│           └── errors.go        # 커스텀 에러 타입
├── db/
│   ├── migrations/
│   │   ├── 000001_init.up.sql
│   │   ├── 000001_init.down.sql
│   │   └── ...
│   ├── queries/                 # sqlc 쿼리 파일
│   │   ├── users.sql
│   │   ├── dots.sql
│   │   ├── daylogs.sql
│   │   ├── rooms.sql
│   │   └── shared.sql
│   └── sqlc.yaml                # sqlc 설정
├── Dockerfile
├── docker-compose.yml           # 로컬 개발 (PostgreSQL + PostGIS)
├── fly.toml                     # Fly.io 배포 설정
├── Makefile
├── go.mod
├── go.sum
└── .env.example
```

---

## 4. 데이터베이스 스키마

### 4.1 초기 마이그레이션 (000001_init.up.sql)

```sql
-- PostGIS 확장 활성화
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- 사용자
-- ========================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    firebase_uid    VARCHAR(128) UNIQUE NOT NULL,
    provider        VARCHAR(20) NOT NULL,           -- 'kakao', 'apple', 'google'
    nickname        VARCHAR(30) NOT NULL,
    profile_image   TEXT,

    -- 캐릭터 설정 (JSON으로 유연하게)
    character_config JSONB NOT NULL DEFAULT '{"color":"blue","accessory":"none","expression":"default"}',

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);

-- ========================================
-- 하루 기록 (세션)
-- ========================================
CREATE TABLE day_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL,
    ended_at        TIMESTAMPTZ,
    is_recording    BOOLEAN NOT NULL DEFAULT true,

    -- 통계 (기록 종료 시 계산)
    total_distance_m    DOUBLE PRECISION,
    place_count         INTEGER,
    total_duration_sec  INTEGER,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, date)  -- 하루에 하나의 기록만
);

CREATE INDEX idx_daylogs_user_date ON day_logs(user_id, date DESC);

-- ========================================
-- Dot (위치 점)
-- ========================================
CREATE TABLE dots (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    day_log_id      UUID NOT NULL REFERENCES day_logs(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- PostGIS geometry
    location        GEOMETRY(Point, 4326) NOT NULL,

    -- 편의를 위한 개별 컬럼 (쿼리 편의)
    latitude        DOUBLE PRECISION NOT NULL,
    longitude       DOUBLE PRECISION NOT NULL,

    timestamp       TIMESTAMPTZ NOT NULL,
    place_name      VARCHAR(200),
    place_category  VARCHAR(50),         -- 'cafe', 'restaurant', 'park', 'home', 'office' 등
    photo_url       TEXT,
    memo            VARCHAR(500),
    emotion         VARCHAR(20),         -- 'happy', 'tired', 'excited', 'hungry' 등

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dots_daylog ON dots(day_log_id, timestamp ASC);
CREATE INDEX idx_dots_user_time ON dots(user_id, timestamp DESC);
CREATE INDEX idx_dots_location ON dots USING GIST(location);  -- 공간 인덱스

-- ========================================
-- 방 (Room)
-- ========================================
CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(50) NOT NULL,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_code     VARCHAR(20) UNIQUE,
    invite_expires  TIMESTAMPTZ,
    max_members     INTEGER NOT NULL DEFAULT 4,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========================================
-- 방 멤버
-- ========================================
CREATE TABLE room_members (
    room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX idx_room_members_user ON room_members(user_id);

-- ========================================
-- 공유된 기록 (방에 공유된 daylog)
-- ========================================
CREATE TABLE shared_day_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    day_log_id      UUID NOT NULL REFERENCES day_logs(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(room_id, day_log_id)  -- 같은 기록 중복 공유 방지
);

CREATE INDEX idx_shared_room_date ON shared_day_logs(room_id, shared_at DESC);

-- ========================================
-- updated_at 자동 갱신 트리거
-- ========================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_daylogs_updated_at
    BEFORE UPDATE ON day_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_rooms_updated_at
    BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

### 4.2 공간 쿼리 예시 (만남 감지)

```sql
-- 같은 시간대(±30분)에 반경 100m 이내에 있는 dot 쌍 찾기
SELECT
    a.id AS dot_a_id,
    b.id AS dot_b_id,
    a.user_id AS user_a,
    b.user_id AS user_b,
    a.timestamp,
    ST_Distance(a.location::geography, b.location::geography) AS distance_m
FROM dots a
JOIN dots b ON a.user_id != b.user_id
    AND a.day_log_id IN (
        SELECT day_log_id FROM shared_day_logs WHERE room_id = $1
    )
    AND b.day_log_id IN (
        SELECT day_log_id FROM shared_day_logs WHERE room_id = $1
    )
    AND ABS(EXTRACT(EPOCH FROM (a.timestamp - b.timestamp))) < 1800  -- 30분
    AND ST_DWithin(a.location::geography, b.location::geography, 100)  -- 100m
ORDER BY a.timestamp;
```

---

## 5. API 상세 명세

### 5.1 공통 규약

**Base URL:** `https://api.dottie.app/v1`

**응답 형식:**

```json
// 성공
{
  "data": { ... },
  "meta": { "page": 1, "total": 50 }
}

// 에러
{
  "error": {
    "code": "ROOM_NOT_FOUND",
    "message": "방을 찾을 수 없습니다"
  }
}
```

**인증:** 모든 요청 (auth 제외)에 `Authorization: Bearer <firebase_id_token>` 필수

**페이지네이션:** cursor 기반 (`?cursor=<last_id>&limit=20`)

### 5.2 엔드포인트 상세

#### POST /auth/login

소셜 로그인 토큰 검증 → Firebase Custom Token 반환

```json
// Request
{
  "provider": "kakao",       // "kakao" | "apple" | "google"
  "token": "oauth_access_token_here",
  "nickname": "도티유저"       // 최초 가입 시에만
}

// Response
{
  "data": {
    "firebase_custom_token": "eyJhbGci...",
    "user": {
      "id": "uuid",
      "nickname": "도티유저",
      "character_config": { "color": "blue", "accessory": "none", "expression": "default" },
      "is_new": true
    }
  }
}
```

**서버 내부 로직:**

1. provider에 따라 토큰 검증
   - kakao: `GET https://kapi.kakao.com/v2/user/me` (Authorization: Bearer <token>)
   - apple: JWT 검증 (Apple public key)
   - google: Firebase Admin SDK `VerifyIDToken`
2. users 테이블에서 조회 or 생성
3. Firebase Admin SDK `CustomToken` 생성
4. 응답

#### POST /dots

Dot 하나 저장

```json
// Request
{
  "day_log_id": "uuid",
  "latitude": 37.5665,
  "longitude": 126.9780,
  "timestamp": "2026-04-27T14:00:00+09:00",
  "place_name": "광화문 스타벅스",     // nullable
  "place_category": "cafe",           // nullable
  "memo": "아메리카노 한 잔",          // nullable
  "emotion": "happy"                  // nullable
}

// Response
{
  "data": {
    "id": "uuid",
    "photo_upload_url": null
  }
}
```

**서버 내부 로직:**

1. Firebase ID Token에서 user_id 추출
2. day_log_id 소유자 확인
3. PostGIS Point 생성: `ST_SetSRID(ST_MakePoint(lng, lat), 4326)`
4. place_name이 null이면 Mapbox Reverse Geocoding으로 자동 채움
5. DB 저장

#### POST /dots/batch

오프라인 동기화 — 여러 dot 일괄 업로드

```json
// Request
{
  "day_log_id": "uuid",
  "dots": [
    {
      "client_id": "local-uuid-1",
      "latitude": 37.5665,
      "longitude": 126.9780,
      "timestamp": "2026-04-27T14:00:00+09:00",
      "memo": "..."
    },
    // ... 최대 50개
  ]
}

// Response
{
  "data": {
    "synced": [
      { "client_id": "local-uuid-1", "server_id": "uuid-abc" }
    ],
    "failed": []
  }
}
```

#### GET /rooms/:id/shared-map?date=2026-04-27

합본 지도 데이터 — 해당 날짜에 이 방에 공유된 모든 멤버의 dot

```json
// Response
{
  "data": {
    "date": "2026-04-27",
    "room_id": "uuid",
    "members": [
      {
        "user_id": "uuid-a",
        "nickname": "민지",
        "character_config": { "color": "coral", "accessory": "hat", "expression": "default" },
        "dots": [
          {
            "id": "uuid",
            "latitude": 37.5665,
            "longitude": 126.9780,
            "timestamp": "2026-04-27T09:00:00+09:00",
            "place_name": "집",
            "place_category": "home",
            "photo_url": null,
            "memo": null,
            "emotion": "tired"
          },
          // ...
        ],
        "stats": {
          "total_distance_m": 12500,
          "place_count": 8,
          "total_duration_sec": 43200
        }
      },
      {
        "user_id": "uuid-b",
        "nickname": "준호",
        "character_config": { "color": "blue", "accessory": "none", "expression": "happy" },
        "dots": [ ... ]
      }
    ],
    "encounters": [
      {
        "timestamp": "2026-04-27T18:30:00+09:00",
        "user_ids": ["uuid-a", "uuid-b"],
        "location": { "latitude": 37.5172, "longitude": 127.0473 },
        "place_name": "코엑스",
        "distance_m": 45
      }
    ]
  }
}
```

#### POST /media/upload

Presigned URL 발급 (Cloudflare R2)

```json
// Request
{
  "content_type": "image/jpeg",
  "file_size": 2048576
}

// Response
{
  "data": {
    "upload_url": "https://r2.dottie.app/...",   // presigned PUT URL
    "public_url": "https://media.dottie.app/...", // CDN URL (업로드 후 접근)
    "expires_in": 3600
  }
}
```

**서버 내부 로직:**

1. 파일 크기 검증 (최대 10MB)
2. content_type 검증 (image/jpeg, image/png, image/heic)
3. R2 presigned URL 생성 (aws-sdk-go-v2, S3 호환)
4. 파일 경로: `users/<user_id>/dots/<date>/<uuid>.jpg`

---

## 6. 인증 플로우 상세

### 6.1 카카오 로그인 (가장 중요)

```
[Flutter]                    [Go Server]                 [Kakao API]          [Firebase]
    |                            |                           |                    |
    |-- 카카오 로그인 SDK -------->|                           |                    |
    |   (access_token 획득)      |                           |                    |
    |                            |                           |                    |
    |-- POST /auth/login ------->|                           |                    |
    |   {provider:"kakao",       |                           |                    |
    |    token:"kakao_token"}    |                           |                    |
    |                            |-- GET /v2/user/me ------->|                    |
    |                            |   (Bearer kakao_token)    |                    |
    |                            |<-- {id, nickname, email} -|                    |
    |                            |                           |                    |
    |                            |-- users 테이블 조회/생성 --|                    |
    |                            |                           |                    |
    |                            |-- CustomToken(uid) -------|------------------>|
    |                            |<-- firebase_custom_token --|<-----------------|
    |                            |                           |                    |
    |<-- {firebase_custom_token, |                           |                    |
    |     user}                  |                           |                    |
    |                            |                           |                    |
    |-- signInWithCustomToken -->|                           |                    |
    |   (Firebase Auth)          |                           |                    |
    |<-- firebase_id_token ------|                           |                    |
    |                            |                           |                    |
    |== 이후 모든 API 요청 ==    |                           |                    |
    |-- Authorization: Bearer    |                           |                    |
    |   <firebase_id_token>      |                           |                    |
```

### 6.2 Firebase Admin SDK 초기화 (Go)

```go
package firebase

import (
    "context"
    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/auth"
    "google.golang.org/api/option"
)

type Client struct {
    auth *auth.Client
}

func NewClient(credentialsJSON []byte) (*Client, error) {
    app, err := firebase.NewApp(context.Background(), nil,
        option.WithCredentialsJSON(credentialsJSON))
    if err != nil {
        return nil, err
    }
    authClient, err := app.Auth(context.Background())
    if err != nil {
        return nil, err
    }
    return &Client{auth: authClient}, nil
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
    return c.auth.VerifyIDToken(ctx, idToken)
}

func (c *Client) CreateCustomToken(ctx context.Context, uid string) (string, error) {
    return c.auth.CustomToken(ctx, uid)
}
```

### 6.3 Auth 미들웨어

```go
func AuthMiddleware(fbClient *firebase.Client) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                return echo.NewHTTPError(401, "missing auth token")
            }
            idToken := strings.TrimPrefix(authHeader, "Bearer ")

            token, err := fbClient.VerifyIDToken(c.Request().Context(), idToken)
            if err != nil {
                return echo.NewHTTPError(401, "invalid auth token")
            }

            // context에 user ID 저장
            c.Set("firebase_uid", token.UID)
            return next(c)
        }
    }
}
```

---

## 7. 역지오코딩 (Mapbox)

### 7.1 Mapbox Geocoding API 호출

```go
package mapbox

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type GeocodingClient struct {
    accessToken string
    httpClient  *http.Client
}

type GeocodingResult struct {
    PlaceName string
    Category  string
}

func (c *GeocodingClient) ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodingResult, error) {
    url := fmt.Sprintf(
        "https://api.mapbox.com/geocoding/v5/mapbox.places/%f,%f.json?access_token=%s&language=ko&types=poi,address",
        lng, lat, c.accessToken,
    )

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Features []struct {
            PlaceName  string   `json:"place_name"`
            Properties struct {
                Category string `json:"category"`
            } `json:"properties"`
        } `json:"features"`
    }

    json.NewDecoder(resp.Body).Decode(&result)

    if len(result.Features) > 0 {
        return &GeocodingResult{
            PlaceName: result.Features[0].PlaceName,
            Category:  result.Features[0].Properties.Category,
        }, nil
    }
    return nil, nil
}
```

### 7.2 사용 시점

- Dot 저장 시 place_name이 null이면 서버에서 자동 호출
- 무료 티어: 월 100,000 요청 → 1인당 하루 20 dot, 5,000명까지 무료

---

## 8. 미디어 스토리지 (Cloudflare R2)

### 8.1 R2 설정

```go
package storage

import (
    "context"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
    client     *s3.Client
    presigner  *s3.PresignClient
    bucket     string
    publicURL  string  // https://media.dottie.app
}

func NewR2Client(accountID, accessKeyID, secretAccessKey, bucket, publicURL string) *R2Client {
    cfg := aws.Config{
        Region: "auto",
        Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
    }

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.BaseEndpoint = aws.String(
            fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
        )
    })

    return &R2Client{
        client:    client,
        presigner: s3.NewPresignClient(client),
        bucket:    bucket,
        publicURL: publicURL,
    }
}

func (r *R2Client) GenerateUploadURL(ctx context.Context, key, contentType string) (string, error) {
    presignResult, err := r.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(r.bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, s3.WithPresignExpires(1*time.Hour))
    if err != nil {
        return "", err
    }
    return presignResult.URL, nil
}

func (r *R2Client) GetPublicURL(key string) string {
    return fmt.Sprintf("%s/%s", r.publicURL, key)
}
```

### 8.2 Cloudflare R2 무료 티어

- 저장: 10GB 무료
- 읽기 요청: 월 1,000만 무료
- 쓰기 요청: 월 100만 무료
- **이그레스(데이터 전송): 완전 무료** ← 이게 핵심, AWS S3 대비 큰 장점
- Cloudflare 커스텀 도메인 연결 가능 (media.dottie.app)

---

## 9. 통계 계산

### 9.1 DayLog 종료 시 통계 계산

```sql
-- daylog 종료 시 호출
-- 이동 거리 계산 (순차적 dot 간 거리 합산)
WITH ordered_dots AS (
    SELECT
        location,
        timestamp,
        LAG(location) OVER (ORDER BY timestamp) AS prev_location
    FROM dots
    WHERE day_log_id = $1
    ORDER BY timestamp
)
SELECT
    COALESCE(SUM(
        ST_Distance(location::geography, prev_location::geography)
    ), 0) AS total_distance_m
FROM ordered_dots
WHERE prev_location IS NOT NULL;

-- 방문 장소 수 (고유 place_name 기준)
SELECT COUNT(DISTINCT place_name)
FROM dots
WHERE day_log_id = $1 AND place_name IS NOT NULL;

-- 총 기록 시간
SELECT EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp)))::integer AS duration_sec
FROM dots
WHERE day_log_id = $1;
```

---

## 10. Docker & 배포

### 10.1 Dockerfile

```dockerfile
# Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

### 10.2 docker-compose.yml (로컬 개발)

```yaml
version: "3.8"
services:
  db:
    image: postgis/postgis:16-3.4
    environment:
      POSTGRES_DB: dottie
      POSTGRES_USER: dottie
      POSTGRES_PASSWORD: dottie_local
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_URL=postgres://dottie:dottie_local@db:5432/dottie?sslmode=disable
      - PORT=8080
    depends_on:
      - db

volumes:
  pgdata:
```

### 10.3 fly.toml (Fly.io 배포)

```toml
app = "dottie-api"
primary_region = "nrt"  # 도쿄 (한국에서 가장 가까운 Fly.io 리전)

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 256
```

### 10.4 배포 명령어

```bash
# Fly.io 초기 설정
fly launch --name dottie-api --region nrt

# 시크릿 설정
fly secrets set DB_URL="postgres://..." \
  FIREBASE_CREDENTIALS='{"type":"service_account",...}' \
  MAPBOX_ACCESS_TOKEN="pk...." \
  R2_ACCOUNT_ID="..." \
  R2_ACCESS_KEY_ID="..." \
  R2_SECRET_ACCESS_KEY="..." \
  R2_BUCKET_NAME="dottie-media"

# 배포
fly deploy
```

---

## 11. Makefile

```makefile
.PHONY: dev build test migrate sqlc

# 로컬 개발
dev:
	docker-compose up -d db
	go run ./cmd/server

# 빌드
build:
	go build -o bin/server ./cmd/server

# 테스트
test:
	go test ./... -v

# 마이그레이션
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

# sqlc 코드 생성
sqlc:
	sqlc generate

# Docker
docker-build:
	docker build -t dottie-server .

# Fly.io 배포
deploy:
	fly deploy
```

---

## 12. MVP 개발 우선순위

### Phase 1 (1주) — 프로젝트 초기화 & 인프라

- [ ] Go 프로젝트 생성 & 폴더 구조
- [ ] Echo 서버 세팅 (미들웨어, CORS, 로깅)
- [ ] PostgreSQL + PostGIS docker-compose
- [ ] 초기 마이그레이션 (전체 스키마)
- [ ] sqlc 설정 & 기본 쿼리
- [ ] Firebase Admin SDK 연동
- [ ] 환경변수 관리 (viper)
- [ ] 헬스체크 엔드포인트

### Phase 2 (1주) — 인증 & 사용자

- [ ] POST /auth/login (카카오, Apple, Google)
- [ ] Auth 미들웨어
- [ ] GET/PUT /users/me
- [ ] PUT /users/me/character
- [ ] 카카오 OAuth 토큰 검증 구현
- [ ] Apple JWT 검증 구현

### Phase 3 (1.5주) — 기록 핵심

- [ ] POST /recordings/start
- [ ] POST /recordings/end (통계 계산 포함)
- [ ] POST /dots (단일)
- [ ] POST /dots/batch (일괄)
- [ ] GET /dots?date=
- [ ] GET /daylogs
- [ ] GET /daylogs/:id
- [ ] Mapbox 역지오코딩 연동
- [ ] Cloudflare R2 presigned URL (POST /media/upload)

### Phase 4 (1.5주) — 소셜

- [ ] POST /rooms (생성)
- [ ] GET /rooms (목록)
- [ ] GET /rooms/:id (상세)
- [ ] POST /rooms/:id/invite (초대 코드)
- [ ] POST /rooms/join
- [ ] POST /rooms/:id/share (기록 공유)
- [ ] GET /rooms/:id/shared-map?date= (합본 데이터)
- [ ] 만남 감지 쿼리

### Phase 5 (0.5주) — 배포 & 안정화

- [ ] Fly.io 배포
- [ ] Neon PostgreSQL 연결
- [ ] Cloudflare R2 프로덕션 설정
- [ ] 레이트 리밋
- [ ] 에러 핸들링 표준화
- [ ] 기본 모니터링 (Fly.io 메트릭)

---

## 13. 환경변수 목록

```env
# 서버
PORT=8080
ENV=development  # development | production

# 데이터베이스
DB_URL=postgres://user:pass@host:5432/dottie?sslmode=require

# Firebase
FIREBASE_CREDENTIALS={"type":"service_account",...}  # JSON string

# 카카오 (서버에서 토큰 검증용)
KAKAO_REST_API_KEY=your_kakao_rest_api_key

# Mapbox
MAPBOX_ACCESS_TOKEN=pk.xxx

# Cloudflare R2
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=dottie-media
R2_PUBLIC_URL=https://media.dottie.app

# 기타
CORS_ORIGINS=https://dottie.app
```

---

## 14. 보안 고려사항

### 14.1 API 보안

- 모든 엔드포인트 Firebase ID Token 검증 (auth 제외)
- Rate limiting: IP당 100 req/min, 유저당 300 req/min
- 요청 바디 크기 제한: 10MB
- SQL injection: sqlc 사용으로 파라미터 바인딩 자동

### 14.2 데이터 보안

- HTTPS only (Fly.io 자동 TLS)
- DB 연결 SSL 필수
- 위치 데이터 암호화는 v2에서 고려 (at-rest encryption은 Neon 기본 제공)
- 사진 URL은 랜덤 UUID 경로 (추측 불가)

### 14.3 한국 위치정보법

- 위치정보 수집 동의 API 필요 (v2)
- 데이터 보유 기간 설정 & 자동 삭제 배치 (v2)
- 개인정보처리방침 페이지 필요

---

## 15. 성능 고려사항

### 15.1 쿼리 최적화

- dots 테이블: `(day_log_id, timestamp)` 복합 인덱스 → 하루치 dot 조회 O(log n)
- PostGIS GIST 인덱스: 공간 쿼리 (만남 감지) 성능 보장
- 합본 지도: 한 날짜 당 최대 4명 × 20 dot = 80행 → 별도 최적화 불필요

### 15.2 캐싱

- MVP에서는 캐싱 불필요 (데이터 규모 작음)
- v2: Redis 도입 시 합본 지도 결과 캐싱 (날짜+방 키)

### 15.3 동기화

- 오프라인 dot은 batch API로 일괄 업로드
- 충돌 해결: client_id 기반 idempotent — 동일 client_id 재전송 시 무시

---

_이 문서는 Claude agent가 Dottie Go 백엔드 서버를 초기 설정하고 MVP를 개발하는 데 필요한 모든 컨텍스트를 담고 있습니다. 각 Phase별로 작업 시 이 문서를 참조하세요._

