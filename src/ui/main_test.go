package ui

import (
	"testing"

	"github.com/ccarstens/ig-saved-posts/src/domain"
	"github.com/stretchr/testify/assert"
)

func Test_GetUsername(t *testing.T) {
	user_a := "user_a"
	tests := map[string]struct {
		*domain.Config
		ValueSources     map[string]domain.GetValue
		expectedUsername *string
		expectedError    assert.ErrorAssertionFunc
	}{
		"happy path - user from config": {
			Config: &domain.Config{
				Users: []domain.User{
					{
						Name:                "user_a",
						SessionBase64String: "a1b2c3",
					},
				},
			},
			ValueSources: map[string]domain.GetValue{
				"SELECT_NAME": func() (*string, error) {
					name := "user_a"
					return &name, nil
				},
				"PROMPT_NAME": nil,
			},
			expectedUsername: &user_a,
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		"happy path - new user from prompt": {
			Config: &domain.Config{
				Users: nil,
			},
			ValueSources: map[string]domain.GetValue{
				"SELECT_NAME": nil,
				"PROMPT_NAME": func() (*string, error) {
					return &user_a, nil
				},
			},
			expectedUsername: &user_a,
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
		"happy path - if user selects to add user, result from prompt is returned": {
			Config: &domain.Config{
				Users: []domain.User{
					{
						Name:                "user_b",
						SessionBase64String: "a1b2c3",
					},
				},
			},
			ValueSources: map[string]domain.GetValue{
				"SELECT_NAME": func() (*string, error) {
					res := addUser
					return &res, nil
				},
				"PROMPT_NAME": func() (*string, error) {
					return &user_a, nil
				},
			},
			expectedUsername: &user_a,
			expectedError: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			input := GetUsername{
				Config:       test.Config,
				ValueSources: test.ValueSources,
			}

			username, err := input.Get()
			if !test.expectedError(t, err) {
				return
			}

			assert.Equal(t, test.expectedUsername, username)
		})
	}
}
