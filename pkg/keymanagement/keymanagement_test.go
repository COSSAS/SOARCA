package keymanagement

import (
	"os"
	"path"
	"slices"
	memory "soarca/internal/database/memory/keymanagement"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testkey string = "test-key"
const testPath = "../../test/unittest/mocks/mock_utils"

func TestRevoke(t *testing.T) {
	keyManagement := New(memory.New())

	pubkey_path := path.Join(testPath, "test-key.pub")
	privkey_path := path.Join(testPath, "test-key")
	pubkey_file, err := os.Open(pubkey_path)
	assert.Nil(t, err)

	privkey_file, err := os.Open(privkey_path)
	assert.Nil(t, err)
	pubkey_buf := make([]byte, 2048)
	privkey_buf := make([]byte, 2048)
	_, err = pubkey_file.Read(pubkey_buf)
	assert.Nil(t, err)
	_, err = privkey_file.Read(privkey_buf)
	assert.Nil(t, err)
	assert.Nil(t, keyManagement.Insert(string(pubkey_buf), string(privkey_buf), "", testkey))
	assert.Nil(t, privkey_file.Close())
	assert.Nil(t, pubkey_file.Close())

	allNames, err := keyManagement.ListAllNames()
	assert.Nil(t, err)
	assert.True(t, slices.Contains(allNames, testkey))
	assert.Nil(t, keyManagement.Revoke(testkey))
	allNames, err = keyManagement.ListAllNames()
	assert.Nil(t, err)
	assert.Equal(t, false, slices.Contains(allNames, testkey))
}

func TestAddKey(t *testing.T) {
	keyManagement := New(memory.New())

	pubkey_path := path.Join(testPath, "test-key.pub")
	privkey_path := path.Join(testPath, "test-key")
	pubkey_file, err := os.Open(pubkey_path)
	assert.Nil(t, err)

	privkey_file, err := os.Open(privkey_path)
	assert.Nil(t, err)
	pubkey_buf := make([]byte, 2048)
	privkey_buf := make([]byte, 2048)
	_, err = pubkey_file.Read(pubkey_buf)
	assert.Nil(t, err)
	_, err = privkey_file.Read(privkey_buf)
	assert.Nil(t, err)
	assert.Nil(t, keyManagement.Insert(string(pubkey_buf), string(privkey_buf), "", testkey))
	assert.Nil(t, privkey_file.Close())
	assert.Nil(t, pubkey_file.Close())

	allNames, err := keyManagement.ListAllNames()
	assert.Nil(t, err)
	assert.True(t, slices.Equal(allNames, []string{testkey}))
}
