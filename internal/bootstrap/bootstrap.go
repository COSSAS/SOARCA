package bootstrap

import (
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	execservice "soarca/internal/services/execution"
	"soarca/pkg/api/fin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// TransportOptions contains the configuration needed by the HTTP transport.
type TransportOptions struct {
	Server  config.ServerConfig
	Fin     config.FinConfig
	HTTP    config.HTTPConfig
	Auth    config.AuthConfig
	TheHive config.TheHiveConfig
	CORS    config.CORSConfig
}

// Container holds the bootstrapped runtime and transport configuration.
// HTTP handlers are composed by the transport layer.
type Container struct {
	Runtime          *appruntime.Runtime
	FinHandler       *fin.FinHandler
	TransportOptions TransportOptions
}

// New bootstraps the execution runtime and injects it into the app runtime.
// HTTP handler construction belongs to the transport layer.
func New(runtime *appruntime.Runtime, cfg config.Config) (*Container, error) {
	c := &Container{
		Runtime: runtime,
		TransportOptions: TransportOptions{
			Server:  cfg.Server,
			Fin:     cfg.Fin,
			HTTP:    cfg.HTTP,
			Auth:    cfg.Auth,
			TheHive: cfg.TheHive,
			CORS:    cfg.CORS,
		},
	}

	// WorkflowFactory owns all capability/executor/reporter wiring.
	// It implements decomposer_controller.IController so ExecutionRuntime can call NewDecomposer().
	wf := newWorkflowFactory(runtime, cfg)

	// Build ExecutionRuntime service
	executionRuntime := execservice.New(runtime, wf)

	// Inject ExecutionRuntime into runtime (also constructs TriggerService there)
	runtime.SetExecutionRuntime(executionRuntime)

	return c, nil
}
