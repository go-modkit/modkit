# hello-mysql

Example consuming app for modkit using MySQL, sqlc, and migrations.

## What This Example Includes
- Modules: `AppModule`, `DatabaseModule`, `UsersModule`, `AuditModule` (consumes `UsersService` export).
- Endpoints:
  - `GET /health` → `{ "status": "ok" }`
  - `GET /users/{id}` → user payload
- Swagger UI at `GET /swagger/index.html`
- MySQL via docker-compose for local runs.
- Testcontainers for integration smoke tests.
- Migrations and sqlc-generated queries.

## Run (Docker Compose + Local Migrate)

```bash
make run
```

This starts MySQL in Docker, runs migrations locally, and starts the app container.

Then hit:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/users/1
open http://localhost:8080/swagger/index.html
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
- `app` on `localhost:8080`
- `migrate` (profiled; not started by default)

## Configuration
Environment variables:
- `HTTP_ADDR` (default `:8080`)
- `MYSQL_DSN` (default `root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true`)
