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
