package trigger_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"soarca/internal/decomposer"
	"soarca/models/cacao"
	"soarca/routes/trigger"
	"soarca/test/mocks/mock_decomposer"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestExecutionOfPlaybook(t *testing.T) {

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
	var workflow = cacao.Decode(byteValue)
	mock_decomposer.On("Execute", *workflow).Return(&decomposer.ExecutionDetails{}, nil)

	recorder := httptest.NewRecorder()
	trigger_api := trigger.New(mock_decomposer)
	trigger.Routes(app, trigger_api)

	request, err := http.NewRequest("POST", "/trigger/workflow", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	mock_decomposer.AssertExpectations(t)

}
