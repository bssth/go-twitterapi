// Posting example: log in once via env vars, post a tweet.
//
//	export TWITTERAPIIO_API_KEY=...
//	export TWITTERAPIIO_USER_NAME=...
//	export TWITTERAPIIO_EMAIL=...
//	export TWITTERAPIIO_PASSWORD=...
//	export TWITTERAPIIO_PROXY=http://user:pass@host:port
//	export TWITTERAPIIO_TOTP_SECRET=...   # optional
//	go run ./examples/post_tweet "hello from the SDK"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bssth/go-twitterapi"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: post_tweet \"text...\"")
	}
	text := strings.Join(os.Args[1:], " ")

	c, err := twitterapi.New(twitterapi.Options{
		// Persist the cookie between runs.
		TokenFile: "twitterapiio.token.json",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	resp, err := c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{
		TweetText: text,
	})
	if err != nil {
		log.Fatalf("tweet: %v", err)
	}
	fmt.Printf("posted: https://x.com/i/web/status/%s\n", resp.TweetID)
}
