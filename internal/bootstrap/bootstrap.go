package bootstrap

import (
	"reflect"

	"soarca/internal/config"
	"soarca/internal/controller/decomposer_controller"
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

// DecomposerFactory creates decomposers (used by runtime for workflow orchestration).
// This is implemented by the HTTP transport layer.
type DecomposerFactory interface {
	decomposer_controller.IController
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
type Container struct {
	ExecutionRuntime  services.ExecutionRuntime
	FinHandler        *fin.FinHandler
	TriggerService    services.TriggerService
	PlaybookService   services.PlaybookService
	ReporterService   services.ReporterService
	ManualInbox       services.ManualInbox
	Runtime           *appruntime.Runtime
	Config            config.Config
	TransportOptions  TransportOptions
	DecomposerFactory DecomposerFactory
}

// New bootstraps all services and handlers.
func New(runtime *appruntime.Runtime, cfg config.Config, decomposerFactory DecomposerFactory) (*Container, error) {
	c := &Container{
		Runtime:           runtime,
		Config:            cfg,
		DecomposerFactory: decomposerFactory,
		TransportOptions: TransportOptions{
			Server:  cfg.Server,
			Fin:     cfg.Fin,
			HTTP:    cfg.HTTP,
			Auth:    cfg.Auth,
			TheHive: cfg.TheHive,
			CORS:    cfg.CORS,
		},
	}

	// Build execution runtime
	c.ExecutionRuntime = execservice.New(runtime, decomposerFactory)

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
