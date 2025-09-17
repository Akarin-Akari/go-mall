# Repository Guidelines

## Project Structure & Module Organization
Runtime starts in `cmd/server/main.go`; domain logic is grouped in `internal/{handler,service,repository,model,middleware,utils}`. Reusable libraries—cache, auth, product, upload, file handling—live in `pkg/`. Configuration templates sit in `configs/`, and database assets in `migrations/`, `init_database.sql`, and the `scripts/` helpers. Long-form design notes are kept in `docs/`. Integration, performance, and scenario suites live in `tests/`, while package-focused samples and fixtures reside beside their source in `pkg/**/_test.go` and the root testing utilities.

## Build, Test, and Development Commands
Run `go mod tidy` whenever dependencies change. Start the API with `go run cmd/server/main.go` or build a binary through `go build -o mall-go cmd/server/main.go`. `scripts/run.sh` combines environment checks with that run command for repeatable local boots. Execute all automated checks via `go test ./...`; narrow to modules such as `go test ./pkg/cache -run TestWarmupManager` or to the integration layer with `go test ./tests/integration`.

## Coding Style & Naming Conventions
Always apply `gofmt` (e.g., `go fmt ./...`) to enforce Go's tabbed indentation. Keep package names lowercase and succinct; exported types and functions use PascalCase that reflects their business role (`InitKeyManager`, `OrderService`). Co-locate HTTP request DTOs with handlers and reuse helpers from `internal/utils`. Stick to structured logging provided by `pkg/logger` and use `internal/config` wrappers for configuration access.

## Testing Guidelines
Mirror the existing `*_test.go` layout and follow table-driven patterns where practical. Unit suites cover caches and services, while integration and benchmarking flows live under `tests/`. When contributing new tests, prefer deterministic fixtures from `tests/helpers` and note any external requirements (MySQL, Redis). Validate coverage with `go test ./... -cover`, and flag lengthy performance runs so teammates can opt into `go test ./tests/performance -run OrderPerformance` only.

## Commit & Pull Request Guidelines
History shows Conventional Commit prefixes (`feat:`, `docs:`) often paired with short emoji or bilingual summaries; continue that style and keep the subject within 72 characters. Reference touched modules or SQL assets in the body when relevant. Pull requests should outline functional scope, config prerequisites, and include verification evidence—typically the `go test ./...` output or screenshots of Swagger checks for new endpoints.

## Environment & Configuration Tips
Copy `configs/config.yaml` and adjust database credentials before running the server. Use `scripts/init_permissions.go` or `create_database.go` when bootstrapping schemas, and keep migration SQL aligned with GORM models. Lightweight experiments can rely on the in-memory SQLite demos in `simple_test.go`, while production-like tests expect MySQL 8 plus Redis configured through environment variables or your shell profile.
