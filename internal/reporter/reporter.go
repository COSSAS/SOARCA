package reporter

import (
	"errors"
	"reflect"
	"strconv"

	downstreamReporter "soarca/internal/reporter/downstream_reporter"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/utils"

	"github.com/google/uuid"
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
	ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook)
	ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error)
}
type IStepReporter interface {
	// -> Give info to downstream reporters
	ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables)
	ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error)
}

const MaxReporters int = 10

// High-level reporter class with injection of specific reporters
type Reporter struct {
	reporters    []downstreamReporter.IDownStreamReporter
	maxReporters int
}

func New(reporters []downstreamReporter.IDownStreamReporter) *Reporter {
	maxReporters, _ := strconv.Atoi(utils.GetEnv("MAX_REPORTERS", strconv.Itoa(MaxReporters)))
	instance := Reporter{
		reporters:    reporters,
		maxReporters: maxReporters,
	}
	return &instance
}

func (reporter *Reporter) RegisterReporters(reporters []downstreamReporter.IDownStreamReporter) error {
	if (len(reporter.reporters) + len(reporters)) > reporter.maxReporters {
		log.Warning("reporter not registered, too many reporters")
		return errors.New("attempting to register too many reporters")
	}
	reporter.reporters = append(reporter.reporters, reporters...)
	return nil
}

func (reporter *Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) {
	log.Trace("reporting workflow")
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowStart(executionId, playbook)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) {
	log.Trace("reporting workflow")
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowEnd(executionId, playbook, workflowError)
		if err != nil {
			log.Warning(err)
		}
	}
}

func (reporter *Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables) {
	log.Trace("reporting step data")
	for _, rep := range reporter.reporters {
		err := rep.ReportStepStart(executionId, step, returnVars)
		if err != nil {
			log.Warning(err)
		}
	}
}

func (reporter *Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error) {
	log.Trace("reporting step data")
	for _, rep := range reporter.reporters {
		err := rep.ReportStepEnd(executionId, step, returnVars, stepError)
		if err != nil {
			log.Warning(err)
		}
	}
}
