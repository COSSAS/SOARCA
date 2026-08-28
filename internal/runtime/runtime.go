package runtime

import (
	"context"
	"fmt"
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	"soarca/internal/storage"
	storagememory "soarca/internal/storage/memory"
	storagemongo "soarca/internal/storage/mongodb"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/reporting/reporter/downstream_reporter/cache"
	timeutil "soarca/pkg/utils/time"
)

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Runtime{}).PkgPath(), logger.Info, "", logger.Json)
}

// Options contains only the configuration needed to build the runtime.
type Options struct {
	Storage config.StorageConfig
	Cache   config.CacheConfig
}

// Runtime holds all wired application dependencies for the core SOAR orchestrator.
type Runtime struct {
	PlaybookStore storage.PlaybookStore
	FinStore      storage.FinStore
	Cache         *cache.Cache
	Interaction   *interaction.InteractionController
	FinQueue      *queue.Queue
}

// New creates and initializes all application dependencies.
func New(opts Options) (*Runtime, error) {
	runtime := &Runtime{}

	if err := runtime.initializeStorage(opts.Storage); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	runtime.Cache = cache.New(&timeutil.Time{}, opts.Cache.MaxExecutions)
	runtime.Interaction = interaction.New([]interaction.IInteractionIntegrationNotifier{})
	runtime.FinQueue = queue.New()

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
