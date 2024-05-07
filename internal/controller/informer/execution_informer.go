package informer

import (
	"soarca/models/cache"

	"github.com/google/uuid"
)

type IExecutionInformer interface {
	GetExecutionsIds() []string
	GetExecutionReport(executionKey uuid.UUID) (cache.ExecutionEntry, error)
}
