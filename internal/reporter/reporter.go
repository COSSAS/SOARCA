package reporter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

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
	reporters          []downstreamReporter.IDownStreamReporter
	maxReporters       int
	reportingWaitGroup sync.WaitGroup
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

// ######################## IWorkflowReporter interface

func (reporter *Reporter) reportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) {
	defer reporter.reportingWaitGroup.Done()
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowStart(executionId, playbook)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) {
	reporter.reportingWaitGroup.Add(1)
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow start", executionId, playbook.ID))
	go reporter.reportWorkflowStart(executionId, playbook)
}

func (reporter *Reporter) reportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) {
	defer reporter.reportingWaitGroup.Done()
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflowEnd(executionId, playbook, workflowError)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) {
	reporter.reportingWaitGroup.Wait()
	reporter.reportingWaitGroup.Add(1)
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow end", executionId, playbook.ID))
	go reporter.reportWorkflowEnd(executionId, playbook, workflowError)
}

// ######################## IStepReporter interface

func (reporter *Reporter) reporStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables) {
	defer reporter.reportingWaitGroup.Done()
	for _, rep := range reporter.reporters {
		err := rep.ReportStepStart(executionId, step, returnVars)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables) {
	reporter.reportingWaitGroup.Add(1)
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step start", executionId, step.ID))
	go reporter.reporStepStart(executionId, step, returnVars)
}

func (reporter *Reporter) reportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error) {
	defer reporter.reportingWaitGroup.Done()
	for _, rep := range reporter.reporters {
		err := rep.ReportStepEnd(executionId, step, returnVars, stepError)
		if err != nil {
			log.Warning(err)
		}
	}
}
func (reporter *Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error) {
	reporter.reportingWaitGroup.Add(1)
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step end", executionId, step.ID))
	go reporter.reportStepEnd(executionId, step, returnVars, stepError)
}
