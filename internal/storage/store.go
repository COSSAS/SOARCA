package storage

import "context"

type Store interface {
	Playbooks() PlaybookStore
	Fins() FinStore
	Close(ctx context.Context) error
}
