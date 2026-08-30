// Package sql implements the storage interfaces on SQLite and PostgreSQL.
//
// Documents (playbooks, fin capabilities) are stored as JSON text; columns are
// extracted only where they need to be queried or listed.
package sql

import (
	"context"
	databasesql "database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"soarca/internal/store"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// target is everything needed to open and migrate one database.
type target struct {
	driverName string // as registered with database/sql
	dialect    string // as known to goose
	dsn        string
	filePath   string // SQLite file to create a directory for, "" otherwise
}

type Store struct {
	db        *databasesql.DB
	playbooks *playbookStore
	fins      *finStore
}

// New opens the database named by a URL, applies migrations, and returns a
// ready store. The scheme selects the backend:
//
//	sqlite://soarca.db
//	sqlite://:memory:
//	file:name?mode=memory&cache=shared
//	postgres://user:pass@host:5432/soarca?sslmode=disable
func New(ctx context.Context, databaseURL string) (*Store, error) {
	target, err := parseURL(databaseURL)
	if err != nil {
		return nil, err
	}

	if err := ensureParentDir(target.filePath); err != nil {
		return nil, err
	}

	db, err := databasesql.Open(target.driverName, target.dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", target.driverName, err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to %s database: %w", target.driverName, err)
	}

	if err := migrate(db, target.dialect); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{
		db:        db,
		playbooks: &playbookStore{db: db},
		fins:      &finStore{db: db},
	}, nil
}

// parseURL picks the driver from the URL scheme. database/sql has no notion of
// schemes: it takes a registered driver name and a driver-specific DSN, so the
// mapping has to happen here. A bare path is treated as SQLite so
// DATABASE_URL=./soarca.db works.
func parseURL(raw string) (target, error) {
	trimmed := strings.TrimSpace(raw)

	sqlite := func(path string) (target, error) {
		return target{
			driverName: "sqlite",
			dialect:    "sqlite3",
			dsn:        sqliteDSN(path),
			filePath:   sqliteFilePath(path),
		}, nil
	}

	switch {
	case trimmed == "":
		return target{}, fmt.Errorf("database URL is empty")
	case strings.HasPrefix(trimmed, "postgres://"), strings.HasPrefix(trimmed, "postgresql://"):
		return target{driverName: "pgx", dialect: "postgres", dsn: trimmed}, nil
	case strings.HasPrefix(trimmed, "sqlite://"):
		return sqlite(strings.TrimPrefix(trimmed, "sqlite://"))
	case strings.HasPrefix(trimmed, "sqlite3://"):
		return sqlite(strings.TrimPrefix(trimmed, "sqlite3://"))
	case strings.HasPrefix(trimmed, "file:"):
		// Raw driver DSN: the caller owns every parameter.
		return target{driverName: "sqlite", dialect: "sqlite3", dsn: trimmed}, nil
	case !strings.Contains(trimmed, "://"):
		return sqlite(trimmed)
	default:
		scheme, _, _ := strings.Cut(trimmed, "://")
		return target{}, fmt.Errorf("unsupported database URL scheme %q, want sqlite or postgres", scheme)
	}
}

const (
	// busyTimeout stops concurrent writers failing immediately with
	// SQLITE_BUSY; they wait for the lock instead.
	busyTimeout = "_pragma=busy_timeout(5000)"
	// walJournal lets readers run alongside a writer on a file database.
	walJournal = "_pragma=journal_mode(WAL)"
)

// sqliteDSN turns a path from a sqlite:// URL into a driver DSN, applying
// defaults that make the pool behave normally. Supplying any query string
// takes full control and disables these defaults.
func sqliteDSN(raw string) string {
	path, query, hasQuery := strings.Cut(raw, "?")

	if path == "" || path == ":memory:" {
		// Without a shared cache each pooled connection would get its own
		// empty database, so the migrated schema would keep disappearing.
		if !hasQuery {
			query = "cache=shared&" + busyTimeout
		}
		return "file::memory:?" + query
	}

	if !hasQuery {
		query = walJournal + "&" + busyTimeout
	}
	return "file:" + path + "?" + query
}

// sqliteFilePath returns the file a sqlite:// URL points at, or "" for an
// in-memory database.
func sqliteFilePath(raw string) string {
	path, _, _ := strings.Cut(raw, "?")
	if path == "" || path == ":memory:" {
		return ""
	}
	return path
}

// ensureParentDir creates the directory for a SQLite file so a configured path
// like sqlite://data/soarca.db works without manual setup.
func ensureParentDir(path string) error {
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create database directory %s: %w", dir, err)
	}
	return nil
}

func migrate(db *databasesql.DB, dialect string) error {
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("set migration dialect: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func (s *Store) Playbooks() storage.PlaybookStore { return s.playbooks }

func (s *Store) Fins() storage.FinStore { return s.fins }

func (s *Store) Close(ctx context.Context) error {
	_ = ctx
	return s.db.Close()
}
