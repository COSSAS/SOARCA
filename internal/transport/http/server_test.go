package httptransport

import (
	"os"
	"path/filepath"
	"testing"

	"soarca/internal/config"
	orchestrator "soarca/internal/orchestrator"
)

func testOperations(t *testing.T) orchestrator.Operations {
	t.Helper()

	app, err := orchestrator.New(orchestrator.Options{
		Storage:  config.StorageConfig{DatabaseURL: "sqlite://:memory:"},
		RunState: config.RunStateConfig{MaxRuns: 2},
	})
	if err != nil {
		t.Fatalf("orchestrator.New() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Fatalf("orchestrator.Close() returned error: %v", err)
		}
	})
	return app.Operations()
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
