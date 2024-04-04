package mock_reporter

import (
	"soarca/internal/reporter/downstream_reporter"

	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Downstream_Reporter) ReportWorkflow(workflowEntry downstream_reporter.WorkflowEntry) error {
	return nil
}

func (reporter *Mock_Downstream_Reporter) ReportStep(stepEntry downstream_reporter.StepEntry) error {
	return nil
}
