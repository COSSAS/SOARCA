package httptransport

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"soarca/internal/config"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	execservice "soarca/internal/services/execution"
	"soarca/internal/storage"
	"soarca/pkg/api"
	finapi "soarca/pkg/api/fin"
	"soarca/pkg/core/capability"
	fincap "soarca/pkg/core/capability/fin"
	httpcap "soarca/pkg/core/capability/http"
	manualcap "soarca/pkg/core/capability/manual"
	openc2cap "soarca/pkg/core/capability/openc2"
	pscap "soarca/pkg/core/capability/powershell"
	sshcap "soarca/pkg/core/capability/ssh"
	"soarca/pkg/core/decomposer"
	actionexec "soarca/pkg/core/executors/action"
	condexec "soarca/pkg/core/executors/condition"
	pbactionexec "soarca/pkg/core/executors/playbook_action"
	"soarca/pkg/extensions/soarca/assignment"
	thehivecases "soarca/pkg/integration/thehive/cases"
	thehiveconnector "soarca/pkg/integration/thehive/common/connector"
	thehivereport "soarca/pkg/integration/thehive/reporter"
	"soarca/pkg/reporting/cases"
	"soarca/pkg/reporting/reporter"
	downstreamreport "soarca/pkg/reporting/reporter/downstream_reporter"
	"soarca/pkg/utils/guid"
	httputil "soarca/pkg/utils/http"
	stixcmp "soarca/pkg/utils/stix/expression/comparison"
	timeutil "soarca/pkg/utils/time"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Options contains only the configuration needed by the HTTP transport.
type Options struct {
	Server  config.ServerConfig
	Fin     config.FinConfig
	HTTP    config.HTTPConfig
	Auth    config.AuthConfig
	TheHive config.TheHiveConfig
	CORS    config.CORSConfig
}

// Server owns the HTTP transport wiring for a runtime app.
type Server struct {
	runtime *appruntime.Runtime
	config  Options
}

// New creates a new HTTP server adapter for the given app runtime.
func New(runtime *appruntime.Runtime, cfg Options) *Server {
	return &Server{runtime: runtime, config: cfg}
}

// SetupServer initializes the Gin engine with all routes and middleware.
func (s *Server) SetupServer() (*gin.Engine, error) {
	engine := gin.New()

	log.Info("Log level is info")
	log.Debug("Log level is debug")
	log.Trace("Log level is trace")

	origins := strings.Split(strings.ReplaceAll(s.config.CORS.AllowedOrigins, " ", ""), ",")
	api.Cors(engine, origins)

	finHandler := s.newFinHandler()
	api.FinPublic(engine, finHandler)

	if err := s.setupAuthMiddleware(engine); err != nil {
		return nil, fmt.Errorf("failed to setup auth middleware: %w", err)
	}

	executionRuntime := execservice.New(s.runtime, s)

	if err := api.Api(engine, executionRuntime, s); err != nil {
		return nil, fmt.Errorf("failed to setup API routes: %w", err)
	}

	if err := api.Database(engine, s); err != nil {
		return nil, fmt.Errorf("failed to setup database routes: %w", err)
	}

	if err := api.Reporter(engine, s.runtime.GetCache()); err != nil {
		return nil, fmt.Errorf("failed to setup reporter routes: %w", err)
	}

	api.Manual(engine, s.runtime.GetInteraction())
	api.FinAdmin(engine, finHandler)
	api.Logging(engine)
	api.Swagger(engine)

	return engine, nil
}

