run:
	go run ./cmd/api

migrate-up:
	migrate -path internal/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/migrations -database "$$DATABASE_URL" down 1

test:
	go test ./...

smoke:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/smoke.ps1
