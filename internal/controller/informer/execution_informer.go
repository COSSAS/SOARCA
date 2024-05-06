package informer

import (
	"soarca/models/report"

	"github.com/google/uuid"
)

type IExecutionInformer interface {
	GetExecutionsIds() []string
	GetExecutionReport(executionKey uuid.UUID) (report.ExecutionEntry, error)
}
