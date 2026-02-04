# hello-mysql

Example consuming app for modkit using MySQL, sqlc, and migrations.

## What This Example Includes
- Modules: `AppModule`, `DatabaseModule`, `UsersModule`, `AuditModule` (consumes `UsersService` export).
- Endpoints:
  - `GET /health` → `{ "status": "ok" }`
  - `POST /users` → create user
  - `GET /users` → list users
  - `GET /users/{id}` → user payload
  - `PUT /users/{id}` → update user
  - `DELETE /users/{id}` → delete user
- Swagger UI at `GET /docs/index.html` (also available at `/swagger/index.html`)
- MySQL via docker-compose for local runs.
- Testcontainers for integration smoke tests.
- Migrations and sqlc-generated queries.
- JSON request logging via `log/slog`.
- Errors use RFC 7807 Problem Details (`application/problem+json`).

## Run (Docker Compose + Local Migrate)

```bash
make run
```

This starts MySQL in Docker, runs migrations locally, seeds data locally, and starts the app container.

Then hit:

```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{"name":"Ada","email":"ada@example.com"}'
curl http://localhost:8080/users
curl http://localhost:8080/users/1
curl -X PUT http://localhost:8080/users/1 -H 'Content-Type: application/json' -d '{"name":"Ada Lovelace","email":"ada@example.com"}'
curl -X DELETE http://localhost:8080/users/1
curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{"name":"Ada","email":"ada@example.com"}'
open http://localhost:8080/docs/index.html
```

The duplicate create request returns `409 Conflict` with `application/problem+json`.

You can seed separately with:

```bash
make seed
```

Swagger docs are checked in. To regenerate them, run:

```bash
make swagger
```

## Test

```bash
make test
```

## Compose Services
- `mysql` on `localhost:3306`
- `app` on `localhost:8080` (runs migrate + seed before starting)

The compose services build from `examples/hello-mysql/Dockerfile`.
`LOG_LEVEL` defaults to `info`, but compose sets it to `debug`.

## Configuration
Environment variables:
- `HTTP_ADDR` (default `:8080`)
- `MYSQL_DSN` (default `root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true`)
- `LOG_FORMAT` (`text` or `json`, default `text`)
- `LOG_LEVEL` (`debug`, `info`, `warn`, `error`, default `info`)
- `LOG_COLOR` (`auto`, `on`, `off`, default `auto`)
- `LOG_TIME` (`local`, `utc`, `none`, default `local`)
- `LOG_STYLE` (`pretty`, `plain`, `multiline`, default `pretty`, text only)
