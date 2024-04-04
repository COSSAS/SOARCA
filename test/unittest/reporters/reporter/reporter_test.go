package reporter_test

import (
	"soarca/internal/reporter"
	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/test/unittest/mocks/mock_reporter"
	"testing"
)

// TODO
func TestRegisterReporter(t *testing.T) {
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{})
	err := reporter.RegisterReporters([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})
	if err != nil {
		t.Fail()
	}
}
