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

	err := conn.CreateCase(caseModel)
	assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}

func TestAddObservableToCase(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api, true)
	assert.NotEqual(t, conn, nil)

	caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	err := conn.CreateCase(caseModel)
	assert.Equal(t, err, nil)

	// response := thehive_models.CaseResponse{}

}
