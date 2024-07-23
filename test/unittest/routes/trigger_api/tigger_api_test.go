package trigger_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"soarca/internal/decomposer"
	"soarca/models/cacao"
	"soarca/routes/trigger"
	mock_decomposer_controller "soarca/test/unittest/mocks/mock_controller/decomposer"
	"soarca/test/unittest/mocks/mock_decomposer"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestTriggerExecutionOfPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	gin.SetMode(gin.DebugMode)
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)
	playbook := cacao.Decode(byteValue)

	trigger_api := trigger.New(mock_controller)
	recorder := httptest.NewRecorder()
	trigger.Routes(app, trigger_api)

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	mock_decomposer.On("ExecuteAsync", *playbook, trigger_api.Executionsch).Return(&decomposer.ExecutionDetails{}, nil, executionId)

	request, err := http.NewRequest("POST", "/trigger/playbook", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fail()
	}

	expected_return_string := `{"execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","payload":"playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562"}`
	app.ServeHTTP(recorder, request)
	assert.Equal(t, expected_return_string, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)
	mock_decomposer.AssertExpectations(t)
}
