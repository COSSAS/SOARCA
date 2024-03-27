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

type IReporter interface {
	ReportWorkflow(workflow cacao.Workflow) error
	ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error
}

// High-level reporter class with injection of specific reporters

type Reporter struct {
	reporters []IReporter
}

func New(reporters []IReporter) *Reporter {

	return &Reporter{reporters: reporters}
}

func (reporter *Reporter) ReportWorkflow(workflow cacao.Workflow) error {
	log.Trace("reporting workflow")
	for _, rep := range reporter.reporters {
		err := rep.ReportWorkflow(workflow)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reporter *Reporter) ReportStep(step cacao.Step, out_vars cacao.Variables, err error) error {
	log.Trace("reporting step data")
	for _, rep := range reporter.reporters {
		err := rep.ReportStep(step, out_vars, err)
		if err != nil {
			return err
		}
	}
	return nil
}
