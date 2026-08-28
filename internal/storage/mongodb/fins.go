package mongodb

import (
	"context"
	"errors"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/fin"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var _ storage.FinStore = (*FinStore)(nil)

type FinStore struct {
	col *mongo.Collection
}

func newFinStore(col *mongo.Collection) *FinStore {
	return &FinStore{col: col}
}

func (s *FinStore) Create(ctx context.Context, record fin.Record) error {
	_, err := s.col.InsertOne(ctx, record)
	if isDuplicate(err) {
		return storage.ErrConflict
	}
	return err
}

func (s *FinStore) Get(ctx context.Context, finID string) (fin.Record, error) {
	var record fin.Record
	err := s.col.FindOne(ctx, bson.M{"_id": finID}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fin.Record{}, storage.ErrNotFound
	}
	return record, err
}

func (s *FinStore) GetByTokenHash(ctx context.Context, tokenHash string) (fin.Record, error) {
	var record fin.Record
	err := s.col.FindOne(ctx, bson.M{"fin_token_hash": tokenHash}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fin.Record{}, storage.ErrNotFound
	}
	return record, err
}

func (s *FinStore) List(ctx context.Context) ([]fin.Record, error) {
	cursor, err := s.col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var result []fin.Record
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *FinStore) Touch(ctx context.Context, finID string, at time.Time) error {
	result, err := s.col.UpdateOne(ctx,
		bson.M{"_id": finID},
		bson.M{"$set": bson.M{"last_seen": at}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *FinStore) Delete(ctx context.Context, finID string) error {
	result, err := s.col.DeleteOne(ctx, bson.M{"_id": finID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

// isDuplicate reports whether err is a MongoDB duplicate-key write error.
func isDuplicate(err error) bool {
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return false
}
