package config

import (
	"errors"
	"reflect"
	"time"

	"soarca/internal/logger"

	"github.com/spf13/viper"
)

var log *logger.Log

// DefaultDatabaseURL keeps a development database in the working directory so a
// fresh checkout runs without any configuration.
const DefaultDatabaseURL = "sqlite://soarca.db"

func init() {
	log = logger.Logger(reflect.TypeOf(Config{}).PkgPath(), logger.Info, "", logger.Json)
}

type Config struct {
	Server  ServerConfig
	Storage StorageConfig
	Fin     FinConfig
	RunState   RunStateConfig
	HTTP    HTTPConfig
	Auth    AuthConfig
	TheHive TheHiveConfig
	CORS    CORSConfig
}

type ServerConfig struct {
	Port      string
	EnableTLS bool
	CertFile  string
	CertKey   string
}

type StorageConfig struct {
	// DatabaseURL selects the backend by scheme, e.g. sqlite://soarca.db or
	// postgres://user:pass@host:5432/soarca.
	DatabaseURL string
}

type FinConfig struct {
	PollIntervalSeconds    int
	LongPollTimeoutSeconds int
	JobLeaseSeconds        int
	StaleAfterMultiplier   int
	StaleAfter             time.Duration // Computed from LongPollTimeoutSeconds * StaleAfterMultiplier
	RegistrationToken      string
}

type RunStateConfig struct {
	MaxRuns int
}

type HTTPConfig struct {
	SkipCertValidation bool
}

type AuthConfig struct {
	Enabled bool
}

type TheHiveConfig struct {
	Activate          bool
	APIToken          string
	APIBaseURL        string
	AllowInsecure     bool
	EnableReporter    bool
	EnableCaseManager bool
}

type CORSConfig struct {
	AllowedOrigins string
}

// Load reads configuration from environment variables using Viper.
func Load() (Config, error) {
	v := viper.New()

	// Bind environment variables. Viper's default replacer is a no-op, which
	// keeps env keys exact; passing nil here would panic in getEnv.
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENABLE_TLS", false)
	v.SetDefault("CERT_FILE", "./certs/server.crt")
	v.SetDefault("CERT_KEY_FILE", "./certs/server.key")
	v.SetDefault("DATABASE_URL", DefaultDatabaseURL)
	v.SetDefault("FIN_POLL_INTERVAL_SECONDS", 5)
	v.SetDefault("FIN_LONG_POLL_TIMEOUT_SECONDS", 25)
	v.SetDefault("FIN_JOB_LEASE_SECONDS", 60)
	v.SetDefault("FIN_STALE_AFTER_MULTIPLIER", 2)
	v.SetDefault("FIN_REGISTRATION_TOKEN", "")
	v.SetDefault("MAX_RUNS", 10)
	v.SetDefault("HTTP_SKIP_CERT_VALIDATION", false)
	v.SetDefault("AUTH_ENABLED", false)
	v.SetDefault("THEHIVE_ACTIVATE", false)
	v.SetDefault("THEHIVE_ALLOW_INSECURE", true)
	v.SetDefault("THEHIVE_REPORTER", false)
	v.SetDefault("SOARCA_ALLOWED_ORIGINS", "*")

	cfg := Config{
		Server: ServerConfig{
			Port:      v.GetString("PORT"),
			EnableTLS: v.GetBool("ENABLE_TLS"),
			CertFile:  v.GetString("CERT_FILE"),
			CertKey:   v.GetString("CERT_KEY_FILE"),
		},
		Storage: StorageConfig{
			DatabaseURL: v.GetString("DATABASE_URL"),
		},
		Fin: FinConfig{
			PollIntervalSeconds:    v.GetInt("FIN_POLL_INTERVAL_SECONDS"),
			LongPollTimeoutSeconds: v.GetInt("FIN_LONG_POLL_TIMEOUT_SECONDS"),
			JobLeaseSeconds:        v.GetInt("FIN_JOB_LEASE_SECONDS"),
			StaleAfterMultiplier:   v.GetInt("FIN_STALE_AFTER_MULTIPLIER"),
			RegistrationToken:      v.GetString("FIN_REGISTRATION_TOKEN"),
		},
		RunState: RunStateConfig{
			MaxRuns: v.GetInt("MAX_RUNS"),
		},
		HTTP: HTTPConfig{
			SkipCertValidation: v.GetBool("HTTP_SKIP_CERT_VALIDATION"),
		},
		Auth: AuthConfig{
			Enabled: v.GetBool("AUTH_ENABLED"),
		},
		TheHive: TheHiveConfig{
			Activate:          v.GetBool("THEHIVE_ACTIVATE"),
			APIToken:          v.GetString("THEHIVE_API_TOKEN"),
			APIBaseURL:        v.GetString("THEHIVE_API_BASE_URL"),
			AllowInsecure:     v.GetBool("THEHIVE_ALLOW_INSECURE"),
			EnableReporter:    v.GetBool("THEHIVE_REPORTER"),
			EnableCaseManager: v.GetBool("THEHIVE_REPORTER"), // Alias to reporter
		},
		CORS: CORSConfig{
			AllowedOrigins: v.GetString("SOARCA_ALLOWED_ORIGINS"),
		},
	}
	cfg.Fin.StaleAfter = time.Duration(cfg.Fin.LongPollTimeoutSeconds*cfg.Fin.StaleAfterMultiplier) * time.Second

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks required configuration values.
func (c Config) Validate() error {
	if c.Storage.DatabaseURL == "" {
		return errors.New("DATABASE_URL must be set")
	}
	if c.Fin.LongPollTimeoutSeconds <= 0 {
		return errors.New("FIN_LONG_POLL_TIMEOUT_SECONDS must be greater than 0")
	}
	if c.Fin.StaleAfterMultiplier <= 0 {
		return errors.New("FIN_STALE_AFTER_MULTIPLIER must be greater than 0")
	}
	if c.Fin.StaleAfter <= 0 {
		return errors.New("FIN staleAfter must be greater than 0")
	}
	if c.TheHive.Activate && (c.TheHive.APIToken == "" || c.TheHive.APIBaseURL == "") {
		return errors.New("THEHIVE_ACTIVATE is enabled but THEHIVE_API_TOKEN or THEHIVE_API_BASE_URL are not set")
	}
	return nil
}

// LogSettings logs the current configuration (sanitizing secrets).
func (c Config) LogSettings() {
	log.Info("Configuration loaded:")
	log.Infof("  Server Port: %s", c.Server.Port)
	log.Infof("  Server TLS: %v", c.Server.EnableTLS)
	log.Infof("  Storage: %s", c.Storage.DatabaseURL)
	log.Infof("  RunState Size: %d", c.RunState.MaxRuns)
	log.Infof("  Auth Enabled: %v", c.Auth.Enabled)
	log.Infof("  TheHive Activated: %v", c.TheHive.Activate)
}

func storageType(cfg StorageConfig) string {
	return cfg.DatabaseURL
}
