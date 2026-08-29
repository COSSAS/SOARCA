package controller

import (
	"testing"

	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"
)

// TestRuntimeBoundary verifies that runtime properly exposes dependencies through narrow accessors.
func TestRuntimeBoundary(t *testing.T) {
	runtime, err := appruntime.New(appruntime.Options{
		Storage: mockStorageConfig(),
		Cache:   mockCacheConfig(),
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	// Verify all getters return non-nil values
	if runtime.GetPlaybookStore() == nil {
		t.Error("GetPlaybookStore returned nil")
	}
	if runtime.GetFinStore() == nil {
		t.Error("GetFinStore returned nil")
	}
	if runtime.GetCache() == nil {
		t.Error("GetCache returned nil")
	}
	if runtime.GetInteraction() == nil {
		t.Error("GetInteraction returned nil")
	}
	if runtime.GetFinQueue() == nil {
		t.Error("GetFinQueue returned nil")
	}
}

// TestTransportBoundary verifies that transport correctly uses runtime accessors.
func TestTransportBoundary(t *testing.T) {
	runtime, err := appruntime.New(appruntime.Options{
		Storage: mockStorageConfig(),
		Cache:   mockCacheConfig(),
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	server := httptransport.New(runtime, mockTransportOptions())

	// Verify transport can call SetupServer successfully
	engine, err := server.SetupServer()
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}
	if engine == nil {
		t.Error("SetupServer returned nil engine")
	}

	// Verify transport can access playbook store through interface
	store := server.GetPlaybookStore()
	if store == nil {
		t.Error("GetPlaybookStore returned nil")
	}

	// Verify transport can create decomposer
	decomposer := server.NewDecomposer()
	if decomposer == nil {
		t.Error("NewDecomposer returned nil")
	}
}

// TestConfigBoundary verifies that config is properly split between layers.
func TestConfigBoundary(t *testing.T) {
	// Verify runtime options are narrow (only storage/cache)
	runtimeOpts := appruntime.Options{
		Storage: mockStorageConfig(),
		Cache:   mockCacheConfig(),
	}

	// Runtime should have cache max executions set
	if runtimeOpts.Cache.MaxExecutions != 5 {
		t.Error("Runtime should have Cache config")
	}

	// Verify transport options include HTTP-specific config
	transportOpts := mockTransportOptions()

	// Transport should have server port
	if transportOpts.Server.Port != "8080" {
		t.Error("Transport should have Server config with port")
	}

	// Transport should have FIN config
	if transportOpts.Fin.RegistrationToken != "test-token" {
		t.Error("Transport should have Fin config")
	}

	// Transport should have HTTP config
	if transportOpts.HTTP.SkipCertValidation != false {
		t.Error("Transport should have HTTP config")
	}

	// Transport should have Auth config
	if transportOpts.Auth.Enabled != false {
		t.Error("Transport should have Auth config")
	}
}

// Helper functions for testing

func mockStorageConfig() config.StorageConfig {
	return config.StorageConfig{
		UseDatabase: false,
		MongoDBURI:  "",
	}
}

func mockCacheConfig() config.CacheConfig {
	return config.CacheConfig{
		MaxExecutions: 5,
	}
}

func mockTransportOptions() httptransport.Options {
	return httptransport.Options{
		Server: config.ServerConfig{
			Port:      "8080",
			EnableTLS: false,
		},
		Fin: config.FinConfig{
			RegistrationToken:      "test-token",
			PollIntervalSeconds:    5,
			LongPollTimeoutSeconds: 30,
			JobLeaseSeconds:        60,
			StaleAfter:             300,
		},
		HTTP: config.HTTPConfig{
			SkipCertValidation: false,
		},
		Auth: config.AuthConfig{
			Enabled: false,
		},
		TheHive: config.TheHiveConfig{
			Activate: false,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: "*",
		},
	}
}
