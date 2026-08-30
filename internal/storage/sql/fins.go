package sql

import (
	"context"
	databasesql "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/fin"
)

type finStore struct {
	db *databasesql.DB
}

func (s *finStore) Create(ctx context.Context, record fin.Record) error {
	capabilities, err := encodeCapabilities(record)
	if err != nil {
		return err
	}

	const query = `INSERT INTO fins
		(fin_id, fin_token_hash, display_name, protocol_version, capabilities, registered_at, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.ExecContext(ctx, query,
		record.FinId, record.FinTokenHash, record.DisplayName,
		record.ProtocolVersion, capabilities,
		record.RegisteredAt, record.LastSeen,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrConflict
		}
		return fmt.Errorf("create fin: %w", err)
	}
	return nil
}

func (s *finStore) Get(ctx context.Context, finID string) (fin.Record, error) {
	return s.queryOne(ctx, finSelect+` WHERE fin_id = $1`, finID)
}

func (s *finStore) GetByTokenHash(ctx context.Context, tokenHash string) (fin.Record, error) {
	return s.queryOne(ctx, finSelect+` WHERE fin_token_hash = $1`, tokenHash)
}

func (s *finStore) List(ctx context.Context) ([]fin.Record, error) {
	rows, err := s.db.QueryContext(ctx, finSelect+` ORDER BY fin_id`)
	if err != nil {
		return nil, fmt.Errorf("list fins: %w", err)
	}
	defer rows.Close()

	records := make([]fin.Record, 0)
	for rows.Next() {
		record, err := scanFin(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (s *finStore) Touch(ctx context.Context, finID string, at time.Time) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE fins SET last_seen = $2 WHERE fin_id = $1`, finID, at)
	if err != nil {
		return fmt.Errorf("touch fin: %w", err)
	}
	return requireOneRow(result, "touch fin")
}

func (s *finStore) Delete(ctx context.Context, finID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM fins WHERE fin_id = $1`, finID)
	if err != nil {
		return fmt.Errorf("delete fin: %w", err)
	}
	return requireOneRow(result, "delete fin")
}

const finSelect = `SELECT fin_id, fin_token_hash, display_name, protocol_version,
	capabilities, registered_at, last_seen FROM fins`

func (s *finStore) queryOne(ctx context.Context, query string, arg any) (fin.Record, error) {
	record, err := scanFin(s.db.QueryRowContext(ctx, query, arg))
	if errors.Is(err, databasesql.ErrNoRows) {
		return fin.Record{}, storage.ErrNotFound
	}
	return record, err
}

// scanner covers both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanFin(row scanner) (fin.Record, error) {
	var (
		record       fin.Record
		capabilities []byte
	)
	err := row.Scan(
		&record.FinId, &record.FinTokenHash, &record.DisplayName,
		&record.ProtocolVersion, &capabilities,
		&record.RegisteredAt, &record.LastSeen,
	)
	if err != nil {
		return fin.Record{}, err
	}
	if err := json.Unmarshal(capabilities, &record.Capabilities); err != nil {
		return fin.Record{}, fmt.Errorf("decode fin capabilities %s: %w", record.FinId, err)
	}
	return record, nil
}

func encodeCapabilities(record fin.Record) ([]byte, error) {
	capabilities := record.Capabilities
	if capabilities == nil {
		capabilities = []fin.Capability{}
	}
	encoded, err := json.Marshal(capabilities)
	if err != nil {
		return nil, fmt.Errorf("encode fin capabilities %s: %w", record.FinId, err)
	}
	return encoded, nil
}
