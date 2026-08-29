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

// Container holds all bootstrapped services and handlers ready for injection into the transport layer.
// FinHandler is the only HTTP handler; other services are retrieved from runtime via getters.
// The HTTP server receives this and only uses it for route registration — no runtime internals exposed.
type Container struct {
	Runtime          *appruntime.Runtime
	FinHandler       *fin.FinHandler
	TransportOptions TransportOptions
}

// New bootstraps services and handlers.
// Runtime already constructs most services. Bootstrap constructs ExecutionRuntime (special case)
// and FinHandler, then injects ExecutionRuntime into runtime.
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

	// Build FinHandler from runtime services
	finHandlerConfig := fin.Config{
		RegistrationToken:      cfg.Fin.RegistrationToken,
		PollIntervalSeconds:    cfg.Fin.PollIntervalSeconds,
		LongPollTimeoutSeconds: cfg.Fin.LongPollTimeoutSeconds,
		JobLeaseSeconds:        cfg.Fin.JobLeaseSeconds,
		StaleAfter:             cfg.Fin.StaleAfter,
	}

	c.FinHandler = fin.NewFinHandler(
		runtime.GetFinRegistry(),
		runtime.GetFinWorkService(),
		finHandlerConfig,
	)

	return c, nil
}
