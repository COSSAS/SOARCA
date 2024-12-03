package reporter

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	downstreamReporter "soarca/pkg/reporter/downstream_reporter"
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
	ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time)
	ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time)
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

func (reporter *Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step start", executionId, step.ID))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportStepStart(executionId, step, returnVars, at)
			if err != nil {
				log.Trace("reportStepStart error")
				log.Warning(err)
			}
		}
	}
}

func (reporter *Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) {
	log.Trace(fmt.Sprintf("[execution: %s, step: %s] reporting step end", executionId, step.ID))
	reporter.wg.Add(1)
	reporter.reportingch <- func() {
		defer reporter.wg.Done()
		for _, downstreamRep := range reporter.reporters {
			err := downstreamRep.ReportStepEnd(executionId, step, returnVars, stepError, at)
			if err != nil {
				log.Trace("reportStepEnd error")
				log.Warning(err)
			}
		}
	}
}
