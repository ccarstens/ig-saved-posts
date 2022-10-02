package main

import (
	"encoding/json"
)

type SaveFileFn func(file []byte, path string) error
type ReadFileFn func(path string) ([]byte, error)

type User struct {
	Name                string `json:"name"`
	SessionBase64String string `json:"session"`
}

type Config struct {
	Users    []User `json:"users"`
	BasePath string `json:"base_path"`
}

func (c *Config) Save(save SaveFileFn) error {
	data, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	return save(data, CONFIG_FILE)
}

func ReadConfig(read ReadFileFn) (*Config, error) {
	data, err := read(CONFIG_FILE)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
