package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Davincible/goinsta/v3"
	"github.com/manifoldco/promptui"
)

const (
	CONFIG_FILE = "~/ig-saved-posts/config.json"
	BASE_FOLDER = "./albums"
)

func main() {

	usernamePrompt := promptui.Prompt{
		Label:    "Username",
		Validate: func(s string) error { return nil },
	}

	pwPrompt := promptui.Prompt{
		Label:    "Password",
		Validate: func(s string) error { return nil },
		Mask:     '*',
	}

	username, err := usernamePrompt.Run()
	fmt.Printf("Username: %s\n", username)

	password, err := pwPrompt.Run()
	fmt.Printf("Password: %s\n", password)

	return
	insta, err := getInsta()
	if err != nil {
		panic(err)
	}

	c := insta.Collections
	fmt.Println("fetching now")
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
		taskCount := downloadCollection(*collection, tasksIn, 0)
		collectionDownloadResults = append(collectionDownloadResults, taskCount)
		if !ShouldContinueBasedOnResults(collectionDownloadResults) {
			fmt.Println("it seems like there isn't any new content, aborting execution")
			break
		}
	}

	close(tasksIn)

	wgDaemon.Wait()
}

func configExists() bool {
	_, err := os.Stat(CONFIG_FILE)
	return !errors.Is(err, os.ErrNotExist)
}

func getInsta() (*goinsta.Instagram, error) {
	if configExists() {
		fmt.Println("config exists")
		insta, err := goinsta.Import(CONFIG_FILE)

		if err != nil {
			return nil, err
		}

		err = insta.OpenApp()
		if err != nil {
			return nil, err
		}

		return insta, nil
	}

	insta := goinsta.New("", "")
	err := insta.Login()
	if err != nil {
		panic(err)
	}
	defer insta.Export("./config.json")
	return insta, nil
}

type DownloadTask struct {
	Post       goinsta.Item `json:"post"`
	FolderPath string       `json:"folder_path"`
	FileName   string       `json:"file_name"`
}

func getDownloadTasksFromItems(items []goinsta.Item, collectionName string) []DownloadTask {
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

			folderPath := path.Join(BASE_FOLDER, collectionName)
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

func downloadCollection(c goinsta.Collection, in chan DownloadTask, i int) int {
	var loaded int
	fmt.Println("preloading", c.Name)
	downloadTaskCounts := []int{}

	hasMore := true
	var wg sync.WaitGroup
	for hasMore {
		hasMore = c.Next()
		fmt.Printf("loaded %d of %d items for %s (%d)\n", loaded+c.NumResults, c.MediaCount, c.Name, i)
		if err := c.Error(); err != nil && !errors.Is(err, goinsta.ErrNoMore) {
			log.Fatalf("unable to get media for collection %s, %v", c.Name, err)
		}

		itemsToProcess := c.Items[loaded:]
		tasks := getDownloadTasksFromItems(itemsToProcess, c.Name)
		downloadTaskCounts = append(downloadTaskCounts, len(tasks))
		if len(tasks) == 0 && !ShouldContinueBasedOnResults(downloadTaskCounts) {
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
		fmt.Println("downloading", path.Join(task.FolderPath, task.FileName))
		data, err := task.Post.Download()
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(data)
		}

		os.MkdirAll(task.FolderPath, os.ModePerm)

		fullFileName := path.Join(task.FolderPath, task.FileName)
		file, err := os.Create(fullFileName)
		file.Write(data)
	}
}

func ShouldContinueBasedOnResults(input []int) bool {
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
