package reporter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

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
	ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time)
	ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time)
}
type IStepReporter interface {
	// -> Give info to downstream reporters
	ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time)
	ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time)
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
	if len(reporters) == 0 {
		log.Warning("reporters list is empty. No action taken.")
		return nil
	}
	if (len(reporter.reporters) + len(reporters)) > reporter.maxReporters {
		log.Warning("reporter not registered, too many reporters")
		return errors.New("attempting to register too many reporters")
	}
	reporter.reporters = append(reporter.reporters, reporters...)
	return nil
}

// ######################## IWorkflowReporter interface

func (reporter *Reporter) reportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) {
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowStart(executionId, playbook, at)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow start", executionId, playbook.ID))
	go reporter.reportWorkflowStart(executionId, playbook, at)
}

func (reporter *Reporter) reportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) {
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowEnd(executionId, playbook, workflowError, at)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow end", executionId, playbook.ID))
	go reporter.reportWorkflowEnd(executionId, playbook, workflowError, at)
}

// ######################## IStepReporter interface

func (reporter *Reporter) reporStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	for _, rep := range reporter.reporters {
		err := rep.ReportStepStart(executionId, step, returnVars, at)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step start", executionId, step.ID))
	go reporter.reporStepStart(executionId, step, returnVars, at)
}

func (reporter *Reporter) reportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) {
	for _, rep := range reporter.reporters {
		err := rep.ReportStepEnd(executionId, step, returnVars, stepError, at)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step end", executionId, step.ID))
	go reporter.reportStepEnd(executionId, step, returnVars, stepError, at)
}
