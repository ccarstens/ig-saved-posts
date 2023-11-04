package main

import (
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"github.com/ccarstens/ig-saved-posts/src/domain"
	"github.com/ccarstens/ig-saved-posts/src/ig"
	"github.com/ccarstens/ig-saved-posts/src/io"
	"github.com/ccarstens/ig-saved-posts/src/ui"
	"time"
)

func main() {

	config, err := domain.ReadConfig(io.ReadFile)
	if config.BasePath == "" {
		basePath, err := ui.PromptBasePath()
		if err != nil {
			panic(err)
		}
		config.BasePath = *basePath
	}
	//
	//getUsername := ui.NewGetUsername(config)
	//
	//username, err := getUsername.Get()
	//if err != nil {
	//	panic(err)
	//}
	//

	username := "luv0151"

	insta, err := ig.GetClient(username, config)
	if err != nil {
		panic(err)
	}
	fmt.Println("logged in user is", username)

	config.ActiveUser = config.GetUserByName(username)

	following := insta.Account.Following("imilenaio", goinsta.DefaultOrder)
	following.Next()

	testUser := following.Users[0]

	testUserFeed := testUser.Feed()
	testUserFeed.Next()

	firstPost := testUserFeed.Items[0]
	fmt.Println(firstPost.Caption.Text)
	time.Sleep(3 * time.Second)

	//err = firstPost.SyncLikers()
	//if err != nil {
	//	panic(err)
	//}
	//likers := firstPost.Likers
	//firstLiker, err := insta.VisitProfile(likers[0].Username)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("first liker is", firstLiker.User.Username)
	//
	//firstLikerFeed := firstLiker.Feed
	//
	//firstLikerFeed.Next()
	//fmt.Println("count", len(firstLikerFeed.Items))

	//err = firstLikerFeed.Sync()
	//if err != nil {
	//	panic(err)
	//}

	//var items []*goinsta.Item
	//for firstLikerFeed.Next() {
	//	fmt.Println("fetching")
	//	time.Sleep(2 * time.Second)
	//	items = append(items, firstLikerFeed.Items...)
	//}
	//
	//fmt.Println(items[1].Caption.Text)
	//fmt.Println("count: ", len(items))

	//return

	comments := firstPost.Comments
	comments.Sync()

	for comments.Next() {
		c := comments.Items[0]
		creation := time.Unix(c.CreatedAtUtc, 0)
		fmt.Println(c.User.Username, "commented", c.Text, "on", creation)
		//if !c.HasLikedComment {
		//	time.Sleep(1 * time.Second)
		//	err = c.Like()
		//	fmt.Println("just liked comment", c.Text, "by user", c.User.Username)
		//	if err != nil {
		//		panic(err)
		//	}
		//} else {
		//	fmt.Println("comment", c.Text, "by user", c.User.Username, "was already liked by me, unliking")
		//	err = c.Unlike()
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//}

		break
	}

	defer ig.SaveSession(username, insta, config)
}
