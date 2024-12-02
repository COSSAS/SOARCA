package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"soarca/pkg/models/cacao"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
)

var PB_PATH string = "../../../test/playbooks/"

func TestNotValidCacaoJsonInvalidAgentTargetType(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "invalid_playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	errValid := IsValidCacaoJson(byteValue)
	if errValid == nil {
		t.Fail()
	}

	t.Log(errValid)
	expected := "value must be 'http-api'"
	assert.Equal(t, strings.Contains(fmt.Sprint(errValid), expected), true)

}

func TestValidCacaoJson(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	errValidation := IsValidCacaoJson(byteValue)
	if errValidation != nil {
		fmt.Println(err)
		t.Fail()
	}

}

func TestValidWorkflow(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)
	var playbook cacao.Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		t.Fail()
	}
	errSafeWorkflow := IsSafeCacaoWorkflow(&playbook)
	assert.Equal(t, errSafeWorkflow, nil)
}

func TestIsSafeCacaoWorkflowFailMissingStep(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "missing_step_playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)

	var playbook cacao.Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		t.Fail()
	}

	errSafeWorkflow := IsSafeCacaoWorkflow(&playbook)

	expected := errors.New(
		"step end--6b23c237-ade8-4d00-9aa1-75999738d558 does not exist")

	assert.Equal(t, errSafeWorkflow, expected)

}

func TestIsSafeCacaoWorkflowFailInfinite(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "infinite_playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)

	var playbook cacao.Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		t.Fail()
	}

	errSafeWorkflow := IsSafeCacaoWorkflow(&playbook)

	expected := "worflow seems to loop on branch sequence"

	assert.Equal(t, strings.Contains(fmt.Sprint(errSafeWorkflow), expected), true)
}

func TestIsSafeCacaoWorkflowFailAgentEmail(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "invalid_email_playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)

	var playbook cacao.Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		t.Fail()
	}

	errSafeWorkflow := IsSafeCacaoWorkflow(&playbook)
	fmt.Println(errSafeWorkflow)

	expected := "invalid email"
	assert.Equal(t, strings.Contains(fmt.Sprint(errSafeWorkflow), expected), true)
}

func TestIsSafeCacaoWorkflow(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "parallels_playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)
	var playbook cacao.Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		t.Fail()
	}
	errSafeWorkflow := IsSafeCacaoWorkflow(&playbook)

	assert.Equal(t, errSafeWorkflow, nil)

}
