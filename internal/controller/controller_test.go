package controller

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
)

type fakeTransport struct {
	setupCalled bool
	runCalled   bool
	cfg         config.Config
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
	var gotCfg config.Config
	fake := &fakeTransport{}

	loadConfig = func() (config.Config, error) {
		return cfg, nil
	}
	newRuntime = func(opts appruntime.Options) (*appruntime.Runtime, error) {
		gotRuntimeOpts = opts
		return &appruntime.Runtime{}, nil
	}
	newTransport = func(runtime *appruntime.Runtime, cfg config.Config) transport {
		gotCfg = cfg
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
	if gotCfg.Server != cfg.Server {
		t.Fatalf("config server mismatch: %#v", gotCfg.Server)
	}
	if gotCfg.Fin != cfg.Fin {
		t.Fatalf("config fin mismatch: %#v", gotCfg.Fin)
	}
	if gotCfg.HTTP != cfg.HTTP {
		t.Fatalf("config http mismatch: %#v", gotCfg.HTTP)
	}
	if gotCfg.Auth != cfg.Auth {
		t.Fatalf("config auth mismatch: %#v", gotCfg.Auth)
	}
	if gotCfg.TheHive != cfg.TheHive {
		t.Fatalf("config thehive mismatch: %#v", gotCfg.TheHive)
	}
	if gotCfg.CORS != cfg.CORS {
		t.Fatalf("config cors mismatch: %#v", gotCfg.CORS)
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
