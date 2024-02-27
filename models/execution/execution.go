package execution

import (
	"github.com/google/uuid"
)

type Metadata struct {
	ExecutionId uuid.UUID
	PlaybookId  string
	StepId      string
}
