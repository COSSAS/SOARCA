package controller

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"soarca/internal/database/memory"
	"soarca/internal/logger"

	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/http"
	"soarca/pkg/core/capability/manual"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/core/capability/openc2"
	"soarca/pkg/core/capability/powershell"
	"soarca/pkg/core/capability/ssh"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/core/executors/action"
	"soarca/pkg/core/executors/condition"
	"soarca/pkg/core/executors/playbook_action"
	"soarca/pkg/extensions/soarca/assignment"
	"soarca/pkg/reporting/cases"
	"soarca/pkg/reporting/reporter"
	"soarca/pkg/utils"
	"soarca/pkg/utils/guid"
	"soarca/pkg/utils/stix/expression/comparison"
	"strconv"
	"strings"
	"time"

	thehiveCases "soarca/pkg/integration/thehive/cases"
	"soarca/pkg/integration/thehive/common/connector"
	thehive "soarca/pkg/integration/thehive/reporter"

	cache "soarca/pkg/reporting/reporter/downstream_reporter/cache"

	httpUtil "soarca/pkg/utils/http"

	timeUtil "soarca/pkg/utils/time"

	downstreamReporter "soarca/pkg/reporting/reporter/downstream_reporter"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"

	finrepository "soarca/internal/database/fin"
	"soarca/internal/database/finmemory"
	mongo "soarca/internal/database/mongodb"
	playbookrepository "soarca/internal/database/playbook"
	routes "soarca/pkg/api"
	fin_handler "soarca/pkg/api/fin"
	fincapability "soarca/pkg/core/capability/fin"
	"soarca/pkg/core/capability/fin/queue"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type Controller struct {
	playbookRepo playbookrepository.IPlaybookRepository
	finRepo      finrepository.IFinRepository
}

var mainController = Controller{}

var mainCache = cache.Cache{}

const defaultCacheSize int = 10

// One manual interaction per SOARCA instance
var mainInteraction = interaction.New(registerManualIntegration())

// One Fin job queue per SOARCA instance, shared between every
// action.Executor built by NewDecomposer() (one per execution/sub-
// execution) and the Fin API's poll/result/status handlers - all Fin jobs,
// regardless of which execution enqueued them, must land in this single
// queue so any live, matching Fin can claim them.
var mainFinQueue = queue.New()

const (
	defaultFinPollIntervalSeconds    = 5
	defaultFinLongPollTimeoutSeconds = 25
	defaultFinJobLeaseSeconds        = 60
	// finStaleAfterMultiplier bounds how long a registered Fin can go
	// without a /poll before FinCapability's fail-fast check stops
	// counting it as live (see fincapability.Capability.checkCapableFin,
	// which fails a step immediately rather than enqueuing it if every
	// Fin declaring its capability type is considered stale). A healthy
	// Fin's long-poll blocks for up to FIN_LONG_POLL_TIMEOUT_SECONDS
	// before it reconnects and updates LastSeen again, so this multiplier
	// is just a safety margin over that cadence for network/scheduling
	// jitter - not a separate, independently-configured timeout.
	finStaleAfterMultiplier = 2
)

// finLongPollTimeoutSeconds reads FIN_LONG_POLL_TIMEOUT_SECONDS (or its
// default), shared by newFinHandler (handed to Fins at registration) and
// NewDecomposer (used to derive the Fin-liveness staleness threshold) so
// both stay in sync from a single source.
func finLongPollTimeoutSeconds() int {
	seconds, _ := strconv.Atoi(utils.GetEnv("FIN_LONG_POLL_TIMEOUT_SECONDS", strconv.Itoa(defaultFinLongPollTimeoutSeconds)))
	return seconds
}

