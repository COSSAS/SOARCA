package cache

import (
	"soarca/models/cacao"
	"soarca/models/report"
	"soarca/utils"
	"strconv"

	"github.com/google/uuid"
)

type Cache struct {
	Size  int
	Cache map[string]report.ExecutionEntry // Cached up to max
}

const MaxExecutions int = 10
const MaxSteps int = 10

func New() *Cache {
	maxExecutions, _ := strconv.Atoi(utils.GetEnv("MAX_EXECUTIONS", strconv.Itoa(MaxExecutions)))
	return &Cache{Size: maxExecutions}
}

func (cacheReporter *Cache) AddExecution(executionEntry report.ExecutionEntry) {
	// TODO
}

func (cacheReporter *Cache) GetExecutionReport(executionId uuid.UUID, playbook cacao.Playbook) error {
	// TODO
	return nil
}

func (cacheReporter *Cache) ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) error {
	// TODO
	return nil
}

func (cacheReporter *Cache) ReportStep(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	// TODO
	return nil
}
