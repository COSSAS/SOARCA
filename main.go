package main

import (
	"fmt"

	"soarca/internal/controller"
	"soarca/logger"
	"soarca/swaggerdocs"
	"soarca/utils"

	"github.com/joho/godotenv"
)

var log *logger.Log

func init() {
	log = logger.Logger("MAIN", logger.Info, "", logger.Json)
}

var (
	Version   string
	Buildtime string
	Host      string
)

const banner = `
   _____  ____          _____   _____          
  / ____|/ __ \   /\   |  __ \ / ____|   /\    
 | (___ | |  | | /  \  | |__) | |       /  \   
  \___ \| |  | |/ /\ \ |  _  /| |      / /\ \  
  ____) | |__| / ____ \| | \ \| |____ / ____ \ 
 |_____/ \____/_/    \_\_|  \_\\_____/_/    \_\
                                               
                                               

`

// @title           SOARCA API
// @version         1.0.0
func main() {
	fmt.Print(banner)
	log.Info("Version: ", Version)
	log.Info("Buildtime: ", Buildtime)

	errenv := godotenv.Load(".env")
	if errenv != nil {
		log.Warning("Failed to read env variable, but will continue")
	}
	Host = "localhost:" + utils.GetEnv("PORT", "8080")
	swaggerdocs.SwaggerInfo.Host = Host

	errinit := controller.Initialize()
	if errinit != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", errinit)
		panic(errinit)
	}
}
