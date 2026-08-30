package reporter

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	downstreamReporter "soarca/pkg/reporting/reporter/downstream_reporter"
	"soarca/pkg/utils"

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
	//
	// metadata.StepExecutionId identifies this specific invocation of
	// metadata.StepId, so downstream reporters can key per-invocation data
	// (e.g. re-executed steps in a while-loop) without collisions.
	ReportStepStart(metadata execution.Metadata, step cacao.Step, returnVars cacao.Variables, at time.Time)
	ReportStepEnd(metadata execution.Metadata, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time)
}

const MaxReporters int = 10

// High-level reporter class with injection of specific reporters
type Reporter struct {
	reporters    []downstreamReporter.IDownStreamReporter
	maxReporters int
	wg           sync.WaitGroup
	reportingch  chan func()
}

func New(reporters []downstreamReporter.IDownStreamReporter) *Reporter {
	maxReporters, _ := strconv.Atoi(utils.GetEnv("MAX_REPORTERS", strconv.Itoa(MaxReporters)))
	instance := Reporter{
		reporters:    reporters,
		maxReporters: maxReporters,
		reportingch:  make(chan func(), 100), // Buffer size can be adjusted
		wg:           sync.WaitGroup{},
	}
	go instance.startReportingProcessor()
	return &instance
}

func (reporter *Reporter) startReportingProcessor() {
	for {
		task, ok := <-reporter.reportingch
		if !ok {
			return
		}
		task()
	}
}

func (reporter *Reporter) RegisterReporters(reporters []downstreamReporter.IDownStreamReporter) {
	if len(reporters) == 0 {
		log.Warning("reporters list is empty. No action taken.")
	}
	if (len(reporter.reporters) + len(reporters)) > reporter.maxReporters {
		log.Warning("too many reporters provided. Not all provided reporters will be instantiated.")
	}

	for _, downstreamRep := range reporters {
		if len(reporter.reporters) >= reporter.maxReporters {
			return
		}
		reporter.reporters = append(reporter.reporters, downstreamRep)
	}
}

// ######################## IWorkflowReporter interface

func (reporter *Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow start", executionId, playbook.ID))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportWorkflowStart(executionId, playbook, at)
			if err != nil {
				log.Trace("reportWorkflowStart error")
				log.Warning(err)
			}
		}
	}
}

func (reporter *Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, playbook: %s] reporting workflow end", executionId, playbook.ID))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportWorkflowEnd(executionId, playbook, workflowError, at)
			if err != nil {
				log.Trace("reportWorkflowEnd error")
				log.Warning(err)
			}
		}
	}
}

// ######################## IStepReporter interface

func (reporter *Reporter) ReportStepStart(metadata execution.Metadata, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s, step-execution: %s] reporting step start", metadata.ExecutionId, step.ID, metadata.StepExecutionId))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportStepStart(metadata, step, returnVars, at)
			if err != nil {
				log.Trace("reportStepStart error")
				log.Warning(err)
			}
		}
	}
}

func (reporter *Reporter) ReportStepEnd(metadata execution.Metadata, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s, step-execution: %s] reporting step end", metadata.ExecutionId, step.ID, metadata.StepExecutionId))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportStepEnd(metadata, step, returnVars, stepError, at)
			if err != nil {
				log.Trace("reportStepEnd error")
				log.Warning(err)
			}
		}
	}
}
