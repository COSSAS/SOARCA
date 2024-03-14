package controller

import (
	"errors"
	"os"
	"reflect"

	"soarca/internal/capability"
	"soarca/internal/capability/http"
	"soarca/internal/capability/ssh"
	"soarca/internal/decomposer"
	"soarca/internal/executer"
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
	playbookRepo playbookrepository.IPlaybookRepository
}

var mainController = Controller{}

func (controller *Controller) NewDecomposer() decomposer.IDecomposer {
	ssh := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}

	http := new(http.HttpCapability)
	capabilities[http.GetType()] = http

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
		mainController.setupDatabase()
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
