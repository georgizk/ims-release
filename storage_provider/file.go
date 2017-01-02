package storage_provider

import "os"
import "io/ioutil"
import "path"
import "path/filepath"
import "errors"

type File struct {
	Root string
}

func (sp *File) Set(key string, data []byte) error {
	if sp.Exists(key) {
		return errors.New("key " + key + " already exists")
	}
	filePath := sp.getPath(key)
	// @TODO should make sure that generated path does not go outside of Root
	// should also make sure that the path name is valid (for use in archives etc.)
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModeDir|0700)
	if err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err == nil {
		_, err = f.Write(data)
	}
	defer f.Close()
	return err
}

func (sp *File) Get(key string) ([]byte, error) {
	filePath := sp.getPath(key)
	f, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (sp *File) Unset(key string) error {
	filePath := sp.getPath(key)
	return os.Remove(filePath)
}

func (sp *File) getPath(key string) string {
	return path.Join(sp.Root, key)
}

func (sp *File) Exists(key string) bool {
	filePath := sp.getPath(key)
	_, err := os.Stat(filePath)

	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
