package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"soarca/internal/logger"
	"soarca/internal/storage"
	storage_memory "soarca/internal/storage/memory"
	storage_mongodb "soarca/internal/storage/mongodb"

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

	thehiveCases "soarca/pkg/integration/thehive/cases"
	"soarca/pkg/integration/thehive/common/connector"
	thehive "soarca/pkg/integration/thehive/reporter"

	cache "soarca/pkg/reporting/reporter/downstream_reporter/cache"

	"soarca/pkg/api"
	"soarca/pkg/api/fin"
	fincapability "soarca/pkg/core/capability/fin"
	"soarca/pkg/core/capability/fin/queue"

	downstreamReporter "soarca/pkg/reporting/reporter/downstream_reporter"

	httpUtil "soarca/pkg/utils/http"
	timeUtil "soarca/pkg/utils/time"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type Controller struct {
	playbookStore storage.PlaybookStore
	finStore      storage.FinStore
}

var mainController = Controller{}
var mainCache = cache.Cache{}

const defaultCacheSize int = 10

var mainInteraction = interaction.New(registerManualIntegration())
var mainFinQueue = queue.New()

const (
	defaultFinPollIntervalSeconds    = 5
	defaultFinLongPollTimeoutSeconds = 25
	defaultFinJobLeaseSeconds        = 60
	finStaleAfterMultiplier          = 2
)

func finLongPollTimeoutSeconds() int {
	seconds, _ := strconv.Atoi(utils.GetEnv("FIN_LONG_POLL_TIMEOUT_SECONDS", strconv.Itoa(defaultFinLongPollTimeoutSeconds)))
	return seconds
}

func (controller *Controller) NewDecomposer() decomposer.IDecomposer {
	sshCap := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{sshCap.GetType(): sshCap}

	skip, _ := strconv.ParseBool(utils.GetEnv("HTTP_SKIP_CERT_VALIDATION", "false"))

	httpUtil := new(httpUtil.HttpRequest)
	httpUtil.SkipCertificateValidation(skip)
	httpCap := http.New(httpUtil)
	capabilities[httpCap.GetType()] = httpCap

	openc2Cap := openc2.New(httpUtil)
	capabilities[openc2Cap.GetType()] = openc2Cap

	powershellCap := powershell.New()
	capabilities[powershellCap.GetType()] = powershellCap

	man := manual.New(mainInteraction)
	capabilities[man.GetType()] = &man

	reporter := reporter.New([]downstreamReporter.IDownStreamReporter{})
	downstreamReporters := []downstreamReporter.IDownStreamReporter{&mainCache}

	thehiveReporter, theHiveCaseManager := initializeIntegrationTheHiveReporting()
	if thehiveReporter != nil {
		downstreamReporters = append(downstreamReporters, thehiveReporter)
	}

	reporter.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeUtil.Time)
	assignmentExtension := assignment.New()
	actionExecutor := action.New(capabilities, reporter, soarcaTime, assignmentExtension)
	staleAfter := time.Duration(finStaleAfterMultiplier*finLongPollTimeoutSeconds()) * time.Second
	actionExecutor.SetFinFallback(fincapability.New(mainFinQueue, new(guid.Guid), controller.finStore, soarcaTime, staleAfter))
	playbookActionExecutor := playbook_action.New(controller, controller.playbookStore, reporter, soarcaTime)
	stixComparison := comparison.New()
	conditionExecutor := condition.New(stixComparison, reporter, soarcaTime)
	guidGen := new(guid.Guid)
	decompose := decomposer.New(actionExecutor, playbookActionExecutor, conditionExecutor, guidGen, reporter, soarcaTime)
	if theHiveCaseManager != nil {
		decompose.SetCaseManager(theHiveCaseManager)
	}
	return decompose
}

func (controller *Controller) setupDatabase() error {
	initMongoDatabase, _ := strconv.ParseBool(utils.GetEnv("DATABASE", "false"))

	var store storage.Store
	var err error
	if initMongoDatabase {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			return errors.New("could not obtain required environment settings")
		}
		store, err = storage_mongodb.New(context.Background(), storage_mongodb.Config{URI: uri})
		if err != nil {
			return err
		}
	} else {
		store = storage_memory.New()
	}

	controller.playbookStore = store.Playbooks()
	controller.finStore = store.Fins()
	return nil
}

func (controller *Controller) GetPlaybookStore() storage.PlaybookStore {
	return controller.playbookStore
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
	api.Cors(app, origins)

	err := mainController.setupDatabase()
	if err != nil {
		log.Error("Failed to setup database:", err)
		return err
	}

	finHandler := newFinHandler()
	api.FinPublic(app, finHandler)

	err = intializeAuthenticationMiddleware(app)
	if err != nil {
		log.Error("Failed to setup Authentication middleware")
		return err
	}

	err = api.Api(app, &mainController, &mainController)
	if err != nil {
		log.Error(err)
		return err
	}

	err = api.Database(app, &mainController)
	if err != nil {
		log.Error(err)
		return err
	}

	err = api.Reporter(app, &mainCache)
	if err != nil {
		log.Error(err)
		return err
	}

	api.Manual(app, mainInteraction)
	api.FinAdmin(app, finHandler)
	api.Logging(app)
	api.Swagger(app)

	return err
}

func newFinHandler() *fin.FinHandler {
	pollIntervalSeconds, _ := strconv.Atoi(utils.GetEnv("FIN_POLL_INTERVAL_SECONDS", strconv.Itoa(defaultFinPollIntervalSeconds)))
	longPollTimeoutSeconds := finLongPollTimeoutSeconds()
	jobLeaseSeconds, _ := strconv.Atoi(utils.GetEnv("FIN_JOB_LEASE_SECONDS", strconv.Itoa(defaultFinJobLeaseSeconds)))

	config := fin.Config{
		RegistrationToken:      utils.GetEnv("FIN_REGISTRATION_TOKEN", ""),
		PollIntervalSeconds:    pollIntervalSeconds,
		LongPollTimeoutSeconds: longPollTimeoutSeconds,
		JobLeaseSeconds:        jobLeaseSeconds,
		StaleAfterSeconds:      finStaleAfterMultiplier * longPollTimeoutSeconds,
	}

	return fin.NewFinHandler(mainController.finStore, mainFinQueue, config, new(guid.Guid))
}

func registerManualIntegration() []interaction.IInteractionIntegrationNotifier {
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
