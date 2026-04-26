// Upload a local file as media, then post a tweet referencing it.
//
//	go run ./examples/upload_media path/to/photo.jpg "caption"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bssth/go-twitterapi"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: upload_media <file> <text>")
	}
	path := os.Args[1]
	text := os.Args[2]

	c, err := twitterapi.New(twitterapi.Options{TokenFile: "twitterapiio.token.json"})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	up, err := c.Media.Upload(ctx, filepath.Base(path), f, nil)
	if err != nil {
		log.Fatalf("upload: %v", err)
	}
	fmt.Printf("uploaded media_id=%s\n", up.MediaID)

	resp, err := c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{
		TweetText: text,
		MediaIDs:  []string{up.MediaID},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("posted: https://x.com/i/web/status/%s\n", resp.TweetID)
}
