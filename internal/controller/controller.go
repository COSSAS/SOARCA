package controller

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"soarca/internal/database/memory"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/caldera"
	"soarca/pkg/core/capability/fin/protocol"
	"soarca/pkg/core/capability/http"
	"soarca/pkg/core/capability/openc2"
	"soarca/pkg/core/capability/powershell"
	"soarca/pkg/core/capability/ssh"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/core/executors/action"
	"soarca/pkg/core/executors/condition"
	"soarca/pkg/core/executors/playbook_action"
	"soarca/pkg/reporter"
	"soarca/pkg/utils"
	"soarca/pkg/utils/guid"
	"soarca/pkg/utils/stix/expression/comparison"
	"strconv"
	"strings"

	capabilityController "soarca/pkg/core/capability/controller"
	finExecutor "soarca/pkg/core/capability/fin"

	thehive "soarca/pkg/integration/thehive/reporter"

	cache "soarca/pkg/reporter/downstream_reporter/cache"

	httpUtil "soarca/pkg/utils/http"

	timeUtil "soarca/pkg/utils/time"

	downstreamReporter "soarca/pkg/reporter/downstream_reporter"

	"github.com/COSSAS/gauth"
	"github.com/gin-gonic/gin"

	mongo "soarca/internal/database/mongodb"
	playbookrepository "soarca/internal/database/playbook"
	routes "soarca/pkg/api"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type Controller struct {
	finController capabilityController.IFinController
	playbookRepo  playbookrepository.IPlaybookRepository
}

var mainController = Controller{}

var mainCache = cache.Cache{}

const defaultCacheSize int = 10

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

	calderaCapability := caldera.New()
	capabilities[calderaCapability.GetType()] = calderaCapability

	enableFins, _ := strconv.ParseBool(utils.GetEnv("ENABLE_FINS", "false"))

	if enableFins {
		broker, port := getMqttDetails()

		finCapabilities := controller.finController.GetRegisteredCapabilities()
		for key := range finCapabilities {
			prot := protocol.New(&guid.Guid{}, protocol.Topic(key), protocol.Broker(broker), port)
			fin := finExecutor.New(&prot)
			capabilities[key] = fin
		}
	}

	// NOTE: Enrolling mainCache by default as reporter
	reporter := reporter.New([]downstreamReporter.IDownStreamReporter{})
	downstreamReporters := []downstreamReporter.IDownStreamReporter{&mainCache}

	// Reporter integrations

	thehive_reporter := initializeIntegrationTheHiveReporting()
	if thehive_reporter != nil {
		downstreamReporters = append(downstreamReporters, thehive_reporter)
	}

	reporter.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeUtil.Time)
	actionExecutor := action.New(capabilities, reporter, soarcaTime)
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
	} else {
		// Use in memory database
		controller.playbookRepo = memory.New()
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

	enableFins, _ := strconv.ParseBool(utils.GetEnv("ENABLE_FINS", "false"))
	if enableFins {
		if err := mainController.setupAndRunMqtt(); err != nil {
			log.Error(err)
		}
	}

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

	err := intializeAuthenticationMiddleware(app)
	if err != nil {
		log.Error("Failed to setup Authentication middleware")
		return err
	}
	err = mainController.setupDatabase()
	if err != nil {
		log.Error("Failed to setup database:", err)
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

	routes.Logging(app)
	routes.Swagger(app)

	return err
}

func (controller *Controller) setupAndRunMqtt() error {
	broker, port := getMqttDetails()
	mqttClient := capabilityController.NewClient(protocol.Broker(broker), port)
	capabilityController := capabilityController.New(*mqttClient)
	controller.finController = capabilityController
	err := capabilityController.ConnectAndSubscribe()
	if err != nil {
		log.Error(err)
		return err
	}
	go capabilityController.Run()
	return nil
}

func initializeIntegrationTheHiveReporting() downstreamReporter.IDownStreamReporter {
	initTheHiveReporter, _ := strconv.ParseBool(utils.GetEnv("THEHIVE_ACTIVATE", "false"))
	if !initTheHiveReporter {
		return nil
	}
	log.Info("initializing The Hive reporting integration")

	thehiveApiToken := utils.GetEnv("THEHIVE_API_TOKEN", "")
	thehiveApiBaseUrl := utils.GetEnv("THEHIVE_API_BASE_URL", "")
	if len(thehiveApiBaseUrl) < 1 || len(thehiveApiToken) < 1 {
		log.Warning("could not initialize The Hive reporting integration. Check to have configured the env variables correctly.")
		return nil
	}

	log.Info(fmt.Sprintf("creating new The hive connector with API base url at : %s", thehiveApiBaseUrl))
	theHiveConnector := thehive.NewConnector(thehiveApiBaseUrl, thehiveApiToken)
	theHiveReporter := thehive.NewReporter(theHiveConnector)
	return theHiveReporter
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

func getMqttDetails() (string, int) {
	broker := utils.GetEnv("MQTT_BROKER", "localhost")
	port, err := strconv.Atoi(utils.GetEnv("MQTT_PORT", "1883"))
	if err != nil {
		port = 1883
	}
	return broker, port
}
