package cacao_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"soarca/internal/cacao"
	"testing"
)

func TestCacao(t *testing.T) {
	p := cacao.Playbook{}
	fmt.Println(p)

	jsonFile, err := os.Open("changed-playbook.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var playbook cacao.Playbook
	err = json.Unmarshal(byteValue, &playbook)

	if err !=nil{
		fmt.Println("Not valid JSON")
		return
	}

	for i := 0; i < len(playbook.Workflow); i++ {
		fmt.Println(playbook.Workflow[i].UUID)

	}

}
