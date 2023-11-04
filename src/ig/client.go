package ig

import (
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"github.com/ccarstens/ig-saved-posts/src/config"
	"github.com/ccarstens/ig-saved-posts/src/domain"
	"github.com/ccarstens/ig-saved-posts/src/io"
	"github.com/ccarstens/ig-saved-posts/src/ui"
)

func GetClient(username string, config *domain.Config) (*goinsta.Instagram, error) {
	fmt.Printf("Login for %s: ", username)
	user := config.GetUserByName(username)
	if user != nil && user.SessionBase64String != "" {
		return GetSessionClient(user.SessionBase64String)
	} else {
		password, err := ui.PromptPassword()
		if err != nil {
			return nil, err
		}
		return GetNewLoginClient(username, *password, config)
	}
}

func GetSessionClient(sessionBase64 string) (*goinsta.Instagram, error) {
	fmt.Print("Reusing Session\n")
	insta, err := goinsta.ImportFromBase64String(sessionBase64)

	if err != nil {
		return nil, err
	}

	err = insta.OpenApp()
	if err != nil {
		return nil, err
	}

	return insta, nil
}

func GetNewLoginClient(user, pw string, config *domain.Config) (*goinsta.Instagram, error) {
	fmt.Print("Creating new Session\n")
	insta := goinsta.New(user, pw)
	err := insta.Login()
	if err != nil {
		panic(err)
	}
	sessionString, err := insta.ExportAsBase64String()
	if err != nil {
		return nil, err
	}
	config.Users = append(config.Users, domain.User{
		Name:                user,
		SessionBase64String: sessionString,
	})
	defer config.Save(io.WriteFile)
	return insta, nil
}

func SaveSession(username string, insta *goinsta.Instagram, cfg *domain.Config) error {
	fmt.Println("Saving Session to", config.GetConfigFilePath())

	sessionString, err := insta.ExportAsBase64String()
	if err != nil {
		return err
	}
	cfg.GetUserByName(username).SessionBase64String = sessionString
	return cfg.Save(io.WriteFile)
}
