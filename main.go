package main

import (
	"fmt"
	"soarca/application"
	"soarca/logger"

	"github.com/joho/godotenv"
)

var log *logger.Log

func init() {
	log = logger.Logger("MAIN", logger.Info, "", logger.Json)
}

var Version string
var Buildtime string

const banner = `
   _____  ____          _____   _____          
  / ____|/ __ \   /\   |  __ \ / ____|   /\    
 | (___ | |  | | /  \  | |__) | |       /  \   
  \___ \| |  | |/ /\ \ |  _  /| |      / /\ \  
  ____) | |__| / ____ \| | \ \| |____ / ____ \ 
 |_____/ \____/_/    \_\_|  \_\\_____/_/    \_\
                                               
                                               

`

func main() {
	fmt.Print(banner)
	log.Info("Version: ", Version)
	log.Info("Buildtime: ", Buildtime)

	errenv := godotenv.Load(".env")
	if errenv != nil {
		log.Warning("Failed to read env variable, but will continue")
	}

	errinit := application.InitialiseAppComponents()
	if errinit != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", errinit)
		panic(errinit)
	}
}
