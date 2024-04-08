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
	ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook)
}
type IStepReporter interface {
	// -> Give info to downstream reporters
	ReportStep(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error)
}

// High-level reporter class with injection of specific reporters
type Reporter struct {
	reporters []downstreamReporter.IDownStreamReporter
}

const MaxReporters int = 10

func New(reporters []downstreamReporter.IDownStreamReporter) *Reporter {
	instance := Reporter{}
	if instance.reporters == nil {
		instance.reporters = reporters
	}
	return &instance
}

func (reporter *Reporter) RegisterReporters(reporters []downstreamReporter.IDownStreamReporter) error {
	maxReporters, _ := strconv.Atoi(utils.GetEnv("MAX_REPORTERS", strconv.Itoa(MaxReporters)))
	if (len(reporter.reporters) + len(reporters)) > maxReporters {
		log.Warning("reporter not registered, too many reporters")
		return errors.New("attempting to register too many reporters")
	}
	reporter.reporters = append(reporter.reporters, reporters...)
	return nil
}

func (reporter *Reporter) ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) {
	log.Trace("reporting workflow")
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflow(executionId, playbook)
		if err != nil {
			log.Warning(err)
		}
	}
}

func (reporter *Reporter) ReportStep(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error) {
	log.Trace("reporting step data")
	for _, rep := range reporter.reporters {
		err := rep.ReportStep(executionId, step, returnVars, err)
		if err != nil {
			log.Warning(err)
		}
	}
}
