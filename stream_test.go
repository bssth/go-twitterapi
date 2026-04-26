package twitterapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWSClient_RequiresAPIKey(t *testing.T) {
	w := &WSClient{}
	err := w.ConnectAndRead(context.Background(), nil)
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("got %v", err)
	}
}

func TestWSClient_ReadsAndReconnects(t *testing.T) {
	var connections int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "k" {
			http.Error(w, "no key", http.StatusUnauthorized)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		n := atomic.AddInt32(&connections, 1)
		if n == 1 {
			// First connection: send one event then close — forces a reconnect.
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"tweet","rule_tag":"t1","tweets":[{"id":"a"}]}`))
			_ = conn.Close()
			return
		}
		// Second connection: another event, then leave open until ctx expires.
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"tweet","rule_tag":"t2","tweets":[{"id":"b"}]}`))
		<-time.After(200 * time.Millisecond)
		_ = conn.Close()
	}))
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)

	w := &WSClient{
		APIKey:         "k",
		URL:            wsURL,
		ReconnectDelay: 5 * time.Millisecond,
		HandshakeTime:  2 * time.Second,
	}

	got := make(chan WSEvent, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	go func() {
		_ = w.ConnectAndRead(ctx, func(ev WSEvent) {
			got <- ev
		})
	}()

	tags := []string{}
	for len(tags) < 2 {
		select {
		case ev := <-got:
			tags = append(tags, ev.RuleTag)
			tweets, _ := ev.DecodeTweets()
			if len(tweets) != 1 {
				t.Errorf("expected 1 tweet, got %d", len(tweets))
			}
		case <-ctx.Done():
			t.Fatalf("timeout, got tags=%v conns=%d", tags, atomic.LoadInt32(&connections))
		}
	}
	if tags[0] != "t1" || tags[1] != "t2" {
		t.Fatalf("tags=%v", tags)
	}
	if atomic.LoadInt32(&connections) < 2 {
		t.Fatalf("expected reconnect, conns=%d", atomic.LoadInt32(&connections))
	}
}

func TestWSEvent_DecodeTweets_EmptyOK(t *testing.T) {
	ev := WSEvent{}
	tw, err := ev.DecodeTweets()
	if err != nil || tw != nil {
		t.Fatalf("tw=%v err=%v", tw, err)
	}
}
