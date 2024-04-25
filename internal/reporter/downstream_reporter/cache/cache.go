package cache

import (
	"soarca/internal/guid"
	"soarca/models/cacao"
	"soarca/models/reporter/cache"
	"soarca/utils"
	"strconv"

	"github.com/google/uuid"
)

type Cache struct {
	Size  int
	Cache map[string]cache.ExecutionEntry // Cached up to max
}

const MaxExecutions int = 10
const MaxSteps int = 10

func New() *Cache {
	maxExecutions, _ := strconv.Atoi(utils.GetEnv("MAX_EXECUTIONS", strconv.Itoa(MaxExecutions)))
	return &Cache{Size: maxExecutions}
}

func (cacheReporter *Cache) AddExecution(executionEntry cache.ExecutionEntry) {
	// TODO
}

func (cacheReporter *Cache) ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) error {
	// TODO
	return nil
}

func (cacheReporter *Cache) ReportStep(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	// TODO
	return nil
}

func (cacheReporter *Cache) ReportWorkflowUpstream(executionId uuid.UUID, playbookId guid.Guid) (cache.WorkflowResult, error) {
	// TODO
	return cache.WorkflowResult{}, nil
}
func (cacheReporter *Cache) ReportStepUpstream(executionId uuid.UUID, stepId guid.Guid) (cache.StepResult, error) {
	// TODO
	return cache.StepResult{}, nil
}
