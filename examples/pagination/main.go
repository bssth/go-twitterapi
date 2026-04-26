// Pagination example: walk every follower of a small account, then a search.
//
//	export TWITTERAPIIO_API_KEY=...
//	go run ./examples/pagination
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bssth/go-twitterapi"
	"github.com/bssth/go-twitterapi/examples/internal/console"
)

func main() {
	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Bound follower walking — full pages cost real money.
	it := c.Users.Followers(ctx, "twitter", &twitterapi.FollowersOpts{PageSize: 200})
	it.MaxItems = 1000
	count := 0
	for it.Next() {
		count++
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("walked %d followers (capped)\n", count)

	// Search with a limit.
	search := c.Tweets.AdvancedSearch(ctx, "from:openai since:2026-01-01", &twitterapi.AdvancedSearchOpts{QueryType: "Latest"})
	search.MaxPages = 2
	for search.Next() {
		t := search.Item()
		fmt.Printf("  [%s] %s\n", t.ID, console.OneLine(t.Text))
	}
	if err := search.Err(); err != nil {
		log.Fatal(err)
	}
}
