// Reconstruct the full reply chain root -> ... -> startTweetID.
//
//	go run ./examples/thread_chain 1234567890123456789
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bssth/go-twitterapi"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: thread_chain <tweet_id>")
	}
	id := os.Args[1]

	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	tweets, err := c.Tweets.ByIDs(ctx, []string{id})
	if err != nil || len(tweets) == 0 {
		log.Fatalf("lookup: %v", err)
	}
	chain, err := c.Tweets.ReplyChainToRoot(ctx, tweets[0], &twitterapi.ReplyChainOpts{MaxContextPages: 5})
	if err != nil {
		log.Fatal(err)
	}
	for i, t := range chain {
		fmt.Printf("%d. @%s [%s]\n   %s\n", i+1, t.Author.UserName, t.ID, t.Text)
	}
}
