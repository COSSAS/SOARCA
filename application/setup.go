package application

import (
	"soarca/utils"

	"github.com/gin-gonic/gin"
)

func InitialiseAppComponents() error {
	app := gin.New()

	initDatabase := utils.GetEnv("DATABASE", "true")
	if initDatabase == "true" {
		errDatabase := InitialliseDatabase(app)
		if errDatabase != nil {
			log.Error("Failed to init core")
			return errDatabase
		}
	}
	errCore := InitialiseCore(app)

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
