package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/Davincible/goinsta/v3"
	"github.com/ccarstens/ig-saved-posts/src/config"
	"github.com/ccarstens/ig-saved-posts/src/domain"
	"github.com/ccarstens/ig-saved-posts/src/ui"
)

var readFile domain.ReadFileFn = func(path string) ([]byte, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

var writeFile domain.SaveFileFn = func(data []byte, path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

var syncAll bool

func main() {
	syncAll = getShouldSyncAll()

	config, err := domain.ReadConfig(readFile)
	if config.BasePath == "" {
		basePath, err := ui.PromptBasePath()
		if err != nil {
			panic(err)
		}
		config.BasePath = *basePath
	}

	getUsername := ui.NewGetUsername(config)

	username, err := getUsername.Get()
	if err != nil {
		panic(err)
	}

	insta, err := getClient(*username, config)
	if err != nil {
		panic(err)
	}

	config.ActiveUser = config.GetUserByName(*username)

	defer saveSession(*username, insta, config)

	c := insta.Collections
	if syncAll {
		fmt.Println("Flag -all passed, syncing everything")
	}
	fmt.Println("Fetching Albums now")
	for c.Next() {
		if c.Error() != nil {
			panic(c.Error())
		}
		fmt.Printf("fetched %d more collections\n", c.NumResults)
	}

	var wgDaemon sync.WaitGroup

	tasksIn := make(chan DownloadTask)

	for i := 0; i < 10; i++ {
		wgDaemon.Add(1)
		go downloadDaemon(tasksIn, &wgDaemon)
	}

	collectionDownloadResults := []int{}
	for _, collection := range c.Items[1:] {
		taskCount := downloadCollection(*collection, tasksIn, config.GetDownloadFolder())
		collectionDownloadResults = append(collectionDownloadResults, taskCount)
		if !ShouldContinueBasedOnResults(collectionDownloadResults, syncAll) {
			fmt.Println("it seems like there isn't any new content, aborting execution")
			break
		}
	}

	close(tasksIn)

	wgDaemon.Wait()
}

func configExists() bool {
	_, err := os.Stat("CONFIG_FILE")
	return !errors.Is(err, os.ErrNotExist)
}

type DownloadTask struct {
	Post       goinsta.Item `json:"post"`
	FolderPath string       `json:"folder_path"`
	FileName   string       `json:"file_name"`
}

func getDownloadTasksFromItems(items []goinsta.Item, collectionName string, downloadFolder string) []DownloadTask {
	var result []DownloadTask
	for _, post := range items {
		var postItems []goinsta.Item
		if post.MediaType == 8 {
			postItems = post.CarouselMedia
		} else {
			postItems = []goinsta.Item{post}
		}

		for i, postItem := range postItems {
			var extension string
			if postItem.MediaType == 1 {
				extension = ".jpg"
			} else {
				extension = ".mp4"
			}
			fileName := fmt.Sprintf("%s %s %d%s", postItem.User.Username, post.Code, i, extension)

			folderPath := path.Join(downloadFolder, collectionName)
			fullPath := path.Join(folderPath, fileName)

			if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
				result = append(result, DownloadTask{
					Post:       postItem,
					FolderPath: folderPath,
					FileName:   fileName,
				})
			}
		}

	}
	return result
}

func downloadCollection(c goinsta.Collection, in chan DownloadTask, downloadFolder string) int {
	var loaded int
	fmt.Println("preloading", c.Name)
	downloadTaskCounts := []int{}

	hasMore := true
	var wg sync.WaitGroup
	for hasMore {
		hasMore = c.Next()
		fmt.Printf("loaded %d of %d items for %s\n", loaded+c.NumResults, c.MediaCount, c.Name)
		if err := c.Error(); err != nil && !errors.Is(err, goinsta.ErrNoMore) {
			log.Fatalf("unable to get media for collection %s, %v", c.Name, err)
		}

		itemsToProcess := c.Items[loaded:]
		tasks := getDownloadTasksFromItems(itemsToProcess, c.Name, downloadFolder)
		downloadTaskCounts = append(downloadTaskCounts, len(tasks))
		if len(tasks) == 0 && !ShouldContinueBasedOnResults(downloadTaskCounts, syncAll) {
			fmt.Println(fmt.Sprintf("skipping album %s for lack of new content", c.Name))
			break
		}
		wg.Add(1)
		go func(tasks []DownloadTask) {
			defer wg.Done()
			for _, task := range tasks {
				in <- task
			}
		}(tasks)

		loaded += c.NumResults
		time.Sleep(500 * time.Millisecond)
	}
	wg.Wait()

	var sum int
	for _, r := range downloadTaskCounts {
		sum += r
	}
	return sum
}

func downloadDaemon(in chan DownloadTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range in {
		fmt.Println("Downloading", path.Join(task.FolderPath, task.FileName))
		data, err := task.Post.Download()
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(data)
		}

		fullFileName := path.Join(task.FolderPath, task.FileName)
		err = writeFile(data, fullFileName)
		if err != nil {
			panic(err)
		}
	}
}

func ShouldContinueBasedOnResults(input []int, syncAll bool) bool {
	if syncAll {
		return true
	}

	const numberOfLastZeroesToTrigger = 2
	if len(input) < 2 {
		return true
	}

	i := len(input) - numberOfLastZeroesToTrigger

	var seenNonZero bool

	for _, num := range input[i:] {
		if num != 0 {
			seenNonZero = true
		}
	}

	return seenNonZero
}

func getClient(username string, config *domain.Config) (*goinsta.Instagram, error) {
	fmt.Printf("Login for %s: ", username)
	user := config.GetUserByName(username)
	if user != nil && user.SessionBase64String != "" {
		return getSessionClient(user.SessionBase64String)
	} else {
		password, err := ui.PromptPassword()
		if err != nil {
			return nil, err
		}
		return getNewLoginClient(username, *password, config)
	}
}

func getSessionClient(sessionBase64 string) (*goinsta.Instagram, error) {
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

func getNewLoginClient(user, pw string, config *domain.Config) (*goinsta.Instagram, error) {
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
	defer config.Save(writeFile)
	return insta, nil
}

func saveSession(username string, insta *goinsta.Instagram, cfg *domain.Config) error {
	fmt.Println("Saving Session to", config.GetConfigFilePath())

	sessionString, err := insta.ExportAsBase64String()
	if err != nil {
		return err
	}
	cfg.GetUserByName(username).SessionBase64String = sessionString
	return cfg.Save(writeFile)
}

func getShouldSyncAll() bool {
	all := flag.Bool("all", false, "pass this flag if nothing should be skipped")
	flag.Parse()
	return *all
}
