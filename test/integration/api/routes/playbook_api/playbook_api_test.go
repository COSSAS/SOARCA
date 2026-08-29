package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	api_routes "soarca/pkg/api"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/decoder"
	"testing"

	playbookservice "soarca/internal/services/playbook"
	mock_playbook "soarca/test/unittest/mocks/mock_playbook_database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func close(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func TestListMeta(t *testing.T) {
	app := gin.New()

	mockPlaybook := new(mock_playbook.MockPlaybook)

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
	mockPlaybook.On("ListMeta", mock.Anything).Return(dummyPlaybookMetas, nil)

	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))
	req, _ := http.NewRequest("GET", "/playbook/meta/", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	mockPlaybook.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlayBookMetas), w.Body.String())
}

func TestList(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	mockPlaybook := new(mock_playbook.MockPlaybook)
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
	mockPlaybook.On("List", mock.Anything).Return(playbooks, nil)
	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))
	req, _ := http.NewRequest("GET", "/playbook/", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	mockPlaybook.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybooks), w.Body.String())
}

func TestGetPlaybookByID(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	mockPlaybook := new(mock_playbook.MockPlaybook)
	dummyPlaybook := decoder.DecodeValidate(byteValue)
	mockPlaybook.On("Get", mock.Anything, dummyPlaybook.ID).Return(*dummyPlaybook, nil)
	marshalledDummyPlaybook, err := json.Marshal(dummyPlaybook)
	if err != nil {
		fmt.Println("Failed to marshall dummy JSON:", err)
		t.Fail()
		return
	}

	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/playbook/%s", dummyPlaybook.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	mockPlaybook.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}

func TestPostPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	mockPlaybook := new(mock_playbook.MockPlaybook)

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
	mockPlaybook.On("Create", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))
	req, _ := http.NewRequest("POST", "/playbook/", bytes.NewBuffer(marshalledDummyPlaybook))
	app.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	mockPlaybook.AssertExpectations(t)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}

func TestDeletePlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	mockPlaybook := new(mock_playbook.MockPlaybook)

	dummyPlaybook := decoder.DecodeValidate(byteValue)
	if dummyPlaybook == nil {
		fmt.Println("got an nil playbook pointer")
		t.Fail()
		return
	}
	mockPlaybook.On("Delete", mock.Anything, dummyPlaybook.ID).Return(nil)
	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/playbook/%s", dummyPlaybook.ID), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUpdatePlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	mockPlaybook := new(mock_playbook.MockPlaybook)

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
	mockPlaybook.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	api_routes.PlaybookRoutesWithService(app, playbookservice.New(mockPlaybook))
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/playbook/%s", dummyPlaybook.ID), bytes.NewBuffer(marshalledDummyPlaybook))
	app.ServeHTTP(w, req)

	mockPlaybook.AssertExpectations(t)
	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, string(marshalledDummyPlaybook), w.Body.String())
}
