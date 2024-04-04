package database

import (
	"reflect"
	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/logger"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type DatabaseReporter struct {
}

// Workflow instantiation reporting logic
func (database_reporter *DatabaseReporter) ReportWorkflow(workflowEntry ds_reporter.WorkflowEntry) error {
	log.Trace("workflow instantiation reported to database")
	// TODO
	return nil
}

// Step execution reporting logic
func (database_reporter *DatabaseReporter) ReportStep(stepEntry ds_reporter.StepEntry) error {
	log.Trace("step execution reported to database")
	// TODO
	return nil
}
