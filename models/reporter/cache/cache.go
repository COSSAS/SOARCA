package cache

import (
	"soarca/internal/guid"
	"soarca/models/cacao"
)

type ExecutionEntry struct {
	Started        string
	Ended          string
	WorkflowResult map[string]WorkflowResult // Cached up to max
	StepResults    map[string]StepResult     // Cached up to max
	Status         string
}

type WorkflowResult struct {
	ExecutionId    guid.IGuid
	PlaybookId     string
	Started        string
	Ended          string
	Status         string
	PlaybookResult error
}

type StepResult struct {
	ExecutionId guid.Guid
	StepId      guid.IGuid
	Started     string
	Ended       string
	Variables   cacao.Variables
	Status      string
	Errors      []error
}
