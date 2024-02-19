package application

import (
	"soarca/utils"

	"github.com/gin-gonic/gin"
)

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
