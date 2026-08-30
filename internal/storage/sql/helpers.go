package sql

import (
	databasesql "database/sql"
	"fmt"
	"strings"

	"soarca/internal/storage"
)

// requireOneRow turns "statement affected nothing" into ErrNotFound, which is
// how the storage contract reports a missing row.
func requireOneRow(result databasesql.Result, action string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	if affected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

// isUniqueViolation reports whether err is a primary key or unique constraint
// failure. SQLite and PostgreSQL surface this differently and neither exposes a
// portable sentinel, so the check is on the message.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") || // sqlite
		strings.Contains(msg, "duplicate key value") // postgres
}
