package cacao_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	cacao "soarca/models/cacao"
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

	var workflow cacao.Playbook
	err = json.Unmarshal(byteValue, &workflow)

	if err !=nil{
		fmt.Println("Not valid JSON")
		return
	}

	for i := 0; i < len(workflow.Workflow); i++ {
		fmt.Println(workflow.Workflow[i].UUID)

	}

}
