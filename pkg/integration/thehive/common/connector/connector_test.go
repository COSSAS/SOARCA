package connector

import (
	thehive_models "soarca/pkg/integration/thehive/common/models"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestCreateCase(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	caseId, err := conn.CreateCase(caseModel)
	assert.Equal(t, err, nil)

	assert.NotEqual(t, caseId, "")
	// response := thehive_models.CaseResponse{}

}

func TestAddObservableToCase(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	caseId, err := conn.CreateCase(caseModel)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, caseId, "")
	observable := thehive_models.Observable{DataType: "ip", Data: []string{"10.0.0.10"}}
	conn.CreateObservable(caseId, observable)
	assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}

func TestGetAllCases(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	// caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	err := conn.GetAllCases()
	assert.Equal(t, err, nil)
	// assert.NotEqual(t, caseId, "")
	// observable := thehive_models.Observable{DataType: "ip", Data: []string{"10.0.0.10"}}
	// conn.CreateObservable(caseId, observable)
	// assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}

func TestGetCaseById(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	// caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	caseObj, err := conn.GetCaseById("~147616")
	assert.Equal(t, err, nil)
	assert.NotEqual(t, caseObj, "")
	// println(caseObj.ExtraData.Data)
	// observable := thehive_models.Observable{DataType: "ip", Data: []string{"10.0.0.10"}}
	// conn.CreateObservable(caseId, observable)
	// assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}

func TestGetAllObservables(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	// caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	caseObj, err := conn.GetObservableFromCase("~147616")
	assert.Equal(t, err, nil)
	assert.NotEqual(t, caseObj, "")
	// println(caseObj.ExtraData.Data)
	// observable := thehive_models.Observable{DataType: "ip", Data: []string{"10.0.0.10"}}
	// conn.CreateObservable(caseId, observable)
	// assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}
