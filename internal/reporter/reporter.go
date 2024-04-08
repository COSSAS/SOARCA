package reporter

import (
	"errors"
	"reflect"

	downstreamReporter "soarca/internal/reporter/downstream_reporter"
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
// Drop error returns
type IWorkflowReporter interface {
	// -> Give info to downstream reporters
	ReportWorkflow(executionContext execution.Metadata, playbook cacao.Playbook)
}
type IStepReporter interface {
	// -> Give info to downstream reporters
	ReportStep(executionContext execution.Metadata, step cacao.Step, outVars cacao.Variables, err error)
}

// High-level reporter class with injection of specific reporters

type Reporter struct {
	reporters []downstreamReporter.IDownStreamReporter
}

const MaxReporters int = 100

func New(reporters []downstreamReporter.IDownStreamReporter) *Reporter {
	instance := Reporter{}
	if instance.reporters == nil {
		instance.reporters = reporters
	}
	return &instance
}

func (reporter *Reporter) RegisterReporters(reporters []downstreamReporter.IDownStreamReporter) error {
	// TODO: how many reporters?
	if (len(reporter.reporters) + len(reporters)) > MaxReporters {
		log.Warning("reporter not registered, too many reporters")
		return errors.New("attempting to register too many reporters")
	}
	reporter.reporters = append(reporter.reporters, reporters...)
	return nil
}

func (reporter *Reporter) ReportWorkflow(executionContext execution.Metadata, playbook cacao.Playbook) {
	log.Trace("reporting workflow")
	workflowEntry := downstreamReporter.WorkflowEntry{ExecutionContext: executionContext, Playbook: playbook}
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflow(workflowEntry)
		if err != nil {
			log.Warning(err)
		}
	}
}

func (reporter *Reporter) ReportStep(executionContext execution.Metadata, step cacao.Step, outVars cacao.Variables, err error) {
	log.Trace("reporting step data")
	stepEntry := downstreamReporter.StepEntry{ExecutionContext: executionContext, Variables: outVars, Error: err}
	for _, rep := range reporter.reporters {
		err := rep.ReportStep(stepEntry)
		if err != nil {
			log.Warning(err)
		}
	}
}
