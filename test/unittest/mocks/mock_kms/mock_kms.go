package mock_kms

import (
	"github.com/stretchr/testify/mock"
)

type MockKms struct {
	mock.Mock
}

func (mock *MockKms) Insert(public string, private string, passphrase string, name string) error {
	args := mock.Called(public, private, passphrase, name)
	return args.Error(0)
}

// Update(public string, private string, passphrase string, name string) error
// ListAllNames() ([]string, error)
// Revoke(keyname string) error
