package main

import (
	"fmt"
	"soarca/internal/lib1"
)

func main() {
	fmt.Println("Let's do some soarca")
	some := lib1.Somestruct{Name: "test"}
	fmt.Println(some.Name)
}
