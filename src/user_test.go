package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testConfigOne = `
{
  "base_path": "~/Pictures/Mood",
  "users": [
    {
      "name": "Test User",
      "session": "a1b2c3"
    }
  ]
}`
)

var capturedWrite string
var capturedPath string

func Test_ReadConfig(t *testing.T) {
	tests := map[string]struct {
		ReadFileFn
		expectedConfig *Config
		expectedError  assert.ErrorAssertionFunc
	}{
		"happy path": {
			ReadFileFn: func(path string) ([]byte, error) {
				return []byte(testConfigOne), nil
			},
			expectedConfig: &Config{
				BasePath: "~/Pictures/Mood",
				Users: []User{
					{
						Name:                "Test User",
						SessionBase64String: "a1b2c3",
					},
				},
			},
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(tt, err)
			},
		},
		"error on read": {
			ReadFileFn: func(path string) ([]byte, error) {
				return nil, fmt.Errorf("error")
			},
			expectedConfig: nil,
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config, err := ReadConfig(test.ReadFileFn)
			if !test.expectedError(t, err) {
				return
			}
			assert.Equal(t, test.expectedConfig, config)

		})
	}
}

func Test_SaveConfig(t *testing.T) {
	tests := map[string]struct {
		SaveFileFn
		Config
		expectedFilePath string
		expectedError    assert.ErrorAssertionFunc
	}{
		"happy path": {
			SaveFileFn: func(file []byte, path string) error {
				capturedPath = path
				capturedWrite = string(file)
				return nil
			},
			Config: Config{
				BasePath: "~/Pictures/Mood",
				Users: []User{
					{
						Name:                "Test User",
						SessionBase64String: "a1b2c3",
					},
				},
			},
			expectedFilePath: CONFIG_FILE,
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			capturedPath = ""
			capturedWrite = ""
			err := test.Config.Save(test.SaveFileFn)
			if !test.expectedError(t, err) {
				return
			}

			assert.Equal(t, test.expectedFilePath, capturedPath)
			assert.NotEmpty(t, capturedWrite)

		})
	}
}
