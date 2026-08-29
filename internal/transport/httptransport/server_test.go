package httptransport

import (
	"os"
	"path/filepath"
	"testing"

	"soarca/internal/config"
	appruntime "soarca/internal/runtime"
)

func testOperations(t *testing.T) appruntime.Operations {
	t.Helper()

	runtime, err := appruntime.New(appruntime.Options{
		Storage: config.StorageConfig{UseDatabase: false},
		Cache:   config.CacheConfig{MaxExecutions: 2},
	})
	if err != nil {
		t.Fatalf("appruntime.New() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Fatalf("runtime.Close() returned error: %v", err)
		}
	})
	return runtime.Operations()
}

func TestSetupServerReturnsEngineWithAuthDisabled(t *testing.T) {
	opts := Options{
		Server: config.ServerConfig{Port: "0"},
		Fin:    config.FinConfig{},
		Auth:   config.AuthConfig{Enabled: false},
		CORS:   config.CORSConfig{AllowedOrigins: "*"},
	}

	server := New(testOperations(t), opts)

	engine, err := server.SetupServer()
	if err != nil {
		t.Fatalf("SetupServer() returned error: %v", err)
	}
	if engine == nil {
		t.Fatal("SetupServer() returned nil engine")
	}
	if got := len(engine.Routes()); got == 0 {
		t.Fatal("expected routes to be registered")
	}
}

func TestValidateCertificates(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "server.crt")
	keyFile := filepath.Join(dir, "server.key")

	if err := os.WriteFile(certFile, []byte("cert"), 0o600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("key"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	if err := validateCertificates(certFile, keyFile); err != nil {
		t.Fatalf("validateCertificates() returned error: %v", err)
	}

	if err := validateCertificates(filepath.Join(dir, "missing.crt"), keyFile); err == nil {
		t.Fatal("expected missing certificate file to fail")
	}
}
