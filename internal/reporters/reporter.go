package reporters

import (
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
)

//TODO:
// DONE Add error logging in the reporter
// DONE In decomposer and executer just discard with _
// Add cache to the reporter for caching reports outputs
// Add tests for caching

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// Reporter interfaces

type IWorkflowReporter interface {
	ReportWorkflow(workflow cacao.Workflow) (interface{}, error)
	//ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error
}

type IStepReporter interface {
	ReportStep(step cacao.Step, out_vars cacao.Variables, err error) (interface{}, error)
}

// High-level reporter class with injection of specific reporters

type Reporter struct {
	workflowReporters    []IWorkflowReporter
	stepReporters        []IStepReporter
	workflowReportsCache ReporterCache
	stepReportsCache     ReporterCache
}

func New(workflowReporters []IWorkflowReporter, stepReporters []IStepReporter) *Reporter {
	return &Reporter{
		workflowReporters:    workflowReporters,
		stepReporters:        stepReporters,
		workflowReportsCache: ReporterCache{size: 5},
		stepReportsCache:     ReporterCache{size: 20},
	}
}

func (reporter *Reporter) RegisterWorkflowReporters(workflowReporters []IWorkflowReporter) []IWorkflowReporter {
	reporter.workflowReporters = append(reporter.workflowReporters, workflowReporters...)
	return reporter.workflowReporters
}

func (reporter *Reporter) RegisterStepReporters(stepReporters []IStepReporter) []IStepReporter {
	reporter.stepReporters = append(reporter.stepReporters, stepReporters...)
	return reporter.stepReporters
}

func (reporter *Reporter) ReportWorkflow(workflow cacao.Workflow) (interface{}, error) {
	log.Trace("reporting workflow")
	for _, rep := range reporter.workflowReporters {
		res, err := rep.ReportWorkflow(workflow)
		if err != nil {
			log.Warning(err)
		}
		reporter.workflowReportsCache.addToCache(res)
	}
	// Errors are handled internally to the Reporter component
	return reporter.workflowReportsCache.cache[0], nil
}

func (reporter *Reporter) ReportStep(step cacao.Step, out_vars cacao.Variables, err error) (interface{}, error) {
	log.Trace("reporting step data")
	for _, rep := range reporter.stepReporters {
		res, err := rep.ReportStep(step, out_vars, err)
		if err != nil {
			log.Warning(err)
		}
		reporter.stepReportsCache.addToCache(res)
	}
	// Errors are handled internally to the Reporter component
	return reporter.stepReportsCache.cache[0], nil
}
