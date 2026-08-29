package bootstrap

import (
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	execservice "soarca/internal/services/execution"
	finsvc "soarca/internal/services/fin"
	manualsvc "soarca/internal/services/manual"
	"soarca/internal/services"
	playbookservice "soarca/internal/services/playbook"
	reporterservice "soarca/internal/services/reporter"
	triggerservice "soarca/internal/services/trigger"
	"soarca/pkg/api/fin"
	"soarca/pkg/utils/guid"
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
// The HTTP server receives this and only uses it for route registration — no runtime internals exposed.
type Container struct {
	ExecutionRuntime services.ExecutionRuntime
	FinHandler       *fin.FinHandler
	TriggerService   services.TriggerService
	PlaybookService  services.PlaybookService
	ReporterService  services.ReporterService
	ManualInbox      services.ManualInbox
	TransportOptions TransportOptions
}

// New bootstraps all services and handlers.
// The workflow factory (capability wiring, decomposer construction) is built here,
// not in the HTTP transport layer.
func New(runtime *appruntime.Runtime, cfg config.Config) (*Container, error) {
	c := &Container{
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
	// It implements decomposer_controller.IController so the runtime can call NewDecomposer().
	wf := newWorkflowFactory(runtime, cfg)

	// Build execution runtime
	c.ExecutionRuntime = execservice.New(runtime, wf)

	// Build FIN services and handler
	finRegistry := finsvc.NewRegistry(
		runtime.GetFinStore(),
		finsvc.RegistryConfig{
			RegistrationToken: cfg.Fin.RegistrationToken,
			StaleAfter:        cfg.Fin.StaleAfter,
		},
		new(guid.Guid),
	)

	finWorkService := finsvc.NewWorkService(
		runtime.GetFinStore(),
		runtime.GetFinQueue(),
		finsvc.WorkServiceConfig{
			LongPollTimeoutSeconds: cfg.Fin.LongPollTimeoutSeconds,
			JobLeaseSeconds:        cfg.Fin.JobLeaseSeconds,
		},
	)

	finHandlerConfig := fin.Config{
		RegistrationToken:      cfg.Fin.RegistrationToken,
		PollIntervalSeconds:    cfg.Fin.PollIntervalSeconds,
		LongPollTimeoutSeconds: cfg.Fin.LongPollTimeoutSeconds,
		JobLeaseSeconds:        cfg.Fin.JobLeaseSeconds,
		StaleAfter:             cfg.Fin.StaleAfter,
	}

	c.FinHandler = fin.NewFinHandler(finRegistry, finWorkService, finHandlerConfig)

	// Build other services
	c.TriggerService = triggerservice.New(c.ExecutionRuntime, runtime.GetPlaybookStore())
	c.PlaybookService = playbookservice.New(runtime.GetPlaybookStore())
	c.ReporterService = reporterservice.New(runtime.GetCache())
	c.ManualInbox = manualsvc.NewInbox(runtime.GetInteraction())

	return c, nil
}
