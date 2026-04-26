// Advanced tweet search with bounded pagination.
//
//	go run ./examples/search "from:openai since:2026-01-01" 50
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/bssth/go-twitterapi"
	"github.com/bssth/go-twitterapi/examples/internal/console"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: search \"<query>\" [limit=100]")
	}
	query := os.Args[1]
	limit := 100
	if len(os.Args) > 2 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			limit = n
		}
	}

	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}

	it := c.Tweets.AdvancedSearch(context.Background(), query, &twitterapi.AdvancedSearchOpts{
		QueryType: "Latest",
	})
	it.MaxItems = limit

	for it.Next() {
		t := it.Item()
		fmt.Printf("[%s] @%-20s likes=%-5d %s\n",
			t.CreatedAt, t.Author.UserName, t.LikeCount, console.OneLine(t.Text))
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
}
