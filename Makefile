.PHONY: all build run test clean frontend backend install-deps migrate

all: build

install-deps:
	go mod download
	cd frontend && npm install

build: backend frontend

backend:
	go build -o bin/waf cmd/waf/main.go

frontend:
	cd frontend && npm run build

run:
	go run cmd/waf/main.go

dev:
	go run cmd/waf/main.go --config config.local.yaml

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf frontend/build/
	rm -rf frontend/dist/

migrate:
	psql -U postgres -d docode_waf -f migrations/init.sql

docker-build:
	docker build -t docode-waf:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down
