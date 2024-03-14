package controller

import (
	"errors"
	"os"
	"reflect"

	"soarca/internal/capability"
	capabilityController "soarca/internal/capability/controller"
	finExecutor "soarca/internal/capability/fin"
	"soarca/internal/capability/http"
	"soarca/internal/capability/ssh"
	"soarca/internal/decomposer"
	"soarca/internal/executer"
	"soarca/internal/fin/protocol"
	"soarca/internal/guid"
	"soarca/logger"
	"soarca/utils"

	"github.com/gin-gonic/gin"

	mongo "soarca/database/mongodb"
	playbookrepository "soarca/database/playbook"
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

	http := new(http.HttpCapability)
	capabilities[http.GetType()] = http

	finCapabilities := controller.finController.GetRegisteredCapabilities()
	for key := range finCapabilities {
		prot := protocol.New(&guid.Guid{}, protocol.Topic(key), "localhost", 1883)
		fin := finExecutor.New(&prot)
		capabilities[key] = fin
	}

	executer := executer.New(capabilities)
	guid := new(guid.Guid)
	decompose := decomposer.New(executer, guid)
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
	log.Info("Testing if info log works")
	log.Debug("Testing if debug log works")
	log.Trace("Testing if Trace log works")

	if err := mainController.setupAndRunMqtt(); err != nil {
		log.Error(err)
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
	mqttClient := capabilityController.NewClient("localhost", 1883)
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
