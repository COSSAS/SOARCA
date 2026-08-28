package mongodb

import (
	"context"
	"fmt"

	"soarca/internal/storage"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

var _ storage.Store = (*Store)(nil)

const (
	playbooksCollectionName = "playbooks"
	finsCollectionName      = "fins"
)

type Store struct {
	client    *mongo.Client
	playbooks *PlaybookStore
	fins      *FinStore
}

func New(ctx context.Context, cfg Config) (*Store, error) {
	// NOTE: This is an internal package, but the only clean way to parse the database from the uri without implementing our own parser.
	cs, err := connstring.ParseAndValidate(cfg.URI)
	if err != nil {
		return nil, err
	}
	if cs.Database == "" {
		return nil, fmt.Errorf("mongodb URI must include a database name")
	}

	opts := options.Client().ApplyURI(cfg.URI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	db := client.Database(cs.Database)

	return &Store{
		client:    client,
		playbooks: newPlaybookStore(db.Collection(playbooksCollectionName)),
		fins:      newFinStore(db.Collection(finsCollectionName)),
	}, nil
}

func (s *Store) Playbooks() storage.PlaybookStore { return s.playbooks }
func (s *Store) Fins() storage.FinStore           { return s.fins }
func (s *Store) Close(ctx context.Context) error  { return s.client.Disconnect(ctx) }
