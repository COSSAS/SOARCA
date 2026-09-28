package loader

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"soarca/internal/logger"
	"strings"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type IKms interface {
	Insert(public string, private string, passphrase string, name string) error
}

func Load(keyDirectoryPath string, api IKms) error {
	info, err := os.Stat(keyDirectoryPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("provided path is not a dir: " + err.Error())
	}
	err = loadDir(keyDirectoryPath, api)
	if err != nil {
		return err
	}
	return nil

}

// LoadDir reads every regular file directly inside dir and returns
// a map of file name to file contents. Subdirectories are skipped.
func loadDir(dir string, api IKms) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue // skip subdirectories, symlinks, sockets
		}
		if strings.HasSuffix(entry.Name(), ".pub") {
			continue // skip the public files as we load them with the private ones.
		}
		private, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		public, err := os.ReadFile(filepath.Join(dir, entry.Name()+".pub"))
		if err != nil {
			continue
		}
		err = api.Insert(string(public), string(private), "", entry.Name())
		if err != nil {
			log.Error("failed to insert key into store")
		}

	}
	return nil
}
