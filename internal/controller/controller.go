package controller

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"
)

var log *logger.Log

type Empty struct{}

type transport interface {
	SetupServer() (*gin.Engine, error)
	RunServer(*gin.Engine) error
}

var loadConfig = config.Load
var newRuntime = appruntime.New
var newTransport = func(runtime *appruntime.Runtime, cfg config.Config) transport {
	server, err := httptransport.New(runtime, cfg)
	if err != nil {
		log.Error("Failed to create HTTP transport:", err)
		panic(err)
	}
	return server
}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Initialize loads configuration, wires the runtime, and starts the HTTP server.
func Initialize() error {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("Failed to load configuration:", err)
		return err
	}

	cfg.LogSettings()

	runtime, err := newRuntime(appruntime.Options{
		Storage: cfg.Storage,
		Cache:   cfg.Cache,
	})
	if err != nil {
		log.Error("Failed to initialize application:", err)
		return err
	}
	defer runtime.Close()

	server := newTransport(runtime, cfg)
	engine, err := server.SetupServer()
	if err != nil {
		log.Error("Failed to setup server:", err)
		return err
	}

	if err := server.RunServer(engine); err != nil {
		log.Error("Failed to run server:", err)
		return err
	}

	log.Info("exit")
	return nil
}
