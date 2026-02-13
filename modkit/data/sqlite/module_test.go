package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
	"github.com/go-modkit/modkit/modkit/testkit"
)

var testDrv = &countingDriver{}

func init() {
	sql.Register(driverName, testDrv)
}

type countingDriver struct {
	mu          sync.Mutex
	openCount   int
	pingCount   int
	closeCount  int
	pingErr     error
	sawDeadline bool
	lastOpenDSN string
}

func (d *countingDriver) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	c := countingDriver{}
	d.openCount = c.openCount
	d.pingCount = c.pingCount
	d.closeCount = c.closeCount
	d.pingErr = nil
	d.sawDeadline = false
	d.lastOpenDSN = ""
}

func (d *countingDriver) SetPingErr(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pingErr = err
}

func (d *countingDriver) Snapshot() (open, ping, closed int, sawDeadline bool, lastOpenDSN string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.openCount, d.pingCount, d.closeCount, d.sawDeadline, d.lastOpenDSN
}

func (d *countingDriver) Open(name string) (driver.Conn, error) {
	d.mu.Lock()
	d.openCount++
	d.lastOpenDSN = name
	d.mu.Unlock()
	return &countingConn{d: d}, nil
}

type countingConn struct {
	d *countingDriver
}

func (c *countingConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *countingConn) Close() error {
	c.d.mu.Lock()
	c.d.closeCount++
	c.d.mu.Unlock()
	return nil
}

func (c *countingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *countingConn) Ping(ctx context.Context) error {
	c.d.mu.Lock()
	c.d.pingCount++
	if _, ok := ctx.Deadline(); ok {
		c.d.sawDeadline = true
	}
	err := c.d.pingErr
	c.d.mu.Unlock()
	return err
}

func TestModuleExportsDialectAndDBTokens(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "test.db")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "0")

	h := testkit.New(t, NewModule(Options{}))
	_ = testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)
	dialect := testkit.Get[sqlmodule.Dialect](t, h, sqlmodule.TokenDialect)
	if dialect != sqlmodule.DialectSQLite {
		t.Fatalf("unexpected dialect: %q", dialect)
	}
}

func TestConnectTimeoutPingBehavior(t *testing.T) {
	cases := []struct {
		name            string
		timeout         string
		wantOpen        int
		wantOpenNonZero bool
		wantPing        int
		wantDeadline    bool
	}{
		{
			name:     "zero timeout skips ping",
			timeout:  "0",
			wantOpen: 0,
			wantPing: 0,
		},
		{
			name:            "non-zero timeout pings with deadline",
			timeout:         "25ms",
			wantOpenNonZero: true,
			wantPing:        1,
			wantDeadline:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDrv.Reset()
			t.Setenv("SQLITE_PATH", "test.db")
			t.Setenv("SQLITE_CONNECT_TIMEOUT", tc.timeout)

			h := testkit.New(t, NewModule(Options{}))
			_ = testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)

			open, ping, _, sawDeadline, _ := testDrv.Snapshot()
			if tc.wantOpenNonZero {
				if open == 0 {
					t.Fatalf("expected open>0, got %d", open)
				}
			} else if open != tc.wantOpen {
				t.Fatalf("expected open=%d, got %d", tc.wantOpen, open)
			}
			if ping != tc.wantPing {
				t.Fatalf("expected ping=%d, got %d", tc.wantPing, ping)
			}
			if tc.wantDeadline != sawDeadline {
				t.Fatalf("expected deadline=%v, got %v", tc.wantDeadline, sawDeadline)
			}
		})
	}
}

func TestPingFailureReturnsTypedBuildErrorAndClosesDB(t *testing.T) {
	testDrv.Reset()
	pingErr := errors.New("ping failed")
	testDrv.SetPingErr(pingErr)
	t.Setenv("SQLITE_PATH", "test.db")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "25ms")

	h := testkit.New(t, NewModule(Options{}))
	_, err := testkit.GetE[*sql.DB](h, sqlmodule.TokenDB)
	if err == nil {
		t.Fatalf("expected error")
	}

	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("expected BuildError, got %T", err)
	}
	if be.Stage != StagePing {
		t.Fatalf("expected stage=%s, got %s", StagePing, be.Stage)
	}
	if be.Provider != driverName {
		t.Fatalf("expected provider=%q, got %q", driverName, be.Provider)
	}
	if be.Token != sqlmodule.TokenDB {
		t.Fatalf("expected token=%q, got %q", sqlmodule.TokenDB, be.Token)
	}
	if !errors.Is(err, pingErr) {
		t.Fatalf("expected error to wrap ping error")
	}

	_, _, closed, _, _ := testDrv.Snapshot()
	if closed == 0 {
		t.Fatalf("expected ping failure path to close the DB")
	}
}

