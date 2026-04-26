// Manage the user-stream monitor list.
//
//	go run ./examples/monitor list
//	go run ./examples/monitor add @elonmusk
//	go run ./examples/monitor remove <id_for_user>
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
		log.Fatal("usage: monitor list|add|remove ...")
	}
	cmd := os.Args[1]

	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	switch cmd {
	case "list":
		users, err := c.Monitor.List(ctx, twitterapi.MonitorAll)
		if err != nil {
			log.Fatal(err)
		}
		for _, u := range users {
			fmt.Printf("%-12s  @%s  tweet=%v profile=%v\n",
				u.IDForUser, u.XUserName, u.IsMonitorTweet, u.IsMonitorProfile)
		}
	case "add":
		if len(os.Args) < 3 {
			log.Fatal("usage: monitor add <username>")
		}
		resp, err := c.Monitor.AddUser(ctx, os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("added:", resp.Message())
	case "remove":
		if len(os.Args) < 3 {
			log.Fatal("usage: monitor remove <id_for_user>")
		}
		resp, err := c.Monitor.RemoveUser(ctx, os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("removed:", resp.Message())
	default:
		log.Fatalf("unknown command %q", cmd)
	}
}
