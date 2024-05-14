package informer

import (
	"soarca/models/cache"

	"github.com/google/uuid"
)

type IExecutionInformer interface {
	GetExecutions() ([]cache.ExecutionEntry, error)
	GetExecutionReport(executionKey uuid.UUID) (cache.ExecutionEntry, error)
}
