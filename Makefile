.PHONY: dev build test migrate-up migrate-down migrate-create sqlc docker-build deploy

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

# DB 로컬 접속
db-shell:
	docker-compose exec db psql -U dottie -d dottie

# 로컬 마이그레이션 (docker-compose DB 사용)
migrate-local:
	migrate -path db/migrations \
		-database "postgres://dottie:dottie_local@localhost:5432/dottie?sslmode=disable" up
