package database

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

type DatabaseReporter struct {
}

// Workflow instantiation reporting logic
func (database_reporter *DatabaseReporter) ReportWorkflow(workflow cacao.Workflow) (interface{}, error) {
	log.Trace("workflow instantiation reported to database")
	// TODO
	return workflow, nil
}

// Step execution reporting logic
func (database_reporter *DatabaseReporter) ReportStep(step cacao.Step, vars cacao.Variables, err error) (interface{}, error) {
	log.Trace("step execution reported to database")
	// TODO
	return new(interface{}), nil
}
