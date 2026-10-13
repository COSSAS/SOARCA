// Package config loads SOARCA's settings from .env and the environment once at
// startup. Static holds settings that are wired into long-lived objects
// (HTTP server, Mongo client, MQTT, TheHive connector, worker pools) and only
// take effect on restart. Dynamic holds the few settings SOARCA reads at the
// point of use, so they can be swapped at runtime, e.g. from an API handler.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Static struct {
	Server   Server
	Database Database
	Log      LogStatic
	Fins     Fins
	TheHive  TheHive
	Auth     Auth
	SshKms   SshKms

	// MAX_EXECUTIONS sizes the execution cache when it is created.
	MaxExecutions int `env:"MAX_EXECUTIONS" envDefault:"10"`
	// MAX_REPORTERS sets the reporter worker count when the reporter is created.
	MaxReporters int `env:"MAX_REPORTERS" envDefault:"10"`
}

type Server struct {
	Port        int    `env:"PORT" envDefault:"8080"`
	EnableTls   bool   `env:"ENABLE_TLS" envDefault:"false"`
	CertFile    string `env:"CERT_FILE" envDefault:"./certs/server.crt"`
	CertKeyFile string `env:"CERT_KEY_FILE" envDefault:"./certs/server.key"`
	// Read by gin itself during package init; kept here so it is documented
	// and can be applied explicitly with gin.SetMode.
	GinMode string `env:"GIN_MODE" envDefault:"release"`
}

type Database struct {
	Enabled  bool   `env:"DATABASE" envDefault:"false"`
	Uri      string `env:"MONGODB_URI"`
	Username string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
}

// LogStatic covers logger settings that are applied when each logger is built.
// LOG_GLOBAL_LEVEL lives in Dynamic.
type LogStatic struct {
	Mode     string `env:"LOG_MODE" envDefault:"production"`
	FilePath string `env:"LOG_FILE_PATH" envDefault:""`
	Format   string `env:"LOG_FORMAT" envDefault:"json"`
}

type Fins struct {
	Enabled    bool   `env:"ENABLE_FINS" envDefault:"false"`
	MqttBroker string `env:"MQTT_BROKER" envDefault:"localhost"`
	MqttPort   int    `env:"MQTT_PORT" envDefault:"1883"`
}

type TheHive struct {
	Activate      bool   `env:"THEHIVE_ACTIVATE" envDefault:"false"`
	ApiToken      string `env:"THEHIVE_API_TOKEN"`
	ApiBaseUrl    string `env:"THEHIVE_API_BASE_URL"`
	AllowInsecure bool   `env:"THEHIVE_ALLOW_INSECURE" envDefault:"true"`
	Reporter      bool   `env:"THEHIVE_REPORTER" envDefault:"false"`
}

// Auth mirrors what github.com/COSSAS/gauth DefaultConfig() reads.
type Auth struct {
	Enabled       bool   `env:"AUTH_ENABLED" envDefault:"false"`
	Issuer        string `env:"OIDC_ISSUER"`
	ClientId      string `env:"OIDC_CLIENT_ID"`
	SkipTlsVerify bool   `env:"OIDC_SKIP_TLS_VERIFY" envDefault:"false"`
}

type SshKms struct {
	Enabled bool   `env:"ENABLE_SSH_KMS" envDefault:"false"`
	Dir     string `env:"SSH_KMS_DIR"`
}

// Dynamic is everything that may be changed while SOARCA runs. Secrets never
// go here: this struct is what an API handler reads and writes.
type Dynamic struct {
	// Applied with logrus SetLevel on live loggers when changed.
	LogLevel string `env:"LOG_GLOBAL_LEVEL" envDefault:"info" json:"log_level"`
	// Read every time a decomposer is built (per trigger/playbook action).
	HttpSkipCertValidation bool `env:"HTTP_SKIP_CERT_VALIDATION" envDefault:"false" json:"http_skip_cert_validation"`
	// Read on every CACAO v2 validation. Empty means use the embedded schema.
	ValidationSchemaUrl string `env:"VALIDATION_SCHEMA_URL" json:"validation_schema_url"`
	// Needs the CORS middleware to read config per request instead of once.
	AllowedOrigins []string `env:"SOARCA_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*" json:"allowed_origins"`
}

