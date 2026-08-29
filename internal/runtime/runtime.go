package runtime

import (
	"context"
	"fmt"
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	"soarca/internal/services"
	finsvc "soarca/internal/services/fin"
	manualsvc "soarca/internal/services/manual"
	playbookservice "soarca/internal/services/playbook"
	reporterservice "soarca/internal/services/reporter"
	triggerservice "soarca/internal/services/trigger"
	"soarca/internal/storage"
	storagememory "soarca/internal/storage/memory"
	storagemongo "soarca/internal/storage/mongodb"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/reporting/reporter/downstream_reporter/cache"
	timeutil "soarca/pkg/utils/time"
	"soarca/pkg/utils/guid"
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
}

// Runtime holds all wired application dependencies for the core SOAR orchestrator.
// It constructs and owns application services (non-ExecutionRuntime).
// Bootstrap handles ExecutionRuntime construction due to circular dependency.
// Transport layers retrieve services from getters, never constructing services directly.
type Runtime struct {
	// Core infrastructure (shared across services)
	PlaybookStore storage.PlaybookStore
	FinStore      storage.FinStore
	Cache         *cache.Cache
	Interaction   *interaction.InteractionController
	FinQueue      *queue.Queue

	// Application services (ExecutionRuntime injected by bootstrap due to circular dep)
	executionRuntime services.ExecutionRuntime
	finRegistry      services.FinRegistry
	finWorkService   services.FinWorkService
	manualInbox      services.ManualInbox
	triggerService   services.TriggerService
	playbookService  services.PlaybookService
	reporterService  services.ReporterService
}

// New creates and initializes application dependencies and services.
// Note: ExecutionRuntime is constructed by bootstrap due to circular dependency
// (ExecutionRuntime needs runtime as a parameter).
func New(opts Options) (*Runtime, error) {
	runtime := &Runtime{}

	// Initialize storage layer
	if err := runtime.initializeStorage(opts.Storage); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize shared infrastructure
	runtime.Cache = cache.New(&timeutil.Time{}, opts.Cache.MaxExecutions)
	runtime.Interaction = interaction.New([]interaction.IInteractionIntegrationNotifier{})
	runtime.FinQueue = queue.New()

	// Create FIN services
	runtime.finRegistry = finsvc.NewRegistry(
		runtime.FinStore,
		finsvc.RegistryConfig{
			RegistrationToken: opts.Fin.RegistrationToken,
			StaleAfter:        opts.Fin.StaleAfter,
		},
		new(guid.Guid),
	)

	runtime.finWorkService = finsvc.NewWorkService(
		runtime.FinStore,
		runtime.FinQueue,
		finsvc.WorkServiceConfig{
			LongPollTimeoutSeconds: opts.Fin.LongPollTimeoutSeconds,
			JobLeaseSeconds:        opts.Fin.JobLeaseSeconds,
		},
	)

	// Create manual service
	runtime.manualInbox = manualsvc.NewInbox(runtime.Interaction)

	// Create trigger service
	// Note: TriggerService depends on ExecutionRuntime, which will be injected by bootstrap
	runtime.triggerService = nil // Injected by bootstrap via SetExecutionRuntime

	// Create playbook service
	runtime.playbookService = playbookservice.New(runtime.PlaybookStore)

	// Create reporter service
	runtime.reporterService = reporterservice.New(runtime.Cache)

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

	r.PlaybookStore = store.Playbooks()
	r.FinStore = store.Fins()
	return nil
}

// Close releases any resources held by the app.
func (r *Runtime) Close() error {
	if r.FinQueue != nil {
		r.FinQueue.Close()
	}
	return nil
}

// GetPlaybookStore returns the playbook store.
func (r *Runtime) GetPlaybookStore() storage.PlaybookStore {
	return r.PlaybookStore
}

// GetFinStore returns the FIN store.
func (r *Runtime) GetFinStore() storage.FinStore {
	return r.FinStore
}

// GetCache returns the execution result cache.
func (r *Runtime) GetCache() *cache.Cache {
	return r.Cache
}

// GetInteraction returns the interaction controller.
func (r *Runtime) GetInteraction() *interaction.InteractionController {
	return r.Interaction
}

// GetFinQueue returns the FIN job queue.
func (r *Runtime) GetFinQueue() *queue.Queue {
	return r.FinQueue
}

// ============================================================================
// SERVICE GETTERS (for HTTP handlers and other transport layers)
// ============================================================================

// GetExecutionRuntime returns the core execution runtime service.
func (r *Runtime) GetExecutionRuntime() services.ExecutionRuntime {
	return r.executionRuntime
}

// GetFinRegistry returns the FIN registry service.
func (r *Runtime) GetFinRegistry() services.FinRegistry {
	return r.finRegistry
}

// GetFinWorkService returns the FIN work service.
func (r *Runtime) GetFinWorkService() services.FinWorkService {
	return r.finWorkService
}

// GetManualInbox returns the manual interaction inbox service.
func (r *Runtime) GetManualInbox() services.ManualInbox {
	return r.manualInbox
}

// GetTriggerService returns the trigger orchestration service.
func (r *Runtime) GetTriggerService() services.TriggerService {
	return r.triggerService
}

// GetPlaybookService returns the playbook CRUD service.
func (r *Runtime) GetPlaybookService() services.PlaybookService {
	return r.playbookService
}

// GetReporterService returns the execution reporting service.
func (r *Runtime) GetReporterService() services.ReporterService {
	return r.reporterService
}

// ============================================================================
// BOOTSTRAP INJECTION (for services with circular dependencies)
// ============================================================================

// SetExecutionRuntime injects the ExecutionRuntime service.
// Called by bootstrap after constructing ExecutionRuntime.
func (r *Runtime) SetExecutionRuntime(executionRuntime services.ExecutionRuntime) {
	// Note: TriggerService can now be constructed with ExecutionRuntime
	r.triggerService = triggerservice.New(executionRuntime, r.PlaybookStore)
}
