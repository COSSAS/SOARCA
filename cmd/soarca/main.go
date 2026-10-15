package main

import (
	"fmt"

	api "soarca/api"
	"soarca/internal/config"
	"soarca/internal/controller"
	"soarca/internal/logger"
	"soarca/pkg/api/status"

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

// @title		SOARCA API
// @version	1.0.0
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				"Bearer " followed by an access token or API key. Not needed when AUTH_ENABLED is false.
func main() {
	fmt.Print(banner)
	log.Info("Version: ", Version)
	log.Info("Buildtime: ", Buildtime)

	err := godotenv.Load(".env")
	if err != nil {
		log.Warning("Failed to read env variables because of error: ", err, " , but will continue.")
	}
	configuration, err := config.Load()
	if err != nil {
		log.Fatal("Invalid configuration: ", err)
		panic(err)
	}

	Host = fmt.Sprintf("localhost:%d", configuration.Static.Server.Port)
	api.SwaggerInfo.Host = Host

	// Version is only available here
	status.SetVersion(Version)
	err = controller.Initialize(configuration)
	if err != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", err)
		panic(err)
	}
}
