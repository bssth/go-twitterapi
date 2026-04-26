// Send a DM (and read history). Requires login env + proxy.
//
//	go run ./examples/dm 1234567890 "hi from the SDK"
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
	if len(os.Args) < 3 {
		log.Fatal("usage: dm <recipient-user-id> <text...>")
	}
	userID := os.Args[1]
	text := strings.Join(os.Args[2:], " ")

	c, err := twitterapi.New(twitterapi.Options{TokenFile: "twitterapiio.token.json"})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resp, err := c.Actions.SendDM(ctx, twitterapi.SendDMParams{UserID: userID, Text: text})
	if err != nil {
		log.Fatalf("send: %v", err)
	}
	fmt.Printf("delivered: message_id=%s\n", resp.MessageID)

	hist, err := c.Actions.DMHistory(ctx, userID, "")
	if err != nil {
		log.Printf("history: %v", err)
		return
	}
	fmt.Printf("history (%d bytes):\n%s\n", len(hist), string(hist))
}
