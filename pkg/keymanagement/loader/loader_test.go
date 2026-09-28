package loader

import (
	"errors"
	"soarca/test/unittest/mocks/mock_kms"
	"testing"

	"github.com/go-playground/assert/v2"
)

var key = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBs4o06933hic/ArsSo0fs9cUTk0AHc2vON1ZqS68Vf7gAAAJhCJkxRQiZM
UQAAAAtzc2gtZWQyNTUxOQAAACBs4o06933hic/ArsSo0fs9cUTk0AHc2vON1ZqS68Vf7g
AAAEBI7PvrlDTFnWuMJHGaLuUkulSpH/Ni378Y2vLZcpldxmzijTr3feGJz8CuxKjR+z1x
ROTQAdza843VmpLrxV/uAAAADnRoaWpzQFBDLTQ0MzE4AQIDBAUGBw==
-----END OPENSSH PRIVATE KEY-----
`

var publicKey = `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGzijTr3feGJz8CuxKjR+z1xROTQAdza843VmpLrxV/u user@PC
`

func TestLoadKeysFromNonExistingDirectory(t *testing.T) {
	kmsApi := mock_kms.MockKms{}

	kmsApi.On("Insert", key, publicKey, "", "test-key").Return(nil)
	err := Load("non-exist", &kmsApi)
	assert.Equal(t, err.Error(), "stat non-exist: no such file or directory")
}

func TestLoadKeys(t *testing.T) {
	kmsApi := mock_kms.MockKms{}

	kmsApi.On("Insert", publicKey, key, "", "test-key").Return(nil)
	err := Load("keys", &kmsApi)
	assert.Equal(t, err, nil)
}

func TestLoadingNonSshKeys(t *testing.T) {
	kmsApi := mock_kms.MockKms{}

	kmsApi.On("Insert", publicKey, key, "", "test-key").Return(errors.New("no correct key file"))
	err := Load("nokeys", &kmsApi)
	assert.Equal(t, err, nil)
}
