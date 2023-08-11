package main

import (

	// docs "github.com/go-project-name/docs"

	routes "soarca/routes"
	"fmt"
)


func main() {
	api := routes.Setup()
	var err = api.Run(":8080")
	if err != nil{
		fmt.Println(err)
	}
}
