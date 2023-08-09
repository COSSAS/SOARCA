package cacao_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"soarca/internal/cacao"
	"testing"
)

func TestAbs(t *testing.T) {
	p := cacao.Playbook{}
	fmt.Println(p)

	jsonFile, err := os.Open("changed-playbook.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var playbook cacao.Playbook
	json.Unmarshal(byteValue, &playbook)

	for i := 0; i < len(playbook.Workflow); i++ {
		fmt.Println(playbook.Workflow[i].UUID)

	}

}
