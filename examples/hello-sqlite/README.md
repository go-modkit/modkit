# hello-sqlite

Example consuming app for modkit using SQLite.

## Run

```bash
go run ./cmd/api
```

Then hit:

```bash
curl http://localhost:8080/health
```

## Run with Docker Compose

```bash
docker compose up -d --build
curl http://localhost:8080/health
docker compose down -v
```

## Configuration

Environment variables:
- `HTTP_ADDR` (default `:8080`)
- `SQLITE_PATH` (example `/tmp/app.db`)
- `SQLITE_CONNECT_TIMEOUT` (default `0`)
