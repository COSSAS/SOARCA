package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"soarca/models/cacao"
	workflow_router "soarca/routes/workflow"
	workflow_mock "soarca/test/mocks/workflow"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Only id field is required for testing the functionality
type Workflow struct {
	ID string `json:"id"`
}

const jsonTestWorkflow = `{
    "type": "playbook",
    "spec_version": "cacao-2.0",
    "id": "playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562",
    "name": "SOARCA Main Flow",
    "description": "This playbook will run for each trigger event in SOARCA",
    "playbook_types": [
        "notification"
    ],
    "created_by": "identity--5abe695c-7bd5-4c31-8824-2528696cdbf1",
    "created": "2023-05-26T15:56:00.123456Z",
    "modified": "2023-05-26T15:56:00.123456Z",
    "valid_from": "2023-05-26T15:56:00.123456Z",
    "valid_until": "2337-05-26T15:56:00.123456Z",
    "priority": 1,
    "severity": 1,
    "impact": 1,
    "labels": [
        "soarca"
    ],
    "authentication_info_definitions": {
        "http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae": {
            "type": "http-basic",
            "user_id": "admin",
            "password": "super-secure-password"
        }
    },
    "external_references": [
        {
            "name": "TNO SOARCA",
            "description": "SOARCA Homepage",
            "source": "TNO CST",
            "url": "http://tno.nl/cst"
        }
    ],
    "workflow_start": "step--a76dbc32-b739-427b-ae13-4ec703d5797e",
    "workflow_exception": "step--40131926-89e9-44df-a018-5f92f2df7914",
    "workflow": {
        "step--a76dbc32-b739-427b-ae13-4ec703d5797e": {
            "type": "action",
            "name": "IMC assets by CVE",
            "description": "Check the IMC for affected assets by CVE",
            "on_completion": "step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4",
            "commands": [
                {
                    "type": "http-api",
                    "command": "GET http://__imc_address__/by/__cve__"
                }
            ]
        },
        "step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4": {
            "type": "action",
            "name": "BIA for CVE",
            "description": "Perform Business Impact Analysis for CVE",
            "on_completion": "step--09b97fab-56a1-45dc-a88f-be3cde3eac33",
            "commands": [
                {
                    "type": "http-api",
                    "command": "GET http://__bia_address__/analysisreport/__cve__"
                }
            ]
        },
        "step--09b97fab-56a1-45dc-a88f-be3cde3eac33": {
            "type": "action",
            "name": "Generate CoAs",
            "description": "Generate Courses of Action",
            "on_completion": "step--2190f685-1857-44ac-ad0e-0ded6c6ef3ce",
            "commands": [
                {
                    "type": "http-api",
                    "command": "GET http://__coagenerator_address__/coa/__assetuuid__"
                }
            ],
            "target": []
        },
        "step--2190f685-1857-44ac-ad0e-0ded6c6ef3ce": {
            "type": "action",
            "name": "BIA for CoAs",
            "description": "Perform Business Impact Analysis for Courses of Action",
            "on_completion": "end--6b23c237-ade8-4d00-9aa1-75999738d557",
            "commands": [
                {
                    "type": "http-api",
                    "command": "GET http://__bia_address__/analysisreport/__coa_list__"
                }
            ]
        },
        "end--6b23c237-ade8-4d00-9aa1-75999738d557": {
            "type": "end",
            "name": "End SOARCA Main Flow"
        }
    }
}`

func TestGetIDs(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)
	testWorkflow.On("GetWorkflowIds").Return([]string{"1", "2", "3"}, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("GET", "/workflow/", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, `["1", "2", "3"]`, w.Body.String())
}

func TestGetByID(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyWorkflow := cacao.Decode([]byte(jsonTestWorkflow))
	testWorkflow.On("Read", dummyWorkflow.ID).Return(*dummyWorkflow, nil)
	marshalledWorkflow, err := json.Marshal(dummyWorkflow)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/workflow/%s", dummyWorkflow.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	testWorkflow.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledWorkflow), w.Body.String())
}

func TestPostWorkflow(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyWorkflow := cacao.Decode([]byte(jsonTestWorkflow))
	marshalledWorkflow, err := json.Marshal(dummyWorkflow)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}

	testWorkflow.On("Create", &marshalledWorkflow).Return(dummyWorkflow.ID, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("POST", "/workflow/", bytes.NewBuffer(marshalledWorkflow))
	app.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	testWorkflow.AssertExpectations(t)
	jsonResult := `{
		"workflow-id": "%s",
		"workflowType": ""
	}`
	assert.JSONEq(t, fmt.Sprintf(jsonResult, dummyWorkflow.ID), w.Body.String())
}

func TestDeleteWorkflow(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyWorkflow := cacao.Decode([]byte(jsonTestWorkflow))
	testWorkflow.On("Delete", dummyWorkflow.ID).Return(nil)
	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/workflow/%s", dummyWorkflow.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateWorkflow(t *testing.T) {
	app := gin.New()
	testWorkflow := new(workflow_mock.MockWorkflow)

	dummyWorkflow := cacao.Decode([]byte(jsonTestWorkflow))
	marshalledWorkflow, err := json.Marshal(dummyWorkflow)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}
	pointerDummyObject := []byte(marshalledWorkflow)
	testWorkflow.On("Update", dummyWorkflow.ID, &pointerDummyObject).Return(*dummyWorkflow, nil)

	w := httptest.NewRecorder()
	workflow_router.Routes(app, testWorkflow)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/workflow/%s", dummyWorkflow.ID), bytes.NewBuffer(marshalledWorkflow))
	app.ServeHTTP(w, req)

	testWorkflow.AssertExpectations(t)
	assert.Equal(t, 200, w.Code)
	fmt.Println(w.Body.String())
	assert.JSONEq(t, string(marshalledWorkflow), w.Body.String())
}
