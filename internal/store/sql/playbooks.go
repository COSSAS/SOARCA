package sql

import (
	"context"
	databasesql "database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"soarca/internal/store"
	"soarca/internal/playbooks"
	"soarca/pkg/cacao"
)

type playbookStore struct {
	db *databasesql.DB
}

func (s *playbookStore) Create(ctx context.Context, pb cacao.Playbook) error {
	doc, labels, err := encodePlaybook(pb)
	if err != nil {
		return err
	}

	const query = `INSERT INTO playbooks
		(id, name, description, created, modified, valid_from, valid_until, labels, doc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = s.db.ExecContext(ctx, query,
		pb.ID, pb.Name, pb.Description,
		pb.Created, pb.Modified, pb.ValidFrom, pb.ValidUntil,
		labels, doc,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrConflict
		}
		return fmt.Errorf("create playbook: %w", err)
	}
	return nil
}

func (s *playbookStore) Update(ctx context.Context, pb cacao.Playbook) error {
	doc, labels, err := encodePlaybook(pb)
	if err != nil {
		return err
	}

	const query = `UPDATE playbooks SET
		name = $2, description = $3, created = $4, modified = $5,
		valid_from = $6, valid_until = $7, labels = $8, doc = $9
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query,
		pb.ID, pb.Name, pb.Description,
		pb.Created, pb.Modified, pb.ValidFrom, pb.ValidUntil,
		labels, doc,
	)
	if err != nil {
		return fmt.Errorf("update playbook: %w", err)
	}
	return requireOneRow(result, "update playbook")
}

func (s *playbookStore) Get(ctx context.Context, id string) (cacao.Playbook, error) {
	var doc []byte
	err := s.db.QueryRowContext(ctx, `SELECT doc FROM playbooks WHERE id = $1`, id).Scan(&doc)
	if errors.Is(err, databasesql.ErrNoRows) {
		return cacao.Playbook{}, storage.ErrNotFound
	}
	if err != nil {
		return cacao.Playbook{}, fmt.Errorf("get playbook: %w", err)
	}

	var pb cacao.Playbook
	if err := json.Unmarshal(doc, &pb); err != nil {
		return cacao.Playbook{}, fmt.Errorf("decode playbook %s: %w", id, err)
	}
	return pb, nil
}

func (s *playbookStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM playbooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete playbook: %w", err)
	}
	return requireOneRow(result, "delete playbook")
}

func (s *playbookStore) List(ctx context.Context) ([]cacao.Playbook, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT doc FROM playbooks ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list playbooks: %w", err)
	}
	defer rows.Close()

	playbooks := make([]cacao.Playbook, 0)
	for rows.Next() {
		var doc []byte
		if err := rows.Scan(&doc); err != nil {
			return nil, fmt.Errorf("list playbooks: %w", err)
		}
		var pb cacao.Playbook
		if err := json.Unmarshal(doc, &pb); err != nil {
			return nil, fmt.Errorf("decode playbook: %w", err)
		}
		playbooks = append(playbooks, pb)
	}
	return playbooks, rows.Err()
}

// ListMeta reads the extracted columns so whole playbooks never need decoding.
func (s *playbookStore) ListMeta(ctx context.Context) ([]playbooks.Meta, error) {
	const query = `SELECT id, name, description, valid_from, valid_until, labels
		FROM playbooks ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list playbook metadata: %w", err)
	}
	defer rows.Close()

	metas := make([]playbooks.Meta, 0)
	for rows.Next() {
		var (
			meta       playbooks.Meta
			validFrom  databasesql.NullTime
			validUntil databasesql.NullTime
			labels     []byte
		)
		if err := rows.Scan(&meta.ID, &meta.Name, &meta.Description, &validFrom, &validUntil, &labels); err != nil {
			return nil, fmt.Errorf("list playbook metadata: %w", err)
		}
		meta.ValidFrom = validFrom.Time
		meta.ValidUntil = validUntil.Time
		if err := json.Unmarshal(labels, &meta.Labels); err != nil {
			return nil, fmt.Errorf("decode playbook labels: %w", err)
		}
		metas = append(metas, meta)
	}
	return metas, rows.Err()
}

func encodePlaybook(pb cacao.Playbook) (doc []byte, labels []byte, err error) {
	doc, err = json.Marshal(pb)
	if err != nil {
		return nil, nil, fmt.Errorf("encode playbook %s: %w", pb.ID, err)
	}
	labelValues := pb.Labels
	if labelValues == nil {
		labelValues = []string{}
	}
	labels, err = json.Marshal(labelValues)
	if err != nil {
		return nil, nil, fmt.Errorf("encode playbook labels %s: %w", pb.ID, err)
	}
	return doc, labels, nil
}