func (controller *Controller) NewDecomposer() decomposer.IDecomposer {
	ssh := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}

	skip, _ := strconv.ParseBool(utils.GetEnv("HTTP_SKIP_CERT_VALIDATION", "false"))

	httpUtil := new(httpUtil.HttpRequest)
	httpUtil.SkipCertificateValidation(skip)
	http := http.New(httpUtil)
	capabilities[http.GetType()] = http

	openc2 := openc2.New(httpUtil)
	capabilities[openc2.GetType()] = openc2

	poswershell := powershell.New()
	capabilities[poswershell.GetType()] = poswershell

	man := manual.New(mainInteraction)
	capabilities[man.GetType()] = &man

	// NOTE: Enrolling mainCache by default as reporter
	reporter := reporter.New([]downstreamReporter.IDownStreamReporter{})
	downstreamReporters := []downstreamReporter.IDownStreamReporter{&mainCache}

	// Reporter integrations

	thehive_reporter, theHiveCaseManager := initializeIntegrationTheHiveReporting()
	if thehive_reporter != nil {
		downstreamReporters = append(downstreamReporters, thehive_reporter)
	}

	reporter.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeUtil.Time)
	assignmentExtension := assignment.New()
	actionExecutor := action.New(capabilities, reporter, soarcaTime, assignmentExtension)
	// Any agent.Type not matching one of the built-in capabilities above
	// falls through to a live, registered Fin declaring that capability
	// type - Fin capability types are dynamic (declared at Fin
	// registration time), so unlike built-ins there is no static entry to
	// add to the capabilities map for them. controller.finRepo lets the
	// fallback fail a step immediately when no live Fin could possibly
	// claim it, instead of always waiting out the step's own timeout.
	staleAfter := time.Duration(finStaleAfterMultiplier*finLongPollTimeoutSeconds()) * time.Second
	actionExecutor.SetFinFallback(fincapability.New(mainFinQueue, new(guid.Guid), controller.finRepo, soarcaTime, staleAfter))
	playbookActionExecutor := playbook_action.New(controller, controller, reporter, soarcaTime)
	stixComparison := comparison.New()
	conditionExecutor := condition.New(stixComparison, reporter, soarcaTime)
	guid := new(guid.Guid)
	decompose := decomposer.New(actionExecutor,
		playbookActionExecutor,
		conditionExecutor,
		guid,
		reporter,
		soarcaTime)
	if theHiveCaseManager != nil {
		decompose.SetCaseManager(theHiveCaseManager)
	}
	return decompose
}

func (controller *Controller) setupDatabase() error {
	initMongoDatabase, _ := strconv.ParseBool(utils.GetEnv("DATABASE", "false"))

	if initMongoDatabase {

		mongo.LoadComponent()

		log.Info("SOARCA API Trying to start")
		uri := os.Getenv("MONGODB_URI")
		username := os.Getenv("DB_USERNAME")
		password := os.Getenv("DB_PASSWORD")

		if uri == "" || username == "" || password == "" {
			log.Error("you must set 'MONGODB_URI' or 'DB_USERNAME' or 'DB_PASSWORD' in the environment variable")
			return errors.New("could not obtain required environment settings")
		}
		err := mongo.SetupMongodb(uri, username, password)
		if err != nil {
			return err
		}
		controller.playbookRepo = playbookrepository.SetupPlaybookRepository(mongo.GetCacaoRepo(), mongo.DefaultLimitOpts())
		controller.finRepo = finrepository.SetupFinRepository(mongo.GetFinRepo())
	} else {
		// Use in memory database
		controller.playbookRepo = memory.New()
		controller.finRepo = finmemory.New()
	}

	return nil
}

func (controller *Controller) GetDatabaseInstance() playbookrepository.IPlaybookRepository {
	return controller.playbookRepo
}

