# hello-postgres

Example consuming app for modkit using Postgres.

## Run

```bash
go run ./cmd/api
```

Then hit:

```bash
curl http://localhost:8080/api/v1/health
```

## Run with Docker Compose

```bash
docker compose up -d --build
curl http://localhost:8080/api/v1/health
docker compose down -v
```

## Configuration

Environment variables:
- `HTTP_ADDR` (default `:8080`)
- `POSTGRES_DSN` (example `postgres://postgres:password@localhost:5432/app?sslmode=disable`)
