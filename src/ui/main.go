package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ccarstens/ig-saved-posts/src/domain"
	"github.com/manifoldco/promptui"
)

const (
	selectName     = "SELECT_NAME"
	promptName     = "PROMPT_NAME"
	addUser        = "Add User"
	promptPw       = "PROMPT_PASSWORD"
	promptBasePath = "PROMPT_BASE_PATH"
)

func NewGetUsername(config *domain.Config) GetUsername {
	element := GetUsername{
		Config: config,
		ValueSources: map[string]domain.GetValue{
			promptName: promptUsername,
		},
	}
	element.ValueSources[selectName] = element.selectUsername
	return element
}

type GetUsername struct {
	Config       *domain.Config
	ValueSources map[string]domain.GetValue
}

func (g GetUsername) Get() (*string, error) {
	if len(g.Config.Users) == 0 {
		source := g.ValueSources[promptName]
		return source()
	} else {
		source := g.ValueSources[selectName]
		result, err := source()

		if err != nil {
			return nil, err
		}

		if *result == addUser {
			source = g.ValueSources[promptName]
			return source()
		}
		return result, nil
	}
}

func (g GetUsername) selectUsername() (*string, error) {
	names := []string{}
	for _, user := range g.Config.Users {
		names = append(names, user.Name)
	}
	names = append(names, addUser)

	p := promptui.Select{
		Label: labels[selectName],
		Items: names,
	}

	_, value, err := p.Run()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func promptUsername() (*string, error) {
	p := promptui.Prompt{
		Label: labels[promptName],
	}

	value, err := p.Run()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func PromptPassword() (*string, error) {
	p := promptui.Prompt{
		Label: labels[promptPw],
		Mask:  '*',
	}
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func PromptBasePath() (*string, error) {
	p := promptui.Prompt{
		Label: labels[promptBasePath],
	}
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(result, "~/") {
		home, _ := os.UserHomeDir()
		result = filepath.Join(home, result[2:])
	}

	return &result, nil
}
