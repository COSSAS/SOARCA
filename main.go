package main

import (
	"soarca/application"
	"soarca/logger"

	"github.com/joho/godotenv"
)

var log *logger.Log

func init() {
	log = logger.Logger("MAIN", logger.Info, "", logger.Json)
}

func main() {
	errenv := godotenv.Load(".env.example")
	if errenv != nil {
		log.Fatal("Failed to read env variable")
	}

	errinit := application.InitialiseAppComponents()
	if errinit != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", errinit)
		panic(errinit)
	}
}
