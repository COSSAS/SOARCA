package main

import (
	"fmt"

	api "soarca/api"
	"soarca/internal/controller"
	"soarca/internal/logger"
	"soarca/pkg/api/status"
	"soarca/pkg/utils"

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
func main() {
	fmt.Print(banner)
	log.Info("Version: ", Version)
	log.Info("Buildtime: ", Buildtime)

	err := godotenv.Load(".env")
	if err != nil {
		log.Warning("Failed to read env variable, but will continue. Error: ", err)
	}
	Host = "localhost:" + utils.GetEnv("PORT", "8080")
	api.SwaggerInfo.Host = Host

	// Version is only available here
	status.SetVersion(Version)
	err = controller.Initialize()
	if err != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", err)
		panic(err)
	}
}
