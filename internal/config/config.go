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
	for _, origin := range dynamic.AllowedOrigins {
		if origin == "*" {
			if len(dynamic.AllowedOrigins) > 1 {
				return errors.New("allowed_origins: \"*\" cannot be combined with other origins")
			}
			continue
		}
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("allowed_origins: %q must start with http:// or https://", origin)
		}
	}
	return nil
}

// clone returns a deep copy. A plain struct copy would share the
// AllowedOrigins backing array with the original.
func (dynamic Dynamic) clone() Dynamic {
	dynamic.AllowedOrigins = append([]string(nil), dynamic.AllowedOrigins...)
	return dynamic
}

type Config struct {
	Static Static

	current   atomic.Pointer[Dynamic]
	mutex     sync.Mutex
	listeners []func(before Dynamic, after Dynamic)
}

// Load parses the process environment. Load any .env file before calling it.
func Load() (*Config, error) {
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

	configuration := &Config{Static: static}
	configuration.current.Store(&dynamic)
	return configuration, nil
}

// Dynamic returns a copy of the current runtime settings. Call it at the
// point of use; don't store the result.
func (configuration *Config) Dynamic() Dynamic {
	return configuration.current.Load().clone()
}

// OnChange registers a listener for components that hold their own state,
// like logrus loggers. Listeners run after the new settings are current.
func (configuration *Config) OnChange(listener func(before Dynamic, after Dynamic)) {
	configuration.mutex.Lock()
	defer configuration.mutex.Unlock()
	configuration.listeners = append(configuration.listeners, listener)
}

// DynamicPatch describes a change to Dynamic. Nil fields are left as they
// are, so a JSON body only needs the fields it changes.
type DynamicPatch struct {
	LogLevel               *string   `json:"log_level"`
	HttpSkipCertValidation *bool     `json:"http_skip_cert_validation"`
	ValidationSchemaUrl    *string   `json:"validation_schema_url"`
	AllowedOrigins         *[]string `json:"allowed_origins"`
}

func (patch DynamicPatch) applyTo(dynamic *Dynamic) {
	if patch.LogLevel != nil {
		dynamic.LogLevel = *patch.LogLevel
	}
	if patch.HttpSkipCertValidation != nil {
		dynamic.HttpSkipCertValidation = *patch.HttpSkipCertValidation
	}
	if patch.ValidationSchemaUrl != nil {
		dynamic.ValidationSchemaUrl = *patch.ValidationSchemaUrl
	}
	if patch.AllowedOrigins != nil {
		// Copy so the caller's slice never becomes part of the live settings.
		dynamic.AllowedOrigins = append([]string(nil), (*patch.AllowedOrigins)...)
	}
}

// Update applies patch to the current settings, validates the result and only
// then makes it current. Reading and writing happen under one lock, so two
// concurrent updates to different fields both survive. On error nothing
// changes. It returns the settings that are now current.
func (configuration *Config) Update(patch DynamicPatch) (Dynamic, error) {
	configuration.mutex.Lock()
	defer configuration.mutex.Unlock()

	before := configuration.Dynamic()
	after := before.clone()
	patch.applyTo(&after)
	if err := after.Validate(); err != nil {
		return before, err
	}
	configuration.current.Store(&after)

	// after is now the live value; listeners get copies so they can't change it.
	for _, listener := range configuration.listeners {
		listener(before.clone(), after.clone())
	}
	return after.clone(), nil
}
