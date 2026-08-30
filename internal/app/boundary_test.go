package app

import (
	"testing"

	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"
)

// TestRuntimeBoundary verifies the runtime exposes its use cases as behaviour
// only. Operations must carry no infrastructure: if a store, cache, queue or
// walker factory ever appears here, transport can reach through it again.
func TestRuntimeBoundary(t *testing.T) {
	runtime, err := appruntime.New(mockRuntimeOptions())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ops := runtime.Operations()

	if ops.Playbooks == nil {
		t.Error("Operations.Playbooks is nil")
	}
	if ops.Executions == nil {
		t.Error("Operations.Executions is nil")
	}
	if ops.Fins == nil {
		t.Error("Operations.Fins is nil")
	}
	if ops.Work == nil {
		t.Error("Operations.Work is nil")
	}
	if ops.Manual == nil {
		t.Error("Operations.Manual is nil")
	}

	// The runtime must not hand out infrastructure. If any of these compile,
	// the boundary has been violated:
	//   _ = runtime.GetPlaybookStore()
	//   _ = runtime.GetCache()
	//   _ = runtime.GetFinQueue()
}

// TestTransportBoundary verifies the transport layer only does route
// registration and startup, driven purely by Operations.
func TestTransportBoundary(t *testing.T) {
	runtime, err := appruntime.New(mockRuntimeOptions())
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	server := httptransport.New(runtime.Operations(), mockTransportOptions())

	engine, err := server.SetupServer()
	if err != nil {
		t.Fatalf("Failed to setup server: %v", err)
	}
	if engine == nil {
		t.Error("SetupServer returned nil engine")
	}
	if got := len(engine.Routes()); got == 0 {
		t.Error("SetupServer registered no routes")
	}

	// The transport must not expose runtime internals. If any of these
	// compile, the boundary has been violated:
	//   _ = server.GetPlaybookStore()
	//   _ = server.NewWalker()
}

// TestConfigBoundary verifies config is split by ownership: the transport is
// handed only what HTTP actually needs, never the whole application config.
func TestConfigBoundary(t *testing.T) {
	runtimeOpts := mockRuntimeOptions()

	if runtimeOpts.Cache.MaxExecutions != 5 {
		t.Error("Runtime should own Cache config")
	}
	if runtimeOpts.TheHive.Activate != false {
		t.Error("Runtime should own TheHive config")
	}
	if runtimeOpts.HTTP.SkipCertValidation != false {
		t.Error("Runtime should own outbound HTTP config")
	}

	transportOpts := mockTransportOptions()

	if transportOpts.Server.Port != "8080" {
		t.Error("Transport should have Server config with port")
	}
	if transportOpts.Fin.RegistrationToken != "test-token" {
		t.Error("Transport should have Fin config")
	}
	if transportOpts.Auth.Enabled != false {
		t.Error("Transport should have Auth config")
	}
	if transportOpts.CORS.AllowedOrigins != "*" {
		t.Error("Transport should have CORS config")
	}

	// Transport options must not carry orchestrator concerns. If any of these
	// compile, the split has regressed:
	//   _ = transportOpts.Storage
	//   _ = transportOpts.TheHive
}

// Helper functions for testing

func mockStorageConfig() config.StorageConfig {
	return config.StorageConfig{
		DatabaseURL: "sqlite://:memory:",
	}
}

func mockCacheConfig() config.CacheConfig {
	return config.CacheConfig{
		MaxExecutions: 5,
	}
}

func mockFinConfig() config.FinConfig {
	return config.FinConfig{
		RegistrationToken:      "test-token",
		PollIntervalSeconds:    5,
		LongPollTimeoutSeconds: 30,
		JobLeaseSeconds:        60,
		StaleAfter:             300,
	}
}

func mockRuntimeOptions() appruntime.Options {
	return appruntime.Options{
		Storage: mockStorageConfig(),
		Cache:   mockCacheConfig(),
		Fin:     mockFinConfig(),
		HTTP: config.HTTPConfig{
			SkipCertValidation: false,
		},
		TheHive: config.TheHiveConfig{
			Activate: false,
		},
	}
}

func mockTransportOptions() httptransport.Options {
	return httptransport.Options{
		Server: config.ServerConfig{
			Port:      "8080",
			EnableTLS: false,
		},
		Fin:  mockFinConfig(),
		Auth: config.AuthConfig{Enabled: false},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
}
