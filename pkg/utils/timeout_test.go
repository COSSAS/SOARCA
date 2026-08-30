package utils

import (
	"os"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestDefaultStepTimeoutFallsBackWhenUnset(t *testing.T) {
	previous, wasSet := os.LookupEnv("DEFAULT_STEP_TIMEOUT_SECONDS")
	os.Unsetenv("DEFAULT_STEP_TIMEOUT_SECONDS")
	t.Cleanup(func() {
		if wasSet {
			os.Setenv("DEFAULT_STEP_TIMEOUT_SECONDS", previous)
		}
	})

	assert.Equal(t, DefaultStepTimeout(), 10*time.Minute)
}

func TestDefaultStepTimeoutReadsEnvVar(t *testing.T) {
	t.Setenv("DEFAULT_STEP_TIMEOUT_SECONDS", "120")
	assert.Equal(t, DefaultStepTimeout(), 2*time.Minute)
}

func TestDefaultStepTimeoutFallsBackWhenEnvVarIsInvalid(t *testing.T) {
	t.Setenv("DEFAULT_STEP_TIMEOUT_SECONDS", "not-a-number")
	assert.Equal(t, DefaultStepTimeout(), 10*time.Minute)
}

func TestDefaultStepTimeoutFallsBackWhenEnvVarIsNonPositive(t *testing.T) {
	t.Setenv("DEFAULT_STEP_TIMEOUT_SECONDS", "0")
	assert.Equal(t, DefaultStepTimeout(), 10*time.Minute)
}
