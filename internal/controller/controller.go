package controller

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"soarca/internal/config"
	keymanagementrepository "soarca/internal/database/keymanagement"
	"soarca/internal/database/memory"
	memorykms "soarca/internal/database/memory/keymanagement"
	"soarca/internal/logger"

	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/fin/protocol"
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
	"soarca/pkg/keymanagement"
	"soarca/pkg/keymanagement/loader"
	"soarca/pkg/reporting/cases"
	"soarca/pkg/reporting/reporter"
	"soarca/pkg/utils/guid"
	"soarca/pkg/utils/stix/expression/comparison"

	finExecutor "soarca/pkg/core/capability/fin"
	finChannelController "soarca/pkg/core/capability/fin/controller"

	thehiveCases "soarca/pkg/integration/thehive/cases"
	"soarca/pkg/integration/thehive/common/connector"
	thehive "soarca/pkg/integration/thehive/reporter"

	cache "soarca/pkg/reporting/reporter/downstream_reporter/cache"

	httpUtil "soarca/pkg/utils/http"

	timeUtil "soarca/pkg/utils/time"

	downstreamReporter "soarca/pkg/reporting/reporter/downstream_reporter"

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
	configuration     *config.Config
	finController     finChannelController.IFinController
	playbookRepo      playbookrepository.IPlaybookRepository
	keyManagementRepo keymanagementrepository.IKeyManagementRepository
	keyManagement     *keymanagement.KeyManagement
}

var mainController = Controller{}

var mainCache = cache.Cache{}

// One manual interaction per SOARCA instance
var mainInteraction = interaction.New(registerManualIntegration())

func (controller *Controller) NewDecomposer() decomposer.IDecomposer {
	ssh := &ssh.SshCapability{Keys: controller.keyManagement}
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}

	// Read per decomposer, so an update applies to the next execution.
	skip := controller.configuration.Dynamic().HttpSkipCertValidation

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

	fins := controller.configuration.Static.Fins
	if fins.Enabled {
		broker, port := fins.MqttBroker, fins.MqttPort

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

	thehive_reporter, theHiveCaseManager := initializeIntegrationTheHiveReporting(controller.configuration.Static.TheHive)
	if thehive_reporter != nil {
		downstreamReporters = append(downstreamReporters, thehive_reporter)
	}

	reporter.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeUtil.Time)
	assignmentExtension := assignment.New()
	actionExecutor := action.New(capabilities, reporter, soarcaTime, assignmentExtension)
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
	database := controller.configuration.Static.Database

	if database.Enabled {
		log.Info("Setting up mongo database")

		mongo.LoadComponent()

		log.Info("SOARCA API Trying to start")
		// config.Load already checked that uri, username and password are set.
		err := mongo.SetupMongodb(database.Uri, database.Username, database.Password)
		if err != nil {
			return err
		}
		controller.playbookRepo = playbookrepository.SetupPlaybookRepository(mongo.GetCacaoRepo(), mongo.DefaultLimitOpts())
		controller.keyManagementRepo = memorykms.New()

		// Ad-hoc creation of mongo client TODO: refactor mongo client.
		kms, err := mongo.NewMongoCollection[keymanagementrepository.KeyPairEntry](mongo.GetCacaoRepo().Collection.Database().Client(), "database-kms", "kms-collection")
		if err != nil {
			return errors.New("could not create database collection for kms")
		}
		controller.keyManagementRepo = keymanagementrepository.SetupKeyManagementRepository(kms, mongo.DefaultLimitOpts())

	} else {
		// Use in memory database
		log.Info("Setting up in-memory database")
		controller.playbookRepo = memory.New()
		controller.keyManagementRepo = memorykms.New()
	}

	return nil
}

func (controller *Controller) setupKeyManagement() error {
	sshKms := controller.configuration.Static.SshKms
	if sshKms.Enabled {
		controller.keyManagement = keymanagement.New(controller.keyManagementRepo)
		// Load keys from directory
		kmsDir := sshKms.Dir
		if kmsDir != "" {
			err := loader.Load(kmsDir, controller.keyManagement)
			if err != nil {
				log.Error("failed to load kms dir: ", err)
			}
			log.Info("loaded kms keys from: ", kmsDir)
		}
		log.Info("KMS is enabled")
	} else {
		log.Trace("KMS is not enabled")
	}
	return nil
}

func (controller *Controller) GetDatabaseInstance() playbookrepository.IPlaybookRepository {
	return controller.playbookRepo
}

// Config returns the configuration SOARCA was initialized with, for API
// handlers that read or update the dynamic settings.
func (controller *Controller) Config() *config.Config {
	return controller.configuration
}

