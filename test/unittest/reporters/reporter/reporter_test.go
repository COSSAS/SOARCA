package reporter_test

import (
	"errors"
	"soarca/internal/reporter"
	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/test/unittest/mocks/mock_reporter"
	"testing"

	"github.com/go-playground/assert/v2"
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

func TestRegisterTooManyReporters(t *testing.T) {
	too_many_reporters := make([]ds_reporter.IDownStreamReporter, reporter.MaxReporters+1)
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
	for i := range too_many_reporters {
		too_many_reporters[i] = &mock_ds_reporter
	}

	reporter := reporter.New([]ds_reporter.IDownStreamReporter{})
	err := reporter.RegisterReporters(too_many_reporters)
	if err == nil {
		t.Fail()
	}
	expected_err := errors.New("attempting to register too many reporters")
	assert.Equal(t, expected_err, err)
}