var validLogLevels = map[string]bool{
	"panic": true, "fatal": true, "error": true, "warning": true,
	"info": true, "debug": true, "trace": true,
}

func (static Static) Validate() error {
	var errs []error
	if static.Server.Port < 1 || static.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("PORT %d out of range", static.Server.Port))
	}
	database := static.Database
	if database.Enabled && (database.Uri == "" || database.Username == "" || database.Password == "") {
		errs = append(errs, errors.New("DATABASE=true requires MONGODB_URI, DB_USERNAME and DB_PASSWORD"))
	}
	if static.Auth.Enabled && (static.Auth.Issuer == "" || static.Auth.ClientId == "") {
		errs = append(errs, errors.New("AUTH_ENABLED=true requires OIDC_ISSUER and OIDC_CLIENT_ID"))
	}
	if static.MaxExecutions < 1 {
		errs = append(errs, errors.New("MAX_EXECUTIONS must be at least 1"))
	}
	if static.MaxReporters < 1 {
		errs = append(errs, errors.New("MAX_REPORTERS must be at least 1"))
	}
	return errors.Join(errs...)
}

func (dynamic Dynamic) Validate() error {
	if !validLogLevels[strings.ToLower(dynamic.LogLevel)] {
		return fmt.Errorf("unknown log level %q", dynamic.LogLevel)
	}
	if dynamic.ValidationSchemaUrl != "" {
		schemaUrl, err := url.Parse(dynamic.ValidationSchemaUrl)
		if err != nil || (schemaUrl.Scheme != "http" && schemaUrl.Scheme != "https") {
			return errors.New("validation_schema_url must be an http(s) URL")
		}
	}
	if len(dynamic.AllowedOrigins) == 0 {
		return errors.New("allowed_origins cannot be empty")
	}
	return nil
}

func (dynamic Dynamic) clone() Dynamic {
	dynamic.AllowedOrigins = append([]string(nil), dynamic.AllowedOrigins...)
	return dynamic
}

type Config struct {
	Static Static

	dynamic   atomic.Pointer[Dynamic]
	mutex     sync.Mutex
	listeners []func(previous Dynamic, current Dynamic)
}

// Load reads .env (if present) and the environment. Real environment
// variables win over .env, since godotenv never overwrites existing ones.
func Load(envFiles ...string) (*Config, error) {
	_ = godotenv.Load(envFiles...)

	var static Static
	if err := env.Parse(&static); err != nil {
		return nil, err
	}
	if err := static.Validate(); err != nil {
		return nil, err
	}

	var dynamic Dynamic
	if err := env.Parse(&dynamic); err != nil {
		return nil, err
	}
	for index, origin := range dynamic.AllowedOrigins {
		dynamic.AllowedOrigins[index] = strings.TrimSpace(origin)
	}
	if err := dynamic.Validate(); err != nil {
		return nil, err
	}

	config := &Config{Static: static}
	config.dynamic.Store(&dynamic)
	return config, nil
}

// Dynamic returns a copy of the current runtime settings. Call it at the
// point of use; don't store the result.
func (config *Config) Dynamic() Dynamic {
	return config.dynamic.Load().clone()
}

// OnChange registers a callback for components that hold their own state,
// like logrus loggers. Callbacks run after the new value is visible.
func (config *Config) OnChange(listener func(previous Dynamic, current Dynamic)) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	config.listeners = append(config.listeners, listener)
}

// Update applies change to a copy of the current Dynamic config, validates it
// and swaps it in. Nothing changes if change or validation fails.
func (config *Config) Update(change func(*Dynamic) error) error {
	config.mutex.Lock()
	defer config.mutex.Unlock()

	previous := config.dynamic.Load().clone()
	next := previous.clone()
	if err := change(&next); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	config.dynamic.Store(&next)
	for _, listener := range config.listeners {
		listener(previous, next.clone())
	}
	return nil
}
