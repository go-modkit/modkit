# hello-mysql

Example consuming app for modkit using MySQL, sqlc, and migrations.

## What This Example Includes
- Modules: `AppModule`, `DatabaseModule`, `UsersModule`, `AuditModule` (consumes `UsersService` export).
- Endpoints:
  - `GET /health` → `{ "status": "ok" }`
  - `POST /auth/login` → demo JWT token
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

## Auth
- Demo login endpoint: `POST /auth/login` returns a JWT.
- Protected routes (require `Authorization: Bearer <token>`):
  - `POST /users`
  - `GET /users/{id}`
  - `PUT /users/{id}`
  - `DELETE /users/{id}`
- Public route:
  - `GET /users`

## Run (Docker Compose + Local Migrate)

```bash
make run
```

This starts MySQL in Docker, runs migrations locally, seeds data locally, and starts the app container.

Then hit:

```bash
curl http://localhost:8080/health

# Login to get a token (demo credentials). Requires `jq` for parsing.
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","password":"demo"}' | jq -r '.token')

# Public route
curl http://localhost:8080/users

# Protected routes (require Authorization header)
curl -X POST http://localhost:8080/users \
  -H 'Authorization: Bearer '"$TOKEN"'' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada","email":"ada@example.com"}'
curl -H 'Authorization: Bearer '"$TOKEN"'' http://localhost:8080/users/1
curl -X PUT http://localhost:8080/users/1 \
  -H 'Authorization: Bearer '"$TOKEN"'' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","email":"ada@example.com"}'
curl -X DELETE http://localhost:8080/users/1 -H 'Authorization: Bearer '"$TOKEN"''

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

## Validation

Request validation returns RFC 7807 Problem Details with `invalidParams` for field-level errors.

Create with missing fields:

```bash
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"","email":""}'
```

Example response:

```json
{
  "type": "https://httpstatuses.com/400",
  "title": "Bad Request",
  "status": 400,
  "detail": "validation failed",
  "instance": "/users",
  "invalidParams": [
    { "name": "name", "reason": "is required" },
    { "name": "email", "reason": "is required" }
  ]
}
```

Query parameter validation (pagination):

```bash
curl "http://localhost:8080/users?page=-1&limit=0"
```

Example response:

```json
{
  "type": "https://httpstatuses.com/400",
  "title": "Bad Request",
  "status": 400,
  "detail": "validation failed",
  "instance": "/users",
  "invalidParams": [
    { "name": "page", "reason": "must be >= 1" },
    { "name": "limit", "reason": "must be >= 1" }
  ]
}
```

## Lifecycle and Cleanup

Cleanup hooks are registered on providers via `ProviderDef.Cleanup`. The database module uses this hook to close the `*sql.DB` pool.

On shutdown, the API server:
- Stops accepting new requests and waits for in-flight requests to finish.
- Runs cleanup hooks in **LIFO** order (last registered, first cleaned).

The users service includes a context cancellation example via `Service.LongOperation`, which exits early with `context.Canceled` when the request is canceled.

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
- `JWT_SECRET` (default `dev-secret-change-me`)
- `JWT_ISSUER` (default `hello-mysql`)
- `JWT_TTL` (default `1h`)
- `AUTH_USERNAME` (default `demo`)
- `AUTH_PASSWORD` (default `demo`)
- `LOG_FORMAT` (`text` or `json`, default `text`)
- `LOG_LEVEL` (`debug`, `info`, `warn`, `error`, default `info`)
- `LOG_COLOR` (`auto`, `on`, `off`, default `auto`)
- `LOG_TIME` (`local`, `utc`, `none`, default `local`)
- `LOG_STYLE` (`pretty`, `plain`, `multiline`, default `pretty`, text only)
