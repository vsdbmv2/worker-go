# makefile
.PHONY: build run test

build:
	docker build -t vsdbm-worker .

run:
	docker run --env-file .env vsdbm-worker

test:
	go test ./src/...

# .env
maxConcurrency=4
websocketHost=http://vsdbm-api.hfabio.dev

# go.mod
module vsdbm-worker

go 1.23

require (
	github.com/gorilla/websocket v1.5.0
	github.com/joho/godotenv v1.5.1
)