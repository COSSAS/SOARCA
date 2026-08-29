package httptransport

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"soarca/internal/bootstrap"
	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	"soarca/pkg/api"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Server owns the HTTP transport wiring for a runtime app.
// It is responsible for route registration, middleware, and listener startup only.
// All application logic lives in the bootstrap container.
type Server struct {
	container *bootstrap.Container
}

// New creates a new HTTP server adapter with a bootstrapped service container.
func New(runtime *appruntime.Runtime, cfg config.Config) (*Server, error) {
	container, err := bootstrap.New(runtime, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap services: %w", err)
	}
	return &Server{container: container}, nil
}

// SetupServer initializes the Gin engine with all routes and middleware.
func (s *Server) SetupServer() (*gin.Engine, error) {
	engine := gin.New()

	log.Info("Log level is info")
	log.Debug("Log level is debug")
	log.Trace("Log level is trace")

	origins := strings.Split(strings.ReplaceAll(s.container.TransportOptions.CORS.AllowedOrigins, " ", ""), ",")
	api.Cors(engine, origins)

	api.FinPublic(engine, s.container.FinHandler)

	if err := s.setupAuthMiddleware(engine); err != nil {
		return nil, fmt.Errorf("failed to setup auth middleware: %w", err)
	}

	api.TriggerRoutes(engine, api.NewTriggerHandler(s.container.TriggerService))
	api.StatusRoutes(engine)
	api.PlaybookRoutesWithService(engine, s.container.PlaybookService)
	api.ReporterRoutesWithService(engine, s.container.ReporterService)
	api.ManualRoutes(engine, api.NewManualHandler(s.container.ManualInbox))
	api.FinAdmin(engine, s.container.FinHandler)
	api.Logging(engine)
	api.Swagger(engine)

	return engine, nil
}

// setupAuthMiddleware configures authentication if enabled.
func (s *Server) setupAuthMiddleware(engine *gin.Engine) error {
	if !s.container.TransportOptions.Auth.Enabled {
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
	if s.container.TransportOptions.Server.EnableTLS {
		if err := validateCertificates(s.container.TransportOptions.Server.CertFile, s.container.TransportOptions.Server.CertKey); err != nil {
			return err
		}
		log.Infof("Starting HTTPS server on port %s", s.container.TransportOptions.Server.Port)
		return engine.RunTLS(":"+s.container.TransportOptions.Server.Port, s.container.TransportOptions.Server.CertFile, s.container.TransportOptions.Server.CertKey)
	}

	log.Infof("Starting HTTP server on port %s", s.container.TransportOptions.Server.Port)
	return engine.Run(":" + s.container.TransportOptions.Server.Port)
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
