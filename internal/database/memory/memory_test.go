package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/decoder"
	"sort"
	"testing"

	"github.com/go-playground/assert/v2"
)

var PB_PATH string = "../../../test/playbooks/"

func TestCreate(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)
	mem := New()
	playbook, err := mem.Create(&byteValue)
	assert.Equal(t, err, nil)
	assert.Equal(t, playbook, workflow)

}

func TestRead(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)

	mem := New()
	empty, err := mem.Read(workflow.ID)
	assert.Equal(t, err, errors.New("playbook is not in repository"))
	assert.Equal(t, empty, cacao.Playbook{})

	playbook, err := mem.Create(&byteValue)
	assert.Equal(t, err, nil)
	result, err := mem.Read(playbook.ID)
	assert.Equal(t, err, nil)
	assert.Equal(t, playbook, result)

}

func TestUpdate(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)

	mem := New()
	empty, err := mem.Update(workflow.ID, nil)
	assert.Equal(t, err, errors.New("playbook is not in repository"))
	assert.Equal(t, empty, cacao.Playbook{})

	playbook, err := mem.Create(&byteValue)
	assert.Equal(t, err, nil)

	emptyBytes, err := json.Marshal(new([]byte))
	assert.Equal(t, err, nil)

	parsingFailed, err := mem.Update(playbook.ID, &emptyBytes)
	assert.Equal(t, err, errors.New("failed to decode"))
	assert.Equal(t, parsingFailed, cacao.Playbook{})

	workflow.Description = "new"
	jsonBytes, err := json.Marshal(workflow)
	assert.Equal(t, err, nil)

	result, err := mem.Update(playbook.ID, &jsonBytes)
	assert.Equal(t, err, nil)
	assert.Equal(t, workflow, result)

}

func TestDelete(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)

	mem := New()
	err = mem.Delete(workflow.ID)
	assert.Equal(t, err, nil)

	playbook, err := mem.Create(&byteValue)
	assert.Equal(t, err, nil)
	assert.Equal(t, playbook, workflow)
	readbackPlaybook, err := mem.Read(workflow.ID)
	assert.Equal(t, err, nil)
	assert.Equal(t, readbackPlaybook, workflow)

	err = mem.Delete(workflow.ID)
	assert.Equal(t, err, nil)

	readbackPlaybook2, err := mem.Read(workflow.ID)
	assert.Equal(t, err, errors.New("playbook is not in repository"))
	assert.Equal(t, readbackPlaybook2, cacao.Playbook{})

}

func TestGetAllPlaybooks(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)

	mem := New()

	list := []string{
		"playbook--f47d4081-21ed-4f21-9d05-6b368d73da30",
		"playbook--d41b1046-9334-400b-baa9-91b2ea431731",
		"playbook--586fa554-448a-427e-a719-938f6e033e0d",
		"playbook--3b1e8e64-bb5c-426f-8086-b5a6f9c565e2",
		"playbook--08a82149-d48d-45dd-b8a2-025713d74742",
		"playbook--dcd634a4-04b5-4fa5-842b-bb842a104fbf",
		"playbook--e848e4d2-f529-46e7-8dc4-6ef7acaad902",
		"playbook--e56eb89e-e8ab-41b8-ba94-f97014670bc7",
		"playbook--c2ae6f09-53c5-4e61-b7fb-f5207b2604b3",
		"playbook--8dbdf991-2ec2-45d6-952d-7ef1ac2a9254",
	}
	sort.Strings(list)

	for _, id := range list {
		workflow.ID = id
		jsonBytes, err := json.Marshal(workflow)
		assert.Equal(t, err, nil)
		_, err = mem.Create(&jsonBytes)
		assert.Equal(t, err, nil)
		// assert.Equal(t, playbook, workflow)
	}

	playbooks, err := mem.GetPlaybooks()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(playbooks), 10)

}

func TestGetAllPlaybookMetas(t *testing.T) {
	jsonFile, err := os.Open(PB_PATH + "playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = decoder.DecodeValidate(byteValue)

	mem := New()

	list := []string{
		"playbook--f47d4081-21ed-4f21-9d05-6b368d73da30",
		"playbook--d41b1046-9334-400b-baa9-91b2ea431731",
		"playbook--586fa554-448a-427e-a719-938f6e033e0d",
		"playbook--3b1e8e64-bb5c-426f-8086-b5a6f9c565e2",
		"playbook--08a82149-d48d-45dd-b8a2-025713d74742",
		"playbook--dcd634a4-04b5-4fa5-842b-bb842a104fbf",
		"playbook--e848e4d2-f529-46e7-8dc4-6ef7acaad902",
		"playbook--e56eb89e-e8ab-41b8-ba94-f97014670bc7",
		"playbook--c2ae6f09-53c5-4e61-b7fb-f5207b2604b3",
		"playbook--8dbdf991-2ec2-45d6-952d-7ef1ac2a9254",
	}
	sort.Strings(list)

	for _, id := range list {
		workflow.ID = id
		jsonBytes, err := json.Marshal(workflow)
		assert.Equal(t, err, nil)
		_, err = mem.Create(&jsonBytes)
		assert.Equal(t, err, nil)
		// assert.Equal(t, playbook, workflow)
	}

	playbooks, err := mem.GetPlaybookMetas()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(playbooks), 10)

}
