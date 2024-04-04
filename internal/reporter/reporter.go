package reporter

import (
	"errors"
	"reflect"

	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// Reporter interfaces

type IWorkflowReporter interface {
	// -> Give info to downstream reporters
	ReportWorkflow(executionContext execution.Metadata, playbook cacao.Playbook) error
}
type IStepReporter interface {
	// -> Give info to downstream reporters
	ReportStep(executionContext execution.Metadata, step cacao.Step, outVars cacao.Variables, err error) error
}

// High-level reporter class with injection of specific reporters

type Reporter struct {
	reporters []ds_reporter.IDownStreamReporter
}

const MaxReporters int = 100

func New(reporters []ds_reporter.IDownStreamReporter) *Reporter {
	instance := Reporter{}
	if instance.reporters == nil {
		instance.reporters = reporters
	}
	return &instance
}

func (reporter *Reporter) RegisterReporters(reporters []ds_reporter.IDownStreamReporter) error {
	// TODO: how many reporters?
	if (len(reporter.reporters) + len(reporters)) > MaxReporters {
		log.Warning("reporter not registered, too many reporters")
		return errors.New("attempting to register too many reporters")
	}
	reporter.reporters = append(reporter.reporters, reporters...)
	return nil
}

func (reporter *Reporter) ReportWorkflow(executionContext execution.Metadata, playbook cacao.Playbook) error {
	log.Trace("reporting workflow")
	workflowEntry := ds_reporter.WorkflowEntry{ExecutionId: executionContext.ExecutionId, Playbook: playbook}
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflow(workflowEntry)
		if err != nil {
			log.Warning(err)
		}
	}
	// Errors are handled internally to the Reporter component
	return nil
}

func (reporter *Reporter) ReportStep(executionContext execution.Metadata, step cacao.Step, outVars cacao.Variables, err error) error {
	log.Trace("reporting step data")
	stepEntry := ds_reporter.StepEntry{ExecutionContext: executionContext, Variables: outVars, Error: err}
	for _, rep := range reporter.reporters {
		err := rep.ReportStep(stepEntry)
		if err != nil {
			log.Warning(err)
		}
	}
	// Errors are handled internally to the Reporter component
	return nil
}
