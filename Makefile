BINARY_HTTP := bin/user-http
BINARY_GRPC := bin/user-grpc

GO ?= go

.PHONY: run-http run-grpc build clean test migrate-up migrate-down coverage

run-http:
	DB_DSN?=postgres://user:pass@localhost:5432/usersvc?sslmode=disable
	REDIS_ADDR?=localhost:6379
	$(GO) run ./cmd/user-http

run-grpc:
	$(GO) run ./cmd/user-grpc

build:
	mkdir -p bin
	$(GO) build -o $(BINARY_HTTP) ./cmd/user-http
	$(GO) build -o $(BINARY_GRPC) ./cmd/user-grpc

clean:
	rm -rf bin

migrate-up:
	migrate -path internal/db/migrations -database "$$DB_DSN" up

migrate-down:
	migrate -path internal/db/migrations -database "$$DB_DSN" down 1

test: coverage

coverage:
        ./scripts/coverage.sh 95
