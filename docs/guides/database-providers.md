# Database Providers

modkit ships a shared SQL contract plus provider modules for Postgres and
SQLite. The goal is to keep feature modules database-agnostic while still
making driver wiring explicit and deterministic.

## Shared SQL Contract

Use the shared contract tokens from `modkit/data/sqlmodule`:

```go
import "github.com/go-modkit/modkit/modkit/data/sqlmodule"

db, err := module.Get[*sql.DB](r, sqlmodule.TokenDB)
if err != nil {
    return nil, err
}
dialect, err := module.Get[sqlmodule.Dialect](r, sqlmodule.TokenDialect)
if err != nil {
    return nil, err
}
```

The contract exports two stable tokens:

- `sqlmodule.TokenDB` -> `*sql.DB`
- `sqlmodule.TokenDialect` -> `sqlmodule.Dialect`

For multi-instance apps, use `sqlmodule.NamedTokens(name)` and pass the same
name into the provider module options.

## Postgres Provider

Package: `modkit/data/postgres`

```go
import "github.com/go-modkit/modkit/modkit/data/postgres"

module.ModuleDef{
    Name: "app",
    Imports: []module.Module{
        postgres.NewModule(postgres.Options{}),
    },
}
```

Configuration:

- Required: `POSTGRES_DSN`
- Optional: `POSTGRES_MAX_OPEN_CONNS`, `POSTGRES_MAX_IDLE_CONNS`,
  `POSTGRES_CONN_MAX_LIFETIME`, `POSTGRES_CONNECT_TIMEOUT`

`POSTGRES_CONNECT_TIMEOUT=0` skips the startup ping. Any non-zero duration
enables a timeout-bound `PingContext` during provider build.

The provider is driver-agnostic. Import a driver in your app (for example,
`_ "github.com/lib/pq"`).

## SQLite Provider

Package: `modkit/data/sqlite`

```go
import "github.com/go-modkit/modkit/modkit/data/sqlite"

module.ModuleDef{
    Name: "app",
    Imports: []module.Module{
        sqlite.NewModule(sqlite.Options{}),
    },
}
```

Configuration:

- Required: `SQLITE_PATH` (path or DSN)
- Optional: `SQLITE_BUSY_TIMEOUT`, `SQLITE_JOURNAL_MODE`,
  `SQLITE_CONNECT_TIMEOUT`

`SQLITE_CONNECT_TIMEOUT=0` skips the startup ping. Any non-zero duration
enables a timeout-bound `PingContext` during provider build.

Like Postgres, the module is driver-agnostic. Import a driver in your app (for
example, `_ "github.com/mattn/go-sqlite3"`).

## Named Instances

For multiple databases in one app, supply a name and use `NamedTokens`:

```go
tokens, err := sqlmodule.NamedTokens("analytics")
if err != nil {
    return err
}

module.ModuleDef{
    Name: "app",
    Imports: []module.Module{
        postgres.NewModule(postgres.Options{Name: "analytics"}),
    },
    Exports: []module.Token{tokens.DB, tokens.Dialect},
}
```

## Migration Note (MySQL -> Shared SQL Contract)

The `hello-mysql` example preserves backward compatibility: it still exports
`database.TokenDB` (value `"database.db"`) and adds `database.TokenDialect`.
For new code, prefer the shared contract tokens (`sqlmodule.TokenDB` and
`sqlmodule.TokenDialect`) and keep driver-specific modules out of feature
packages.

## Examples

- `examples/hello-postgres` — Postgres provider + smoke test
- `examples/hello-sqlite` — SQLite provider + file/in-memory smoke tests
- `examples/hello-mysql` — legacy MySQL example (still compatible)
