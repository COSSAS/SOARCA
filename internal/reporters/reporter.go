package reporters

import (
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// Reporter interfaces

type IWorkflowReporter interface {
	ReportWorkflow(workflow cacao.Workflow) error
	//ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error
}

type IStepReporter interface {
	ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error
}

// High-level reporter class with injection of specific reporters

type Reporter struct {
	workflowReporters []IWorkflowReporter
	stepReporters     []IStepReporter
}

func New(workflowReporters []IWorkflowReporter, stepReporters []IStepReporter) *Reporter {
	return &Reporter{workflowReporters: workflowReporters, stepReporters: stepReporters}
}

func (reporter *Reporter) RegisterWorkflowReporters(workflowReporters []IWorkflowReporter) []IWorkflowReporter {
	reporter.workflowReporters = append(reporter.workflowReporters, workflowReporters...)
	return reporter.workflowReporters
}

func (reporter *Reporter) RegisterStepReporters(stepReporters []IStepReporter) []IStepReporter {
	reporter.stepReporters = append(reporter.stepReporters, stepReporters...)
	return reporter.stepReporters
}

func (reporter *Reporter) ReportWorkflow(workflow cacao.Workflow) error {
	log.Trace("reporting workflow")
	for _, rep := range reporter.workflowReporters {
		err := rep.ReportWorkflow(workflow)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reporter *Reporter) ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error {
	log.Trace("reporting step data")
	for _, rep := range reporter.stepReporters {
		err := rep.ReportStep(step, out_vars, err)
		if err != nil {
			return err
		}
	}
	return nil
}