func Initialize() error {
	app := gin.New()
	log.Info("Log level is info")
	log.Debug("Log level is debug")
	log.Trace("Log level is trace")

	cacheSize, _ := strconv.Atoi(utils.GetEnv("MAX_EXECUTIONS", strconv.Itoa(defaultCacheSize)))
	mainCache = *cache.New(&timeUtil.Time{}, cacheSize)

	err := initializeCore(app)
	if err != nil {
		log.Error("Failed to init core")
		return err
	}

	err = run(app)
	if err != nil {
		log.Error("failed to run gin")
	}
	log.Info("exit")
	return err
}

func validateCertificates(certFile string, keyFile string) error {
	_, err := os.Stat(certFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found: %s", certFile)
	}

	_, err = os.Stat(keyFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("key file not found: %s", keyFile)
	}
	return nil
}

func run(app *gin.Engine) error {
	port := utils.GetEnv("PORT", "8080")
	port = ":" + port
	enableTLS, _ := strconv.ParseBool(utils.GetEnv("ENABLE_TLS", "false"))
	certFile := utils.GetEnv("CERT_FILE", "./certs/server.crt")
	keyFile := utils.GetEnv("CERT_KEY_FILE", "./certs/server.key")

	if enableTLS {
		err := validateCertificates(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("TLS configuration error: %w", err)
		}
		log.Infof("Starting HTTPS server on port %s", port)
		return app.RunTLS(port, certFile, keyFile)

	}

	log.Infof("Starting HTTP server on port %s", port)
	return app.Run(port)
}

func initializeCore(app *gin.Engine) error {
	origins := strings.Split(strings.ReplaceAll(utils.GetEnv("SOARCA_ALLOWED_ORIGINS", "*"), " ", ""), ",")
	routes.Cors(app, origins)

	err := mainController.setupDatabase()
	if err != nil {
		log.Error("Failed to setup database:", err)
		return err
	}

	// Fin-token-authenticated routes (register/poll/jobs/status/unregister)
	// MUST be registered before intializeAuthenticationMiddleware below -
	// see FinPublic's doc comment and the warning at that call site. This
	// requires setupDatabase() (which populates mainController.finRepo) to
	// have already run, which is why it's been moved ahead of the auth
	// middleware too; it registers no routes itself, so this reordering is
	// safe with respect to auth.
	finHandler := newFinHandler()
	routes.FinPublic(app, finHandler)

	// #############################################################
	// WARNING: intializeAuthenticationMiddleware installs the global
	// soarca_admin JWT middleware via app.Use(); gin copies engine-level
	// middleware into a route's handler chain at the time the route is
	// registered, so anything registered above this line does NOT get
	// gated by it, and anything registered below DOES. routes.FinPublic
	// (above) deliberately relies on being above this line - do not reorder
	// it below, and do not move this call above it, or Fin processes
	// (which authenticate via fin_token, not a JWT) will be locked out of
	// their own protocol entirely.
	// #############################################################
	err = intializeAuthenticationMiddleware(app)
	if err != nil {
		log.Error("Failed to setup Authentication middleware")
		return err
	}

	err = routes.Api(app, &mainController, &mainController)
	if err != nil {
		log.Error(err)
		return err
	}

	err = routes.Database(app, &mainController)
	if err != nil {
		log.Error(err)
		return err
	}

	// NOTE: Assuming that the cache is the main information mediator for
	// the reporter API
	err = routes.Reporter(app, &mainCache)
	if err != nil {
		log.Error(err)
		return err
	}

	// Manual capability native routes
	routes.Manual(app, mainInteraction)

	// Fin discovery routes (list/get) - ordinary admin/dashboard reads,
	// registered here (behind the admin auth middleware above) like the
	// rest of the admin API. Unlike FinPublic, there is no ordering
	// constraint on these.
	routes.FinAdmin(app, finHandler)

	routes.Logging(app)
	routes.Swagger(app)

	return err
}

