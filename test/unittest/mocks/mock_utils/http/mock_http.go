package mock_executor

import (
	"soarca/utils/http"

	"github.com/stretchr/testify/mock"
)

type MockHttpOptions struct {
	mock.Mock
}

type MockHttpRequest struct {
	mock.Mock
}

func (httpOptions *MockHttpOptions) ExtractUrl() (string, error) {
	args := httpOptions.Called()
	return args.String(0), args.Error(1)
}

func (httpOptions *MockHttpRequest) Request(options http.HttpOptions) ([]byte, error) {
	args := httpOptions.Called(options)
	return args.Get(0).([]byte), args.Error(1)
}
