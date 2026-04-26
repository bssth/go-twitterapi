// Print top trends for a few WOEIDs side-by-side.
//
//	go run ./examples/trends_dashboard
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bssth/go-twitterapi"
)

func main() {
	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	regions := []struct {
		name  string
		woeid int64
	}{
		{"Worldwide", 1},
		{"USA", 23424977},
		{"Japan", 23424856},
		{"UK", 23424975},
	}
	for _, r := range regions {
		fmt.Printf("\n=== %s (woeid=%d) ===\n", r.name, r.woeid)
		trends, err := c.Trends.Get(ctx, r.woeid, 10)
		if err != nil {
			fmt.Printf("  error: %v\n", err)
			continue
		}
		for i, t := range trends {
			vol := ""
			if t.TweetCount > 0 {
				vol = fmt.Sprintf(" (%d)", t.TweetCount)
			}
			fmt.Printf("  %2d. %s%s\n", i+1, t.Name, vol)
		}
	}
}
