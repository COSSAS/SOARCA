package controller

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
	httptransport "soarca/internal/transport/httptransport"
)

type fakeTransport struct {
	setupCalled bool
	runCalled   bool
	cfg         httptransport.Options
}

func (f *fakeTransport) SetupServer() (*gin.Engine, error) {
	f.setupCalled = true
	return gin.New(), nil
}

func (f *fakeTransport) RunServer(*gin.Engine) error {
	f.runCalled = true
	return nil
}

func TestInitializeWiresRuntimeAndTransport(t *testing.T) {
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
	newTransport = func(runtime *appruntime.Runtime, opts httptransport.Options) transport {
		gotTransportOpts = opts
		return fake
	}

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() returned error: %v", err)
	}

	if gotRuntimeOpts.Storage != cfg.Storage {
		t.Fatalf("runtime options storage mismatch: %#v", gotRuntimeOpts.Storage)
	}
	if gotRuntimeOpts.Cache != cfg.Cache {
		t.Fatalf("runtime options cache mismatch: %#v", gotRuntimeOpts.Cache)
	}
	if gotTransportOpts.Server != cfg.Server {
		t.Fatalf("transport options server mismatch: %#v", gotTransportOpts.Server)
	}
	if gotTransportOpts.Fin != cfg.Fin {
		t.Fatalf("transport options fin mismatch: %#v", gotTransportOpts.Fin)
	}
	if gotTransportOpts.HTTP != cfg.HTTP {
		t.Fatalf("transport options http mismatch: %#v", gotTransportOpts.HTTP)
	}
	if gotTransportOpts.Auth != cfg.Auth {
		t.Fatalf("transport options auth mismatch: %#v", gotTransportOpts.Auth)
	}
	if gotTransportOpts.TheHive != cfg.TheHive {
		t.Fatalf("transport options thehive mismatch: %#v", gotTransportOpts.TheHive)
	}
	if gotTransportOpts.CORS != cfg.CORS {
		t.Fatalf("transport options cors mismatch: %#v", gotTransportOpts.CORS)
	}
	if !fake.setupCalled {
		t.Fatal("expected SetupServer to be called")
	}
	if !fake.runCalled {
		t.Fatal("expected RunServer to be called")
	}
}

func TestInitializeReturnsConfigLoadError(t *testing.T) {
	origLoadConfig := loadConfig
	t.Cleanup(func() {
		loadConfig = origLoadConfig
	})

	loadConfig = func() (config.Config, error) {
		return config.Config{}, errors.New("boom")
	}

	if err := Initialize(); err == nil {
		t.Fatal("expected Initialize to return an error")
	}
}
