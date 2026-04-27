DROP TRIGGER IF EXISTS trg_rooms_updated_at ON rooms;
DROP TRIGGER IF EXISTS trg_daylogs_updated_at ON day_logs;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS shared_day_logs;
DROP TABLE IF EXISTS room_members;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS dots;
DROP TABLE IF EXISTS day_logs;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS postgis;
