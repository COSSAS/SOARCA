package informer

import (
	"soarca/pkg/models/cache"

	"github.com/google/uuid"
)

type IExecutionInformer interface {
	GetExecutions() ([]cache.ExecutionEntry, error)
	GetExecutionReport(executionKey uuid.UUID) (cache.ExecutionEntry, error)
}