// newFinHandler builds the Fin protocol's API handler, sharing the same
// job queue (mainFinQueue) that action.Executor instances enqueue onto (see
// NewDecomposer) and the Fin registry populated by setupDatabase.
func newFinHandler() *fin_handler.FinHandler {
	pollIntervalSeconds, _ := strconv.Atoi(utils.GetEnv("FIN_POLL_INTERVAL_SECONDS", strconv.Itoa(defaultFinPollIntervalSeconds)))
	longPollTimeoutSeconds := finLongPollTimeoutSeconds()
	jobLeaseSeconds, _ := strconv.Atoi(utils.GetEnv("FIN_JOB_LEASE_SECONDS", strconv.Itoa(defaultFinJobLeaseSeconds)))

	config := fin_handler.Config{
		// Empty by default: Register then always fails closed (see
		// fin_handler.Config's doc comment) rather than silently accepting
		// any registration attempt when an operator forgets to set this.
		RegistrationToken:      utils.GetEnv("FIN_REGISTRATION_TOKEN", ""),
		PollIntervalSeconds:    pollIntervalSeconds,
		LongPollTimeoutSeconds: longPollTimeoutSeconds,
		JobLeaseSeconds:        jobLeaseSeconds,
		// Matches the threshold fed into fincapability.New in
		// NewDecomposer, so a Fin flagged Stale here is the same Fin that
		// fails fast as "only stale" in the capability's liveness check.
		StaleAfterSeconds: finStaleAfterMultiplier * longPollTimeoutSeconds,
	}

	return fin_handler.NewFinHandler(mainController.finRepo, mainFinQueue, config, new(guid.Guid))
}

func registerManualIntegration() []interaction.IInteractionIntegrationNotifier {
	// Manual interaction integrations will be initialized here when implemented
	// Here we should check ENV variables, see if a manual interaction integration is selected,
	// And populate the returned array via generating an instance of the notifier associated with
	// the integration - which should be found in the integration code.
	return []interaction.IInteractionIntegrationNotifier{}
}

func initializeIntegrationTheHiveReporting() (downstreamReporter.IDownStreamReporter, cases.ICasesManager) {
	initTheHiveReporter, _ := strconv.ParseBool(utils.GetEnv("THEHIVE_ACTIVATE", "false"))
	if !initTheHiveReporter {
		return nil, nil
	}
	log.Info("initializing The Hive reporting integration")

	thehiveApiToken := utils.GetEnv("THEHIVE_API_TOKEN", "")
	thehiveApiBaseUrl := utils.GetEnv("THEHIVE_API_BASE_URL", "")
	if len(thehiveApiBaseUrl) < 1 || len(thehiveApiToken) < 1 {
		log.Warning("could not initialize The Hive reporting integration. Check to have configured the env variables correctly.")
		return nil, nil
	}

	theHiveInsecureConnection, _ := strconv.ParseBool(utils.GetEnv("THEHIVE_ALLOW_INSECURE", "true"))

	log.Info(fmt.Sprintf("creating new The hive connector with API base url at : %s", thehiveApiBaseUrl))
	theHiveConnector := connector.NewConnector(thehiveApiBaseUrl, thehiveApiToken, theHiveInsecureConnection)
	caseReporting, _ := strconv.ParseBool(utils.GetEnv("THEHIVE_REPORTER", "false"))
	if caseReporting {
		log.Info("enabling the hive reporter")
		theHiveCases := thehiveCases.NewCaseManager(theHiveConnector)
		return theHiveCases, theHiveCases
	}
	theHiveReporter := thehive.NewReporter(theHiveConnector)
	return theHiveReporter, nil
}

func intializeAuthenticationMiddleware(app *gin.Engine) error {
	authEnabled, _ := strconv.ParseBool(utils.GetEnv("AUTH_ENABLED", "false"))
	if authEnabled {
		auth, err := gauth.New(gauth.DefaultConfig())
		if err != nil {
			log.Error("Failed to initialize authenticator:", err)
			return err
		}
		app.Use(auth.LoadAuthContext())
		app.Use(auth.Middleware([]string{"soarca_admin"}))

	}
	return nil
}
