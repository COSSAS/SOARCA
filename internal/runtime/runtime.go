package runtime

import (
	"context"
	"fmt"
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	"soarca/internal/runs"
	"soarca/internal/runs/engine"
	"soarca/internal/services"
	finsvc "soarca/internal/services/fin"
	manualsvc "soarca/internal/services/manual"
	playbookservice "soarca/internal/services/playbook"
	"soarca/internal/storage"
	storagememory "soarca/internal/storage/memory"
	storagemongo "soarca/internal/storage/mongodb"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/reporting/reporter/downstream_reporter/cache"
	"soarca/pkg/utils/guid"
	timeutil "soarca/pkg/utils/time"
)

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Runtime{}).PkgPath(), logger.Info, "", logger.Json)
}

// Options contains all configuration needed to build the runtime and its services.
type Options struct {
	Storage config.StorageConfig
	Cache   config.CacheConfig
	Fin     config.FinConfig
	HTTP    config.HTTPConfig
	TheHive config.TheHiveConfig
}

// Operations is the use case surface the runtime offers to any driver
// (HTTP, gRPC, CLI, embedded SDK). It carries behaviour only: no
// infrastructure and no getters to reach through. Drivers hold this value,
// never the Runtime itself.
type Operations struct {
	Playbooks  services.PlaybookService
	Executions runs.Runner
	Fins       services.FinRegistry
	Work       services.FinWorkService
	Manual     services.ManualInbox
}

// Runtime owns construction and lifetime of the orchestrator's dependencies.
type Runtime struct {
	// Core infrastructure (shared across services)
	playbookStore storage.PlaybookStore
	finStore      storage.FinStore
	cache         *cache.Cache
	interaction   *interaction.InteractionController
	finQueue      *queue.Queue

	// Application services
	executions      runs.Runner
	finRegistry     services.FinRegistry
	finWorkService  services.FinWorkService
	manualInbox     services.ManualInbox
	playbookService services.PlaybookService
}

// New creates and initializes application dependencies and services.
func New(opts Options) (*Runtime, error) {
	runtime := &Runtime{}

	// Initialize storage layer
	if err := runtime.initializeStorage(opts.Storage); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize shared infrastructure
	runtime.cache = cache.New(&timeutil.Time{}, opts.Cache.MaxExecutions)
	runtime.interaction = interaction.New([]interaction.IInteractionIntegrationNotifier{})
	runtime.finQueue = queue.New()

	// Create FIN services
	runtime.finRegistry = finsvc.NewRegistry(
		runtime.finStore,
		finsvc.RegistryConfig{
			RegistrationToken: opts.Fin.RegistrationToken,
			StaleAfter:        opts.Fin.StaleAfter,
		},
		new(guid.Guid),
	)

	runtime.finWorkService = finsvc.NewWorkService(
		runtime.finStore,
		runtime.finQueue,
		finsvc.WorkServiceConfig{
			LongPollTimeoutSeconds: opts.Fin.LongPollTimeoutSeconds,
			JobLeaseSeconds:        opts.Fin.JobLeaseSeconds,
		},
	)

	// Create manual service
	runtime.manualInbox = manualsvc.NewInbox(runtime.interaction)

	// Create playbook service
	runtime.playbookService = playbookservice.New(runtime.playbookStore)

	// Create the execution engine and the execution service that drives it.
	executionEngine := engine.New(engine.Deps{
		Interaction:        runtime.interaction,
		Cache:              runtime.cache,
		FinQueue:           runtime.finQueue,
		FinStore:           runtime.finStore,
		PlaybookStore:      runtime.playbookStore,
		SkipCertValidation: opts.HTTP.SkipCertValidation,
		FinStaleAfter:      opts.Fin.StaleAfter,
		TheHive:            opts.TheHive,
	})
	runtime.executions = runs.New(executionEngine.NewWalker, runtime.playbookStore, runtime.cache)

	return runtime, nil
}

// initializeStorage sets up the playbook and fin stores based on config.
func (r *Runtime) initializeStorage(storageCfg config.StorageConfig) error {
	var store storage.Store
	var err error

	if storageCfg.UseDatabase {
		log.Info("Initializing MongoDB storage")
		store, err = storagemongo.New(context.Background(), storagemongo.Config{URI: storageCfg.MongoDBURI})
		if err != nil {
			return fmt.Errorf("failed to create MongoDB store: %w", err)
		}
	} else {
		log.Info("Initializing in-memory storage")
		store = storagememory.New()
	}

	r.playbookStore = store.Playbooks()
	r.finStore = store.Fins()
	return nil
}

// Close releases any resources held by the app.
func (r *Runtime) Close() error {
	if r.finQueue != nil {
		r.finQueue.Close()
	}
	return nil
}

// Operations returns the use case surface for drivers.
func (r *Runtime) Operations() Operations {
	return Operations{
		Playbooks:  r.playbookService,
		Executions: r.executions,
		Fins:       r.finRegistry,
		Work:       r.finWorkService,
		Manual:     r.manualInbox,
	}
}
