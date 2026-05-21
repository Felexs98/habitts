.PHONY: run build test swagger docker-up docker-down tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v

test-integration:
	go test -tags=integration -v -count=1 ./test/integration/...

tidy:
	go mod tidy

swagger:
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
