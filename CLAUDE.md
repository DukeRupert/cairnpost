# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is CairnPost

A multi-tenant CRM for small service businesses (contractors, consultants, tradespeople). Part of the Firefly Software product suite. Go backend with htmx + templ for server-rendered UI and Svelte islands for high-interactivity components (pipeline kanban).

## Commands

```bash
make dev              # Live reload via Air (runs templ generate + go build)
make build            # Build binary to bin/cairnpost
make test             # Run all tests: go test ./...
go test ./internal/database/...  # Run tests for a single package
make db               # Start PostgreSQL via Docker Compose
make migrate-up       # Run migrations (requires DATABASE_URL env var and golang-migrate CLI)
make migrate-down     # Roll back one migration
make templ            # Generate templ files
make tailwind         # Build Tailwind CSS
make tailwind-watch   # Watch mode for Tailwind CSS
```

## Architecture

- **Entry point**: `cmd/server/main.go` — loads config, connects to DB, sets up `http.NewServeMux` routes
- **Config**: `internal/config/` — reads from env vars (`DATABASE_URL` required, `PORT` defaults 8080, `ENVIRONMENT` defaults development)
- **Database**: `internal/database/` — PostgreSQL via `sqlx` + `lib/pq`. Migrations in `internal/database/migrations/` using golang-migrate format
- **Models**: `internal/model/` — structs with `db:` and `json:` tags matching the DB schema. Uses `uuid.UUID` for IDs, `decimal.Decimal` for monetary values, `*T` for nullable fields
- **Frontend**: `web/` — templ templates, Tailwind CSS (`web/static/css/input.css`), generated `*_templ.go` files are gitignored

## Multi-tenancy

Every table (except `orgs`) has an `org_id` foreign key. All queries must be scoped by `org_id` to enforce tenant isolation. The schema enforces this with cascading deletes from `orgs`.

## Key conventions

- Standard library `net/http` router (no framework) — routes use Go 1.22+ method patterns like `"GET /health"`
- Air config (`.air.toml`) watches `.go`, `.templ`, `.sql` files and runs `templ generate` before building
- Deal stages are plain text strings (not enums), with defaults defined in `model.DefaultStages`
- Copy `.env.example` to `.env` for local development; Docker Compose provides its own env vars for the `app` service
