package logger_test

import (
	"testing"

	logger "soarca/logger"
)

func TestLogTrace(t *testing.T) {
	log := logger.Logger("testing", logger.Trace, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestLogDebug(t *testing.T) {
	log := logger.Logger("testing", logger.Debug, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestLogInfo(t *testing.T) {
	log := logger.Logger("testing", logger.Info, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestLogInfoToFile(t *testing.T) {
	log := logger.Logger("testing", logger.Info, "test.log", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestLogInfoMultiple(t *testing.T) {
	log := logger.Logger("logger 1", logger.Info, "", logger.Json)
	log2 := logger.Logger("logger 2", logger.Debug, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
	log2.Info("info")
	log2.Debug("debug")
	log2.Trace("trace")
}
