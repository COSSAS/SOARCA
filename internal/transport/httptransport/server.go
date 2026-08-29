package httptransport

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	"soarca/pkg/api"
	finapi "soarca/pkg/api/fin"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Options is the configuration the HTTP transport needs. It is deliberately
// narrower than the application config: anything the orchestrator owns
// (storage, TheHive, outbound TLS) does not belong here.
type Options struct {
	Server config.ServerConfig
	Fin    config.FinConfig
	Auth   config.AuthConfig
	CORS   config.CORSConfig
}

// Server owns the HTTP transport wiring: route registration, middleware and
// listener startup. It holds the orchestrator's use case surface and nothing else.
type Server struct {
	ops        appruntime.Operations
	opts       Options
	finHandler *finapi.FinHandler
}

// New creates an HTTP server adapter over the given operations.
func New(ops appruntime.Operations, opts Options) *Server {
	finHandler := finapi.NewFinHandler(
		ops.Fins,
		ops.Work,
		finapi.Config{
			RegistrationToken:      opts.Fin.RegistrationToken,
			PollIntervalSeconds:    opts.Fin.PollIntervalSeconds,
			LongPollTimeoutSeconds: opts.Fin.LongPollTimeoutSeconds,
			JobLeaseSeconds:        opts.Fin.JobLeaseSeconds,
			StaleAfter:             opts.Fin.StaleAfter,
		},
	)

	return &Server{ops: ops, opts: opts, finHandler: finHandler}
}

// SetupServer initializes the Gin engine with all routes and middleware.
func (s *Server) SetupServer() (*gin.Engine, error) {
	engine := gin.New()

	log.Info("Log level is info")
	log.Debug("Log level is debug")
	log.Trace("Log level is trace")

	origins := strings.Split(strings.ReplaceAll(s.opts.CORS.AllowedOrigins, " ", ""), ",")
	api.Cors(engine, origins)

	api.FinPublic(engine, s.finHandler)

	if err := s.setupAuthMiddleware(engine); err != nil {
		return nil, fmt.Errorf("failed to setup auth middleware: %w", err)
	}

	api.TriggerRoutes(engine, api.NewTriggerHandler(s.ops.Executions))
	api.StatusRoutes(engine)
	api.PlaybookRoutesWithService(engine, s.ops.Playbooks)
	api.ReporterRoutesWithService(engine, s.ops.Executions)
	api.ManualRoutes(engine, api.NewManualHandler(s.ops.Manual))
	api.FinAdmin(engine, s.finHandler)
	api.Logging(engine)
	api.Swagger(engine)

	return engine, nil
}

// setupAuthMiddleware configures authentication if enabled.
func (s *Server) setupAuthMiddleware(engine *gin.Engine) error {
	if !s.opts.Auth.Enabled {
		return nil
	}

	log.Info("Enabling authentication middleware")
	auth, err := gauth.New(gauth.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize authenticator: %w", err)
	}
	engine.Use(auth.LoadAuthContext())
	engine.Use(auth.Middleware([]string{"soarca_admin"}))
	return nil
}

// RunServer starts the HTTP server on the configured port.
func (s *Server) RunServer(engine *gin.Engine) error {
	if s.opts.Server.EnableTLS {
		if err := validateCertificates(s.opts.Server.CertFile, s.opts.Server.CertKey); err != nil {
			return err
		}
		log.Infof("Starting HTTPS server on port %s", s.opts.Server.Port)
		return engine.RunTLS(":"+s.opts.Server.Port, s.opts.Server.CertFile, s.opts.Server.CertKey)
	}

	log.Infof("Starting HTTP server on port %s", s.opts.Server.Port)
	return engine.Run(":" + s.opts.Server.Port)
}

// validateCertificates checks that TLS certificate files exist.
func validateCertificates(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("key file not found: %s", keyFile)
	}
	return nil
}
