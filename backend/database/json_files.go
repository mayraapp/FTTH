package database

import (
	"errors"
	"os"
	"path/filepath"
)

func Read(path string) (contents []byte, exists bool, err error) {
	contents, err = os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return contents, err == nil, err
}

func Write(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	return os.WriteFile(path, contents, 0600)
}
