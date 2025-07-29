package connector

import (
	"encoding/json"
	thehive_models "soarca/pkg/integration/thehive/common/models"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestCreateCase(t *testing.T) {

	host := "http://localhost:9000/thehive/api/v1"
	api := "f2eAPRxxq8Wej7OodikGkmyeottz0xGy"

	conn := NewConnector(host, api)
	assert.NotEqual(t, conn, nil)

	caseModel := thehive_models.Case{Title: "test-title", Description: "some description"}

	body, err := conn.sendRequest("POST", host+"/case", caseModel)
	assert.Equal(t, err, nil)

	response := thehive_models.CaseResponse{}

	err = json.Unmarshal(body, &response)

	assert.Equal(t, err, nil)
	println(response.ID)

}
