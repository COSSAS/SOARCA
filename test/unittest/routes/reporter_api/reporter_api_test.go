package reporter_test

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
)

func TestGetExecutions(t *testing.T) {
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
	mock_decomposer.On("Execute", *playbook).Return(&decomposer.ExecutionDetails{}, nil)

	recorder := httptest.NewRecorder()
	trigger_api := trigger.New(mock_controller)
	trigger.Routes(app, trigger_api)

	request, err := http.NewRequest("POST", "/trigger/playbook", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	mock_decomposer.AssertExpectations(t)
}
