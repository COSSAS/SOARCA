package app

import (
	"errors"
	"testing"

	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"

	"github.com/gin-gonic/gin"
)

type fakeTransport struct {
	setupCalled bool
	runCalled   bool
}

func (f *fakeTransport) SetupServer() (*gin.Engine, error) {
	f.setupCalled = true
	return gin.New(), nil
}

func (f *fakeTransport) RunServer(*gin.Engine) error {
	f.runCalled = true
	return nil
}

func TestRunWiresRuntimeAndTransport(t *testing.T) {
	origLoadConfig := loadConfig
	origNewRuntime := newRuntime
	origNewTransport := newTransport
	t.Cleanup(func() {
		loadConfig = origLoadConfig
		newRuntime = origNewRuntime
		newTransport = origNewTransport
	})

	cfg := config.Config{
		Server: config.ServerConfig{Port: "8080"},
		Storage: config.StorageConfig{
			UseDatabase: false,
		},
		Fin:   config.FinConfig{StaleAfterMultiplier: 1, LongPollTimeoutSeconds: 1},
		Cache: config.CacheConfig{MaxExecutions: 5},
		HTTP:  config.HTTPConfig{SkipCertValidation: true},
		Auth:  config.AuthConfig{Enabled: false},
		TheHive: config.TheHiveConfig{
			Activate: false,
		},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}

	var gotRuntimeOpts appruntime.Options
	var gotTransportOpts httptransport.Options
	fake := &fakeTransport{}

	loadConfig = func() (config.Config, error) {
		return cfg, nil
	}
	newRuntime = func(opts appruntime.Options) (*appruntime.Runtime, error) {
		gotRuntimeOpts = opts
		return &appruntime.Runtime{}, nil
	}
	newTransport = func(ops appruntime.Operations, opts httptransport.Options) transport {
		gotTransportOpts = opts
		return fake
	}

	if err := Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if gotRuntimeOpts.Storage != cfg.Storage {
		t.Fatalf("runtime options storage mismatch: %#v", gotRuntimeOpts.Storage)
	}
	if gotRuntimeOpts.Cache != cfg.Cache {
		t.Fatalf("runtime options cache mismatch: %#v", gotRuntimeOpts.Cache)
	}
	if gotRuntimeOpts.HTTP != cfg.HTTP {
		t.Fatalf("runtime options http mismatch: %#v", gotRuntimeOpts.HTTP)
	}
	if gotRuntimeOpts.TheHive != cfg.TheHive {
		t.Fatalf("runtime options thehive mismatch: %#v", gotRuntimeOpts.TheHive)
	}
	if gotTransportOpts.Server != cfg.Server {
		t.Fatalf("transport server mismatch: %#v", gotTransportOpts.Server)
	}
	if gotTransportOpts.Fin != cfg.Fin {
		t.Fatalf("transport fin mismatch: %#v", gotTransportOpts.Fin)
	}
	if gotTransportOpts.Auth != cfg.Auth {
		t.Fatalf("transport auth mismatch: %#v", gotTransportOpts.Auth)
	}
	if gotTransportOpts.CORS != cfg.CORS {
		t.Fatalf("transport cors mismatch: %#v", gotTransportOpts.CORS)
	}
	if !fake.setupCalled {
		t.Fatal("expected SetupServer to be called")
	}
	if !fake.runCalled {
		t.Fatal("expected RunServer to be called")
	}
}

func TestRunReturnsConfigLoadError(t *testing.T) {
	origLoadConfig := loadConfig
	t.Cleanup(func() {
		loadConfig = origLoadConfig
	})

	loadConfig = func() (config.Config, error) {
		return config.Config{}, errors.New("boom")
	}

	if err := Run(); err == nil {
		t.Fatal("expected Run to return an error")
	}
}
