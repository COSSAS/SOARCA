package logger_test

import (
	"os"
	"testing"

	logger "soarca/internal/logger"
)

func TestDebugModeLogTrace(t *testing.T) {
	os.Setenv("LOG_MODE", "development")
	log := logger.Logger("testing", logger.Trace, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestDebugModeLogDebug(t *testing.T) {
	os.Setenv("LOG_MODE", "development")
	log := logger.Logger("testing", logger.Debug, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestDebugModeLogInfo(t *testing.T) {
	os.Setenv("LOG_MODE", "development")
	log := logger.Logger("testing", logger.Info, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestDebugModeLogInfoToFile(t *testing.T) {
	os.Setenv("LOG_MODE", "development")
	log := logger.Logger("testing", logger.Info, "test.log", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
}

func TestDebugModeLogInfoMultiple(t *testing.T) {
	os.Setenv("LOG_MODE", "development")
	log := logger.Logger("logger 1", logger.Info, "", logger.Json)
	log2 := logger.Logger("logger 2", logger.Debug, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
	log2.Info("info")
	log2.Debug("debug")
	log2.Trace("trace")
}

func TestProductionModeLogInfoMultiple(t *testing.T) {
	os.Setenv("LOG_MODE", "production")
	log := logger.Logger("logger 1", logger.Info, "", logger.Json)
	log2 := logger.Logger("logger 2", logger.Debug, "", logger.Json)

	log.Info("info")
	log.Debug("debug")
	log.Trace("trace")
	log2.Info("info")
	log2.Debug("debug")
	log2.Trace("trace")
}
