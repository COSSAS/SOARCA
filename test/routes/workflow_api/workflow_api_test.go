package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"soarca/models/api"
	"soarca/models/cacao"
	"soarca/models/decoder"
	workflow_router "soarca/routes/workflow"
	workflow_mock "soarca/test/mocks/workflow"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Only id field is required for testing the functionality
type Workflow struct {
	ID string `json:"id"`
}

const jsonTestPlayBookMeta = `{
	"id": "playbook--300270f9-0e64-42c8-93cc-0927edbe3ae7",
	"name": "Block malware",
	"description": "This playbook will block malware by performing multiple actions",
	"valid_from": "2023-11-20T15:56:00.123Z",
	"valid_until": "2123-11-20T15:56:00.123Z",
	"labels": [
		"soarca",
		"coa9",
		"coa7"
		]
	}`

func TestGetMetas(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	var dummyPlaybookMeta api.PlaybookMeta

	if err := json.Unmarshal([]byte(jsonTestPlayBookMeta), &dummyPlaybookMeta); err != nil {
		fmt.Println("Failed to unmarshall test playbookmeta type:", err)
		t.Fail()
		return
	}

	secondDummyPlaybookMeta := dummyPlaybookMeta
	secondDummyPlaybookMeta.ID = "playbook-more-random--300270f9-0e64-42dddc8-93cc-0927edbe3ae7"
	secondDummyPlaybookMeta.Description = "just a second random playbook for testing"

	dummyPlaybookMetas := []api.PlaybookMeta{dummyPlaybookMeta, secondDummyPlaybookMeta}
	marshalledDummyPlayBookMetas, err := json.Marshal(dummyPlaybookMetas)
	if err != nil {
		fmt.Println("failed to marshall test playbookmeta type byte array", err)
		t.Fail()
		return
	}
	testWorkflow.On("GetWorkflowMetas").Return(dummyPlaybookMetas, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("GET", "/workflow/meta/", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlayBookMetas), w.Body.String())
}

func TestGetPlaybooks(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	testWorkflow := new(workflow_mock.MockWorkflow)
	dummyPlaybook := decoder.DecodeValidate(byteValue)
	if dummyPlaybook == nil {
		fmt.Println("got an nil playbook pointer")
		t.Fail()
		return
	}

	secondDummyPlaybook := *dummyPlaybook
	secondDummyPlaybook.ID = "playbook-more-random--300270f9-0e64-42dddc8-93cc-0927edbe3ae7"
	secondDummyPlaybook.Description = "just a second random playbook for testing"

	playbooks := []cacao.Playbook{*dummyPlaybook, secondDummyPlaybook}
	marshalledDummyPlaybooks, err := json.Marshal(playbooks)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}
	testWorkflow.On("GetWorkflows").Return(playbooks, nil)
	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("GET", "/workflow/", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybooks), w.Body.String())
}

func TestGetByID(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)
	dummyPlaybook := decoder.DecodeValidate(byteValue)
	testWorkflow.On("Read", dummyPlaybook.ID).Return(*dummyPlaybook, nil)
	marshalledDummyPlaybook, err := json.Marshal(dummyPlaybook)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/workflow/%s", dummyPlaybook.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}

func TestPostWorkflow(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyPlaybook := decoder.DecodeValidate(byteValue)
	if dummyPlaybook == nil {
		fmt.Println("got an nil playbook pointer")
		t.Fail()
		return
	}
	marshalledDummyPlaybook, err := json.Marshal(dummyPlaybook)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}
	pointerDummyObject := []byte(marshalledDummyPlaybook)
	testWorkflow.On("Create", &pointerDummyObject).Return(*dummyPlaybook, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("POST", "/workflow/", bytes.NewBuffer(marshalledDummyPlaybook))
	app.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}

func TestDeleteWorkflow(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyPlaybook := decoder.DecodeValidate(byteValue)
	if dummyPlaybook == nil {
		fmt.Println("got an nil playbook pointer")
		t.Fail()
		return
	}
	testWorkflow.On("Delete", dummyPlaybook.ID).Return(nil)
	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/workflow/%s", dummyPlaybook.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateWorkflow(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyPlaybook := decoder.DecodeValidate(byteValue)
	if dummyPlaybook == nil {
		fmt.Println("got an nil playbook pointer")
		t.Fail()
		return
	}
	marshalledDummyPlaybook, err := json.Marshal(dummyPlaybook)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}
	pointerDummyObject := []byte(marshalledDummyPlaybook)
	testWorkflow.On("Update", dummyPlaybook.ID, &pointerDummyObject).Return(*dummyPlaybook, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/workflow/%s", dummyPlaybook.ID), bytes.NewBuffer(marshalledDummyPlaybook))
	app.ServeHTTP(w, req)

	testWorkflow.AssertExpectations(t)
	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}
