.PHONY: all up down logs server worker

# Default target
all: up

# Docker Compose ile tüm servisleri ayağa kaldır
up:
	docker-compose up --build -d

# Docker Compose ile tüm servisleri durdur
down:
	docker-compose down

# Docker container loglarını takip et
logs:
	docker-compose logs -f

# Server'ı lokal olarak çalıştırmak için
server:
	go run cmd/server/main.go

# Worker'ı lokal olarak çalıştırmak için
worker:
	go run cmd/worker/main.go
