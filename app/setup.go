package app

import (
	"errors"
	"os"

	mongo "soarca/database/mongodb"
	workflowRepo "soarca/database/workflow"
	routes "soarca/routes"
	"soarca/utils"
)

func SetupAndRunApp() error {
	LoadComponent()
	mongo.LoadComponent()

	log.Info("SOARCA API Trying to start")
	mongo_uri := os.Getenv("MONGODB_URI")
	db_username := os.Getenv("DB_USERNAME")
	db_password := os.Getenv("DB_PASSWORD")

	if mongo_uri == "" || db_username == "" || db_password == "" {
		log.Error("you must set 'MONGODB_URI' or 'DB_USERNAME' or 'DB_PASSWORD' in the environment variable")
		return errors.New("Could not obtain required environment settings")
	}
	err := mongo.SetupMongodb(mongo_uri, db_username, db_password)
	if err != nil {
		return err
	}
	// defer database.GetMongoClient().CloseMongoDB()

	workflowrepo := workflowRepo.SetupWorkflowRepository(mongo.GetCacaoRepo(), mongo.DefaultLimitOpts())
	api := routes.SetupRoutes(workflowrepo)
	// get the port and start
	port := utils.GetEnv("PORT", "8080")

	err = api.Run(":" + port)
	log.Info("Started the app")

	return err
}
