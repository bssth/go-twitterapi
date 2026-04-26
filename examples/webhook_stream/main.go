// Webhook + WS streaming example: register a filter rule, then read tweets
// off the experimental WebSocket. Activate the rule beforehand via UpdateRule
// (or the dashboard).
//
//	export TWITTERAPIIO_API_KEY=...
//	go run ./examples/webhook_stream "from:elonmusk OR from:openai"
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bssth/go-twitterapi"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: webhook_stream \"<advanced search query>\"")
	}
	query := os.Args[1]

	c, err := twitterapi.New(twitterapi.Options{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rule, err := c.Webhook.AddRule(ctx, "demo-stream", query, 60)
	if err != nil {
		log.Fatalf("add rule: %v", err)
	}
	fmt.Printf("created rule %s — activating...\n", rule.RuleID)

	if _, err := c.Webhook.UpdateRule(ctx, rule.RuleID, "demo-stream", query, 60, true); err != nil {
		log.Fatalf("activate rule: %v", err)
	}

	defer func() {
		_, _ = c.Webhook.DeleteRule(context.Background(), rule.RuleID)
	}()

	ws := twitterapi.NewWSClient(c)
	ws.Logger = log.Printf
	err = ws.ConnectAndRead(ctx, func(ev twitterapi.WSEvent) {
		tweets, _ := ev.DecodeTweets()
		for _, t := range tweets {
			fmt.Printf("[%s] @%s — %s\n", ev.RuleTag, t.Author.UserName, t.Text)
		}
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
