package mongodb

import (
	"context"
	"errors"

	"soarca/internal/storage"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ storage.PlaybookStore = (*PlaybookStore)(nil)

type PlaybookStore struct {
	col *mongo.Collection
}

func newPlaybookStore(col *mongo.Collection) *PlaybookStore {
	return &PlaybookStore{col: col}
}

func (s *PlaybookStore) Create(ctx context.Context, pb cacao.Playbook) error {
	_, err := s.col.InsertOne(ctx, pb)
	if isDuplicate(err) {
		return storage.ErrConflict
	}
	return err
}

func (s *PlaybookStore) Update(ctx context.Context, pb cacao.Playbook) error {
	result, err := s.col.ReplaceOne(ctx, bson.M{"_id": pb.ID}, pb)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *PlaybookStore) Get(ctx context.Context, id string) (cacao.Playbook, error) {
	var pb cacao.Playbook
	err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&pb)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return cacao.Playbook{}, storage.ErrNotFound
	}
	return pb, err
}

func (s *PlaybookStore) Delete(ctx context.Context, id string) error {
	result, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *PlaybookStore) List(ctx context.Context) ([]cacao.Playbook, error) {
	cursor, err := s.col.Find(ctx, bson.D{}, options.Find().SetLimit(100))
	if err != nil {
		return nil, err
	}
	var result []cacao.Playbook
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

var playbookMetaProjection = bson.D{
	{Key: "_id", Value: 1},
	{Key: "name", Value: 1},
	{Key: "description", Value: 1},
	{Key: "valid_from", Value: 1},
	{Key: "valid_until", Value: 1},
	{Key: "labels", Value: 1},
}

func (s *PlaybookStore) ListMeta(ctx context.Context) ([]api.PlaybookMeta, error) {
	opts := options.Find().SetProjection(playbookMetaProjection).SetLimit(100)
	cursor, err := s.col.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	var playbooks []cacao.Playbook
	if err := cursor.All(ctx, &playbooks); err != nil {
		return nil, err
	}
	result := make([]api.PlaybookMeta, len(playbooks))
	for i, pb := range playbooks {
		result[i] = api.PlaybookMeta{
			ID:          pb.ID,
			Name:        pb.Name,
			Description: pb.Description,
			ValidFrom:   pb.ValidFrom,
			ValidUntil:  pb.ValidUntil,
			Labels:      pb.Labels,
		}
	}
	return result, nil
}
