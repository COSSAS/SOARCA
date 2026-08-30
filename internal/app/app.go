package app

import (
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	orchestrator "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

type transport interface {
	SetupServer() (*gin.Engine, error)
	RunServer(*gin.Engine) error
}

var loadConfig = config.Load
var newRuntime = orchestrator.New
var newTransport = func(ops orchestrator.Operations, opts httptransport.Options) transport {
	return httptransport.New(ops, opts)
}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Run loads configuration, wires the orchestrator, and starts the HTTP server.
func Run() error {
	cfg, err := loadConfig()
	if err != nil {
		log.Error("Failed to load configuration:", err)
		return err
	}

	cfg.LogSettings()

	runtime, err := newRuntime(orchestrator.Options{
		Storage: cfg.Storage,
		Cache:   cfg.Cache,
		Fin:     cfg.Fin,
		HTTP:    cfg.HTTP,
		TheHive: cfg.TheHive,
	})
	if err != nil {
		log.Error("Failed to initialize application:", err)
		return err
	}
	defer runtime.Close()

	server := newTransport(runtime.Operations(), httptransport.Options{
		Server: cfg.Server,
		Fin:    cfg.Fin,
		Auth:   cfg.Auth,
		CORS:   cfg.CORS,
	})
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
