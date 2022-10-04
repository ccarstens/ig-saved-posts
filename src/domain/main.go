package domain

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/ccarstens/ig-saved-posts/src/config"
)

type SaveFileFn func(file []byte, path string) error
type ReadFileFn func(path string) ([]byte, error)
type GetValue func() (*string, error)

type User struct {
	Name                string `json:"name"`
	SessionBase64String string `json:"session"`
}

type Config struct {
	Users      []User `json:"users"`
	BasePath   string `json:"base_path"`
	ActiveUser *User  `json:"-"`
}

func (c *Config) Save(save SaveFileFn) error {
	data, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	return save(data, config.GetConfigFilePath())
}

func (c *Config) GetUserByName(name string) *User {
	for i, user := range c.Users {
		if user.Name == name {
			return &c.Users[i]
		}
	}
	return nil
}

func (c *Config) GetDownloadFolder() string {
	return path.Join(c.BasePath, c.ActiveUser.Name, "albums")
}

func ReadConfig(read ReadFileFn) (*Config, error) {
	data, err := read(config.GetConfigFilePath())
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	} else if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
