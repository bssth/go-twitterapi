// Basic read example: profile + last 5 tweets.
//
//	export TWITTERAPIIO_API_KEY=...
//	go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bssth/go-twitterapi"
	"github.com/bssth/go-twitterapi/examples/internal/console"
)

func main() {
	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := c.Users.GetByUsername(ctx, "elonmusk")
	if err != nil {
		log.Fatalf("user lookup: %v", err)
	}
	fmt.Printf("@%s — %s — %d followers\n", user.UserName, user.Name, user.Followers)

	it := c.Users.LastTweets(ctx, twitterapi.LastTweetsOpts{UserName: user.UserName})
	it.MaxItems = 5
	for it.Next() {
		t := it.Item()
		fmt.Printf("  %s  %s\n", t.CreatedAt, console.OneLine(t.Text, 80))
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
}
