package main

import (

	// docs "github.com/go-project-name/docs"

	"fmt"
	routes "soarca/routes"
)

func main() {
	api := routes.Setup()
	var err = api.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
