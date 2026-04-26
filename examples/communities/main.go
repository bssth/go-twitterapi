// Print a community's info and the first 50 members.
//
//	go run ./examples/communities 1234567890
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
		log.Fatal("usage: communities <community_id>")
	}
	cid := os.Args[1]

	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	info, err := c.Communities.Info(ctx, cid)
	if err != nil {
		log.Fatalf("info: %v", err)
	}
	fmt.Printf("Community: %s\n  members=%d  moderators=%d  nsfw=%v\n  %s\n\n",
		info.Name, info.MemberCount, info.ModeratorCount, info.IsNSFW, info.Description)

	fmt.Println("First 50 members:")
	it := c.Communities.Members(ctx, cid)
	it.MaxItems = 50
	for it.Next() {
		u := it.Item()
		fmt.Printf("  @%-20s  %s\n", u.UserName, u.Name)
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
}
