package main

import (
	app "soarca/app"
	"soarca/logger"

	"github.com/joho/godotenv"
)

var log *logger.Log

func init() {
	log = logger.Logger("MAIN", logger.Info, "", logger.Json)
}

func main() {
	err := godotenv.Load(".env.example")
	if err != nil {
		log.Fatal("Failed to read env variable")
	}
	app.LoadComponent()
	err = app.SetupAndRunApp()
	if err != nil {
		log.Fatal("Something Went wrong with setting-up the app, msg: ", err)
		panic(err)
	}
}