func Initialize(configuration *config.Config) error {
	if configuration == nil {
		return errors.New("controller needs a configuration")
	}
	mainController.configuration = configuration

	gin.SetMode(configuration.Static.Server.GinMode)
	app := gin.New()
	log.Info("Log level is info")
	log.Debug("Log level is debug")
	log.Trace("Log level is trace")

	if configuration.Static.Fins.Enabled {
		if err := mainController.setupAndRunMqtt(); err != nil {
			log.Error(err)
		}
	}

	mainCache = *cache.New(&timeUtil.Time{}, configuration.Static.MaxExecutions)

	err := initializeCore(app, configuration)
	if err != nil {
		log.Error("Failed to init core")
		return err
	}

	err = run(app, configuration.Static.Server)
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

func run(app *gin.Engine, server config.Server) error {
	port := fmt.Sprintf(":%d", server.Port)
	certFile := server.CertFile
	keyFile := server.CertKeyFile

	if server.EnableTls {
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

func initializeCore(app *gin.Engine, configuration *config.Config) error {
	routes.Cors(app, func() []string {
		return configuration.Dynamic().AllowedOrigins
	})

	err := intializeAuthenticationMiddleware(app, configuration.Static.Auth)
	if err != nil {
		log.Error("Failed to setup Authentication middleware")
		return err
	}
	err = mainController.setupDatabase()
	if err != nil {
		log.Error("Failed to setup database:", err)
		return err
	}

	err = mainController.setupKeyManagement()
	if err != nil {
		log.Error("Failed to setup key management:", err)
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

	routes.KeyManagement(app, mainController.keyManagement)

	// Manual capability native routes
	routes.Manual(app, mainInteraction)

	routes.Logging(app)
	routes.Swagger(app)

	return err
}

func (controller *Controller) setupAndRunMqtt() error {
	fins := controller.configuration.Static.Fins
	broker, port := fins.MqttBroker, fins.MqttPort
	mqttClient := finChannelController.NewClient(protocol.Broker(broker), port)
	finChannelController := finChannelController.New(*mqttClient)
	controller.finController = finChannelController
	err := finChannelController.ConnectAndSubscribe()
	if err != nil {
		log.Error(err)
		return err
	}
	go finChannelController.Run()
	return nil
}

func registerManualIntegration() []interaction.IInteractionIntegrationNotifier {
	// Manual interaction integrations will be initialized here when implemented
	// Here we should check ENV variables, see if a manual interaction integration is selected,
	// And populate the returned array via generating an instance of the notifier associated with
	// the integration - which should be found in the integration code.
	return []interaction.IInteractionIntegrationNotifier{}
}

func initializeIntegrationTheHiveReporting(theHive config.TheHive) (downstreamReporter.IDownStreamReporter, cases.ICasesManager) {
	if !theHive.Activate {
		return nil, nil
	}
	log.Info("initializing The Hive reporting integration")

	thehiveApiToken := theHive.ApiToken
	thehiveApiBaseUrl := theHive.ApiBaseUrl
	if len(thehiveApiBaseUrl) < 1 || len(thehiveApiToken) < 1 {
		log.Warning("could not initialize The Hive reporting integration. Check to have configured the env variables correctly.")
		return nil, nil
	}

	theHiveInsecureConnection := theHive.AllowInsecure

	log.Info(fmt.Sprintf("creating new The hive connector with API base url at : %s", thehiveApiBaseUrl))
	theHiveConnector := connector.NewConnector(thehiveApiBaseUrl, thehiveApiToken, theHiveInsecureConnection)
	if theHive.Reporter {
		log.Info("enabling the hive reporter")
		theHiveCases := thehiveCases.NewCaseManager(theHiveConnector)
		return theHiveCases, theHiveCases
	}
	theHiveReporter := thehive.NewReporter(theHiveConnector)
	return theHiveReporter, nil
}

func intializeAuthenticationMiddleware(app *gin.Engine, authConfig config.Auth) error {
	if authConfig.Enabled {
		// Same settings gauth.DefaultConfig() builds, but from SOARCA's config
		// instead of gauth reading the environment itself.
		auth, err := gauth.New(&gauth.Config{
			Mode:              gauth.ModeVerify,
			IssuerUri:         authConfig.Issuer,
			ClientID:          authConfig.ClientId,
			Provider:          gauth.Authentik,
			SkipTLSValidation: authConfig.SkipTlsVerify,
		})
		if err != nil {
			log.Error("Failed to initialize authenticator:", err)
			return err
		}
		app.Use(auth.LoadAuthContext())
		app.Use(auth.Middleware([]string{"soarca_admin"}))

	}
	return nil
}
