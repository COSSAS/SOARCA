package controller

import (
	"errors"
	"os"
	"reflect"

	"soarca/internal/capability"
	"soarca/internal/capability/ssh"
	"soarca/internal/decomposer"
	"soarca/internal/executer"
	"soarca/internal/guid"
	"soarca/logger"
	"soarca/utils"

	"github.com/gin-gonic/gin"

	mongo "soarca/database/mongodb"
	playbookRepo "soarca/database/playbook"
	routes "soarca/routes"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

func InitializeAppComponents() error {
	app := gin.New()
	log.Info("Testing if this works")

	initDatabase := utils.GetEnv("DATABASE", "false")
	if initDatabase == "true" {
		errDatabase := InitializeDatabase(app)
		if errDatabase != nil {
			log.Error("Failed to init core")
			return errDatabase
		}
	}
	errCore := InitializeCore(app)

	if errCore != nil {
		log.Error("Failed to init core")
		return errCore
	}

	port := utils.GetEnv("PORT", "8080")
	err := app.Run(":" + port)
	if err != nil {
		log.Error("failed to run gin")
	}
	return err
}

func InitializeDatabase(app *gin.Engine) error {
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
	// defer database.GetMongoClient().CloseMongoDB()

	playbookRepo := playbookRepo.SetupPlaybookRepository(mongo.GetCacaoRepo(), mongo.DefaultLimitOpts())

	// setup database routes
	err = routes.Database(app, playbookRepo)

	return err
}

func InitializeCore(app *gin.Engine) error {
	ssh := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}
	executer := executer.New(capabilities)
	guid := new(guid.Guid)
	decompose := decomposer.New(executer, guid)

	err := routes.Api(app, decompose)
	if err != nil {
		log.Error(err)
		return err
	}
	routes.Logging(app)
	routes.Swagger(app)
	return err
}
