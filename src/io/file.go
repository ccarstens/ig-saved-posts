package io

import (
	"errors"
	"github.com/ccarstens/ig-saved-posts/src/domain"
	"os"
	"path/filepath"
)

var ReadFile domain.ReadFileFn = func(path string) ([]byte, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

var WriteFile domain.SaveFileFn = func(data []byte, path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}
