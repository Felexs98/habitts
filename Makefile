.PHONY: run build swagger docker-up docker-down tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

swagger:
	swag init -g cmd/api/main.go -o docs

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
