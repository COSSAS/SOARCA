package controller

import (
	"errors"
	"os"
	"reflect"
	"strconv"

	"soarca/internal/capability"
	capabilityController "soarca/internal/capability/controller"
	finExecutor "soarca/internal/capability/fin"
	"soarca/internal/capability/http"
	"soarca/internal/capability/openc2"
	"soarca/internal/capability/ssh"
	"soarca/internal/decomposer"
	"soarca/internal/executors/action"
	"soarca/internal/fin/protocol"
	"soarca/internal/guid"
	"soarca/internal/reporters"
	"soarca/logger"
	"soarca/utils"
	httpUtil "soarca/utils/http"

	//mockReporter "soarca/test/unittest/mocks/mock_reporter"

	"github.com/gin-gonic/gin"

	mongo "soarca/database/mongodb"
	playbookrepository "soarca/database/playbook"
	db_reporter "soarca/internal/reporters/database"
	routes "soarca/routes"
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

func (controller *Controller) NewDecomposer() decomposer.IDecomposer {
	ssh := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}

	httpUtil := new(httpUtil.HttpRequest)
	http := http.New(httpUtil)
	capabilities[http.GetType()] = http

	openc2 := openc2.New(httpUtil)
	capabilities[openc2.GetType()] = openc2

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

	// TODO: Instantiate reporters from config
	reporter := reporters.New(
		[]reporters.IWorkflowReporter{new(db_reporter.DatabaseReporter)},
		[]reporters.IStepReporter{new(db_reporter.DatabaseReporter)},
	)

	actionExecutor := action.New(capabilities, reporter)
	guid := new(guid.Guid)
	decompose := decomposer.New(actionExecutor, guid, reporter)
	return decompose
}

func (controller *Controller) setupDatabase() error {
	mongo.LoadComponent()

	log.Info("SOARCA API Trying to start")
	mongo_uri := os.Getenv("MONGODB_URI")
	db_username := os.Getenv("DB_USERNAME")
	db_password := os.Getenv("DB_PASSWORD")

	if mongo_uri == "" || db_username == "" || db_password == "" {
		log.Error("you must set 'MONGODB_URI' or 'DB_USERNAME' or 'DB_PASSWORD' in the environment variable")
		return errors.New("could not obtain required environment settings")
	}
	err := mongo.SetupMongodb(mongo_uri, db_username, db_password)
	if err != nil {
		return err
	}
	controller.playbookRepo = playbookrepository.SetupPlaybookRepository(mongo.GetCacaoRepo(), mongo.DefaultLimitOpts())

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

	errCore := initializeCore(app)

	if errCore != nil {
		log.Error("Failed to init core")
		return errCore
	}

	port := utils.GetEnv("PORT", "8080")
	err := app.Run(":" + port)
	if err != nil {
		log.Error("failed to run gin")
	}
	log.Info("exit")

	return err
}

func initializeCore(app *gin.Engine) error {

	err := routes.Api(app, &mainController)
	if err != nil {
		log.Error(err)
		return err
	}

	initDatabase := utils.GetEnv("DATABASE", "false")
	if initDatabase == "true" {
		err = mainController.setupDatabase()
		if err != nil {
			log.Error(err)
			return err
		}
		err = routes.Database(app, &mainController)
		if err != nil {
			log.Error(err)
			return err
		}
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

func getMqttDetails() (string, int) {
	broker := utils.GetEnv("MQTT_BROKER", "localhost")
	port, err := strconv.Atoi(utils.GetEnv("MQTT_PORT", "1883"))
	if err != nil {
		port = 1883
	}
	return broker, port
}
