package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"sync"
	"testing"
)

const cleanupDriverName = "sqlite-cleanup-driver"

var (
	cleanupOnce sync.Once
	cleanupDrv  = &cleanupDriver{}
)

type cleanupDriver struct {
	mu       sync.Mutex
	closeErr error
}

func (d *cleanupDriver) SetCloseErr(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closeErr = err
}

func (d *cleanupDriver) getCloseErr() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.closeErr
}

func (d *cleanupDriver) Open(_ string) (driver.Conn, error) {
	return &cleanupConn{d: d}, nil
}

type cleanupConn struct {
	d *cleanupDriver
}

func (c *cleanupConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *cleanupConn) Close() error {
	return c.d.getCloseErr()
}

func (c *cleanupConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *cleanupConn) Ping(_ context.Context) error {
	return nil
}

func registerCleanupDriver() {
	cleanupOnce.Do(func() {
		sql.Register(cleanupDriverName, cleanupDrv)
	})
}

func TestCleanupDBWrapsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CleanupDB(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected wrapped context error")
	}
	if !strings.Contains(err.Error(), "cleanup") {
		t.Fatalf("expected cleanup context in error, got %q", err.Error())
	}
}

func TestCleanupDBWrapsCloseError(t *testing.T) {
	registerCleanupDriver()
	closeErr := errors.New("close failed")
	cleanupDrv.SetCloseErr(closeErr)

	db, err := sql.Open(cleanupDriverName, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	err = CleanupDB(context.Background(), db)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("expected wrapped close error")
	}
	if !strings.Contains(err.Error(), "cleanup") {
		t.Fatalf("expected cleanup context in error, got %q", err.Error())
	}
}