func TestPathConfigBuildsDSNWithSQLiteOptions(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "test.db")
	t.Setenv("SQLITE_BUSY_TIMEOUT", "150ms")
	t.Setenv("SQLITE_JOURNAL_MODE", "wal")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "25ms")

	h := testkit.New(t, NewModule(Options{}))
	_ = testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)

	_, _, _, _, openDSN := testDrv.Snapshot()
	parts := strings.SplitN(openDSN, "?", 2)
	if len(parts) != 2 {
		t.Fatalf("expected DSN to contain query, got %q", openDSN)
	}
	if parts[0] != "test.db" {
		t.Fatalf("expected base path %q, got %q", "test.db", parts[0])
	}
	q, err := url.ParseQuery(parts[1])
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if got := q.Get("_busy_timeout"); got != "150" {
		t.Fatalf("expected _busy_timeout=150, got %q", got)
	}
	if got := q.Get("_journal_mode"); got != "wal" {
		t.Fatalf("expected _journal_mode=wal, got %q", got)
	}
}

func TestDSNConfigAppendsSQLiteOptionsToExistingQuery(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "file:test.db?cache=shared")
	t.Setenv("SQLITE_BUSY_TIMEOUT", "200ms")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "25ms")

	h := testkit.New(t, NewModule(Options{}))
	_ = testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)

	_, _, _, _, openDSN := testDrv.Snapshot()
	parts := strings.SplitN(openDSN, "?", 2)
	if len(parts) != 2 {
		t.Fatalf("expected DSN to contain query, got %q", openDSN)
	}
	if parts[0] != "file:test.db" {
		t.Fatalf("expected base DSN %q, got %q", "file:test.db", parts[0])
	}
	q, err := url.ParseQuery(parts[1])
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if got := q.Get("cache"); got != "shared" {
		t.Fatalf("expected cache=shared, got %q", got)
	}
	if got := q.Get("_busy_timeout"); got != "200" {
		t.Fatalf("expected _busy_timeout=200, got %q", got)
	}
}

func TestNegativeConnectTimeoutFailsWithInvalidConfig(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "test.db")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "-1ms")

	h := testkit.New(t, NewModule(Options{}))
	_, err := testkit.GetE[*sql.DB](h, sqlmodule.TokenDB)
	if err == nil {
		t.Fatalf("expected error")
	}

	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("expected BuildError, got %T", err)
	}
	if be.Stage != StageInvalidConfig {
		t.Fatalf("expected stage=%s, got %s", StageInvalidConfig, be.Stage)
	}
	if be.Provider != driverName {
		t.Fatalf("expected provider=%q, got %q", driverName, be.Provider)
	}
}

func TestCleanupClosesDB(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "test.db")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "25ms")

	h := testkit.New(t, NewModule(Options{}))
	_ = testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)
	if err := h.Close(); err != nil {
		t.Fatalf("close harness: %v", err)
	}

	_, _, closed, _, _ := testDrv.Snapshot()
	if closed == 0 {
		t.Fatalf("expected cleanup to close a DB connection")
	}
}

func TestMultipleSQLiteInstancesBootstrap(t *testing.T) {
	testDrv.Reset()
	t.Setenv("SQLITE_PATH", "file:memdb1?mode=memory&cache=shared")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "0")

	primaryTokens, err := sqlmodule.NamedTokens("primary")
	if err != nil {
		t.Fatalf("primary tokens: %v", err)
	}
	analyticsTokens, err := sqlmodule.NamedTokens("analytics")
	if err != nil {
		t.Fatalf("analytics tokens: %v", err)
	}
	configMod := DefaultConfigModule()

	root := &multiInstanceRootModule{
		imports: []module.Module{
			NewModule(Options{Name: "primary", Config: configMod}),
			NewModule(Options{Name: "analytics", Config: configMod}),
		},
		exports: []module.Token{
			primaryTokens.DB,
			primaryTokens.Dialect,
			analyticsTokens.DB,
			analyticsTokens.Dialect,
		},
	}
	app, err := kernel.Bootstrap(root)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	if _, err := app.Get(primaryTokens.DB); err != nil {
		t.Fatalf("primary db: %v", err)
	}
	if _, err := app.Get(analyticsTokens.DB); err != nil {
		t.Fatalf("analytics db: %v", err)
	}
}

func TestInvalidNameFailsAtBootstrap(t *testing.T) {
	root := &multiInstanceRootModule{
		imports: []module.Module{
			NewModule(Options{Name: "bad name"}),
		},
	}
	_, err := kernel.Bootstrap(root)
	if err == nil {
		t.Fatal("expected bootstrap error")
	}
	var invalidNameErr *sqlmodule.InvalidNameError
	if !errors.As(err, &invalidNameErr) {
		t.Fatalf("expected InvalidNameError, got %T", err)
	}
}

type multiInstanceRootModule struct {
	imports []module.Module
	exports []module.Token
}

func (m *multiInstanceRootModule) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:    "root",
		Imports: m.imports,
		Exports: m.exports,
	}
}
