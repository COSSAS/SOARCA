package ssh

import (
	"os"
	"path"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testdir string = ".ssh_test"
const testkey string = "test-key"

func testkey_dir() string {
	return path.Join("..", "..", "..", "..", "test", "unittest", "mocks", "mock_utils")
}

func init() {
	err := os.Mkdir(testdir, 0777)
	if err != nil {
		if err.(*os.PathError).Err.Error() == "file exists" {
			if err := os.RemoveAll(testdir); err != nil {
				panic(err)
			}
			if err := os.Mkdir(testdir, 0777); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	if _, err = InitKeyManagement(testdir); err != nil {
		panic(err)
	}
}

func TestRevoke(t *testing.T) {
	addKey(t, testkey)
	assert.True(t, slices.Contains(globalKeyManagement.ListAllNames(), testkey))
	assert.Nil(t, globalKeyManagement.Revoke(testkey))
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
	assert.Nil(t, globalKeyManagement.Insert(string(pubkey_buf), string(privkey_buf), keyname))
	assert.Nil(t, privkey_file.Close())
	assert.Nil(t, pubkey_file.Close())
}

func TestAddKey(t *testing.T) {
	addKey(t, testkey)
	assert.True(t, slices.Equal(globalKeyManagement.ListAllNames(), []string{testkey}))
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0666)
}
func TestRefresh(t *testing.T) {
	assert.False(t, slices.Contains(globalKeyManagement.ListAllNames(), testkey+"1"))
	pubkey_path := path.Join(testkey_dir(), "test-key.pub")
	privkey_path := path.Join(testkey_dir(), "test-key")
	assert.Nil(t, copyFile(pubkey_path, path.Join(testdir, testkey+"1.pub")))
	assert.Nil(t, copyFile(privkey_path, path.Join(testdir, testkey+"1")))
	assert.Nil(t, globalKeyManagement.Refresh())
	assert.True(t, slices.Contains(globalKeyManagement.ListAllNames(), testkey+"1"))
}
