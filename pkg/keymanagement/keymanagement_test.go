package keymanagement

import (
	"os"
	"path"
	"slices"
	"soarca/internal/database/memory"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testkey string = "test-key"

func testkey_dir() string {
	return path.Join("..", "..", "test", "unittest", "mocks", "mock_utils")
}

var globalKeyManagement *KeyManagement

func init() {
	globalKeyManagement = InitKeyManagement(memory.NewKeyManagementDatabase())
}

func TestRevoke(t *testing.T) {
	addKey(t, testkey)
	assert.True(t, slices.Contains(allNames(t), testkey))
	assert.Nil(t, globalKeyManagement.Revoke(testkey))
	assert.False(t, slices.Contains(allNames(t), testkey))
}
func addKey(t *testing.T, keyname string) {
	pubkey_path := path.Join(testkey_dir(), "test-key.pub")
	privkey_path := path.Join(testkey_dir(), "test-key")
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
	assert.Nil(t, globalKeyManagement.Insert(string(pubkey_buf), string(privkey_buf), "", keyname))
	assert.Nil(t, privkey_file.Close())
	assert.Nil(t, pubkey_file.Close())
}

func allNames(t *testing.T) []string {
	allnames, err := globalKeyManagement.ListAllNames()
	assert.Nil(t, err)
	return allnames
}

func TestAddKey(t *testing.T) {
	addKey(t, testkey)
	assert.True(t, slices.Equal(allNames(t), []string{testkey}))
}
