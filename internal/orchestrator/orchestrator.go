package orchestrator

import (
	"context"
	"fmt"
	"reflect"

	"soarca/internal/config"
	fins "soarca/internal/fins"
	"soarca/internal/logger"
	manualsvc "soarca/internal/manual"
	playbookservice "soarca/internal/playbooks/library"
	"soarca/internal/reporting/reporter/downstream_reporter/runstate"
	"soarca/internal/runs"
	"soarca/internal/runs/engine"
	storage "soarca/internal/store"
	storagesql "soarca/internal/store/sql"
	"soarca/internal/workflow/capability/fin/queue"
	"soarca/internal/workflow/capability/manual/inbox"
	"soarca/pkg/utils/guid"
	timeutil "soarca/pkg/utils/time"
)

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Runtime{}).PkgPath(), logger.Info, "", logger.Json)
}

// Options contains all configuration needed to build the runtime and its orchestrator.
type Options struct {
	Storage  config.StorageConfig
	RunState config.RunStateConfig
	Fin      config.FinConfig
	HTTP     config.HTTPConfig
	TheHive  config.TheHiveConfig
}

// Operations is the use case surface the runtime offers to any driver
// (HTTP, gRPC, CLI, embedded SDK). It carries behaviour only: no
// infrastructure and no getters to reach through. Drivers hold this value,
// never the Runtime itself.
type Operations struct {
	Playbooks PlaybookService
	Runs      runs.Runner
	Fins      FinRegistry
	Work      FinWorkService
	Manual    ManualInbox
}

// Runtime owns construction and lifetime of the orchestrator's dependencies.
type Runtime struct {
	// Core infrastructure (shared across services)
	store         storage.Store
	playbookStore storage.PlaybookStore
	finStore      storage.FinStore
	runState      *runstate.RunState
	manual        *inbox.Inbox
	finQueue      *queue.Queue

	// Application services
	runs            runs.Runner
	finRegistry     FinRegistry
	finWorkService  FinWorkService
	manualInbox     ManualInbox
	playbookService PlaybookService
}

// New creates and initializes application dependencies and orchestrator.
func New(opts Options) (*Runtime, error) {
	runtime := &Runtime{}

	// Initialize storage layer
	if err := runtime.initializeStorage(opts.Storage); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize shared infrastructure
	runtime.runState = runstate.New(&timeutil.Time{}, opts.RunState.MaxRuns)
	runtime.manual = inbox.New([]inbox.Notifier{})
	runtime.finQueue = queue.New()

	// Create FIN services
	runtime.finRegistry = fins.NewRegistry(
		runtime.finStore,
		fins.RegistryConfig{
			RegistrationToken: opts.Fin.RegistrationToken,
			StaleAfter:        opts.Fin.StaleAfter,
		},
		new(guid.Guid),
	)

	runtime.finWorkService = fins.NewWorkService(
		runtime.finStore,
		runtime.finQueue,
		fins.WorkServiceConfig{
			LongPollTimeoutSeconds: opts.Fin.LongPollTimeoutSeconds,
			JobLeaseSeconds:        opts.Fin.JobLeaseSeconds,
		},
	)

	// Create manual service
	runtime.manualInbox = manualsvc.NewInbox(runtime.manual)

	// Create playbook service
	runtime.playbookService = playbookservice.New(runtime.playbookStore)

	// Create the run engine and the run service that drives it.
	runEngine := engine.New(engine.Deps{
		ManualInbox:        runtime.manual,
		RunState:           runtime.runState,
		FinQueue:           runtime.finQueue,
		FinStore:           runtime.finStore,
		PlaybookStore:      runtime.playbookStore,
		SkipCertValidation: opts.HTTP.SkipCertValidation,
		FinStaleAfter:      opts.Fin.StaleAfter,
		TheHive:            opts.TheHive,
	})
	runtime.runs = runs.New(runEngine.NewWalker, runtime.playbookStore, runtime.runState)

	return runtime, nil
}

// initializeStorage opens the SQL store and applies migrations.
func (r *Runtime) initializeStorage(storageCfg config.StorageConfig) error {
	log.Infof("Initializing storage from %s", storageCfg.DatabaseURL)
	store, err := storagesql.New(context.Background(), storageCfg.DatabaseURL)
	if err != nil {
		return err
	}

	r.store = store
	r.playbookStore = store.Playbooks()
	r.finStore = store.Fins()
	return nil
}

// Close releases any resources held by the app.
func (r *Runtime) Close() error {
	if r.finQueue != nil {
		r.finQueue.Close()
	}
	if r.store != nil {
		return r.store.Close(context.Background())
	}
	return nil
}

// Operations returns the use case surface for drivers.
func (r *Runtime) Operations() Operations {
	return Operations{
		Playbooks: r.playbookService,
		Runs:      r.runs,
		Fins:      r.finRegistry,
		Work:      r.finWorkService,
		Manual:    r.manualInbox,
	}
}