// setupAuthMiddleware configures authentication if enabled.
func (s *Server) setupAuthMiddleware(engine *gin.Engine) error {
	if !s.config.Auth.Enabled {
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

// newFinHandler creates and configures a FinHandler.
func (s *Server) newFinHandler() *finapi.FinHandler {
	cfg := finapi.Config{
		RegistrationToken:      s.config.Fin.RegistrationToken,
		PollIntervalSeconds:    s.config.Fin.PollIntervalSeconds,
		LongPollTimeoutSeconds: s.config.Fin.LongPollTimeoutSeconds,
		JobLeaseSeconds:        s.config.Fin.JobLeaseSeconds,
		StaleAfter:             s.config.Fin.StaleAfter,
	}
	return finapi.NewFinHandler(finapi.HandlerDependencies{
		Store:  s.runtime.GetFinStore(),
		Queue:  s.runtime.GetFinQueue(),
		Config: cfg,
		GUID:   new(guid.Guid),
	})
}

// RunServer starts the HTTP server on the configured port.
func (s *Server) RunServer(engine *gin.Engine) error {
	if s.config.Server.EnableTLS {
		if err := validateCertificates(s.config.Server.CertFile, s.config.Server.CertKey); err != nil {
			return err
		}
		log.Infof("Starting HTTPS server on port %s", s.config.Server.Port)
		return engine.RunTLS(":"+s.config.Server.Port, s.config.Server.CertFile, s.config.Server.CertKey)
	}

	log.Infof("Starting HTTP server on port %s", s.config.Server.Port)
	return engine.Run(":" + s.config.Server.Port)
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

// NewDecomposer creates a new decomposer with all capabilities and executors wired.
// Implements decomposer_controller.IController.
func (s *Server) NewDecomposer() decomposer.IDecomposer {
	sshCap := new(sshcap.SshCapability)
	capabilities := map[string]capability.ICapability{sshCap.GetType(): sshCap}

	httpUtil := new(httputil.HttpRequest)
	httpUtil.SkipCertificateValidation(s.config.HTTP.SkipCertValidation)
	httpCap := httpcap.New(httpUtil)
	capabilities[httpCap.GetType()] = httpCap

	openc2Cap := openc2cap.New(httpUtil)
	capabilities[openc2Cap.GetType()] = openc2Cap

	powershellCap := pscap.New()
	capabilities[powershellCap.GetType()] = powershellCap

	man := manualcap.New(s.runtime.GetInteraction())
	capabilities[man.GetType()] = &man

	report := reporter.New([]downstreamreport.IDownStreamReporter{})
	downstreamReporters := []downstreamreport.IDownStreamReporter{s.runtime.GetCache()}

	thehiveReporter, theHiveCaseManager := s.initializeTheHiveReporting()
	if thehiveReporter != nil {
		downstreamReporters = append(downstreamReporters, thehiveReporter)
	}

	report.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeutil.Time)
	assignmentExt := assignment.New()
	actionExec := actionexec.New(capabilities, report, soarcaTime, assignmentExt)

	actionExec.SetFinFallback(fincap.New(fincap.Dependencies{
		Queue:      s.runtime.GetFinQueue(),
		GUID:       new(guid.Guid),
		Store:      s.runtime.GetFinStore(),
		Time:       soarcaTime,
		StaleAfter: s.config.Fin.StaleAfter,
	}))

	pbExec := pbactionexec.New(s, s.runtime.GetPlaybookStore(), report, soarcaTime)
	stixCmp := stixcmp.New()
	condExec := condexec.New(stixCmp, report, soarcaTime)
	guidGen := new(guid.Guid)
	decomp := decomposer.New(actionExec, pbExec, condExec, guidGen, report, soarcaTime)

	if theHiveCaseManager != nil {
		decomp.SetCaseManager(theHiveCaseManager)
	}

	return decomp
}

// GetPlaybookStore returns the playbook store.
// Implements database.IController.
func (s *Server) GetPlaybookStore() storage.PlaybookStore {
	return s.runtime.GetPlaybookStore()
}

// initializeTheHiveReporting sets up The Hive integration if configured.
func (s *Server) initializeTheHiveReporting() (downstreamreport.IDownStreamReporter, cases.ICasesManager) {
	if !s.config.TheHive.Activate {
		return nil, nil
	}

	log.Info("Initializing The Hive reporting integration")

	if len(s.config.TheHive.APIBaseURL) < 1 || len(s.config.TheHive.APIToken) < 1 {
		log.Warning("Could not initialize The Hive reporting integration. Check environment variables.")
		return nil, nil
	}

	log.Infof("Creating The Hive connector with API base URL: %s", s.config.TheHive.APIBaseURL)
	conn := thehiveconnector.NewConnector(s.config.TheHive.APIBaseURL, s.config.TheHive.APIToken, s.config.TheHive.AllowInsecure)

	if s.config.TheHive.EnableCaseManager {
		log.Info("Enabling The Hive case manager")
		caseMgr := thehivecases.NewCaseManager(conn)
		return caseMgr, caseMgr
	}

	if s.config.TheHive.EnableReporter {
		log.Info("Enabling The Hive reporter")
		rep := thehivereport.NewReporter(conn)
		return rep, nil
	}

	return nil, nil
}
