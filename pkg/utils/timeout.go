package utils

import (
	"strconv"
	"time"
)

// defaultStepTimeoutSecondsFallback is used when DEFAULT_STEP_TIMEOUT_SECONDS is unset or invalid.
const defaultStepTimeoutSecondsFallback = 600

// DefaultStepTimeout reads DEFAULT_STEP_TIMEOUT_SECONDS with fallback validation.
func DefaultStepTimeout() time.Duration {
	seconds, err := strconv.Atoi(GetEnv("DEFAULT_STEP_TIMEOUT_SECONDS", strconv.Itoa(defaultStepTimeoutSecondsFallback)))
	if err != nil || seconds <= 0 {
		seconds = defaultStepTimeoutSecondsFallback
	}
	return time.Duration(seconds) * time.Second
}
