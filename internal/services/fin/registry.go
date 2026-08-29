package fin

import (
	"context"
	"reflect"
	"time"

	"soarca/internal/logger"
	"soarca/internal/storage"
	"soarca/pkg/core/capability/fin/token"
	"soarca/pkg/models/fin"
	"soarca/pkg/utils/guid"
)

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(struct{}{}).PkgPath(), logger.Info, "", logger.Json)
}

// Registry implements the services.FinRegistry interface.
type Registry struct {
	store  storage.FinStore
	config RegistryConfig
	guid   guid.IGuid
}

// RegistryConfig holds configuration for the FIN registry service.
type RegistryConfig struct {
	RegistrationToken string
	StaleAfter        time.Duration
}

// NewRegistry creates a new FinRegistry service.
func NewRegistry(store storage.FinStore, config RegistryConfig, guid guid.IGuid) *Registry {
	return &Registry{
		store:  store,
		config: config,
		guid:   guid,
	}
}

// RegisterFin registers a new FIN with the given capabilities.
func (r *Registry) RegisterFin(ctx context.Context, req fin.RegisterRequest) (finID string, finToken string, err error) {
	if r.config.RegistrationToken == "" {
		return "", "", fin.ErrRegistrationDisabled{}
	}
	if !token.Equal(req.RegistrationToken, r.config.RegistrationToken) {
		return "", "", fin.ErrRegistrationTokenInvalid{}
	}
	if len(req.Capabilities) == 0 {
		return "", "", fin.ErrNoCapabilities{}
	}
	for _, capability := range req.Capabilities {
		if capability.Type == "" {
			return "", "", fin.ErrCapabilityTypeEmpty{}
		}
	}

	finToken, err = token.Generate()
	if err != nil {
		return "", "", err
	}

	record := fin.Record{
		FinId:           r.guid.New().String(),
		FinTokenHash:    token.Hash(finToken),
		DisplayName:     req.DisplayName,
		ProtocolVersion: req.ProtocolVersion,
		Capabilities:    req.Capabilities,
		RegisteredAt:    time.Now(),
		LastSeen:        time.Now(),
	}

	if err := r.store.Create(ctx, record); err != nil {
		return "", "", err
	}

	log.Info("registered fin ", record.FinId, " (", record.DisplayName, ") with capabilities ", capabilityTypes(record.Capabilities))
	return record.FinId, finToken, nil
}

// UnregisterFin unregisters a FIN by its token.
func (r *Registry) UnregisterFin(ctx context.Context, finToken string) error {
	record, err := r.store.GetByTokenHash(ctx, token.Hash(finToken))
	if err != nil {
		return err
	}
	return r.store.Delete(ctx, record.FinId)
}

// ListFins returns all registered FINs with staleness information.
func (r *Registry) ListFins(ctx context.Context) ([]fin.Record, error) {
	records, err := r.store.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i].Stale = r.isStale(records[i])
	}
	return records, nil
}

// GetFin retrieves a specific FIN record by ID.
func (r *Registry) GetFin(ctx context.Context, finID string) (fin.Record, error) {
	record, err := r.store.Get(ctx, finID)
	if err != nil {
		return fin.Record{}, err
	}
	record.Stale = r.isStale(record)
	return record, nil
}

// DeleteFin removes a FIN record by ID (admin use).
func (r *Registry) DeleteFin(ctx context.Context, finID string) error {
	return r.store.Delete(ctx, finID)
}

// ValidateToken checks if a token is valid and returns the associated FIN record.
// Returns an error if the token is not found or invalid.
func (r *Registry) ValidateToken(ctx context.Context, finToken string) (string, error) {
	record, err := r.store.GetByTokenHash(ctx, token.Hash(finToken))
	if err != nil {
		return "", err
	}
	return record.FinId, nil
}

// isStale checks if a FIN has exceeded the staleness threshold.
func (r *Registry) isStale(record fin.Record) bool {
	return time.Since(record.LastSeen) > r.config.StaleAfter
}

// capabilityTypes extracts capability type strings from capability records.
func capabilityTypes(capabilities []fin.Capability) []string {
	types := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		types = append(types, capability.Type)
	}
	return types
}
