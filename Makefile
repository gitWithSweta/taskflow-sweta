.PHONY: up down backend frontend

up:
	docker compose up -d

down:
	docker compose down

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev
