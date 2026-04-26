package twitterapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WSEvent is one message off the experimental tweet-stream WebSocket. The raw
// JSON is preserved in Raw for callers that want fields not modeled here.
type WSEvent struct {
	EventType string          `json:"event_type"`
	RuleID    string          `json:"rule_id,omitempty"`
	RuleTag   string          `json:"rule_tag,omitempty"`
	Tweets    json.RawMessage `json:"tweets,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

// DecodeTweets unmarshals the embedded tweets array into typed Tweets.
func (e WSEvent) DecodeTweets() ([]Tweet, error) {
	if len(e.Tweets) == 0 {
		return nil, nil
	}
	var tw []Tweet
	if err := json.Unmarshal(e.Tweets, &tw); err != nil {
		return nil, err
	}
	return tw, nil
}

// WSClient is an experimental WebSocket client for the dashboard-registered
// filter rules (see WebhookService). Reconnects automatically on read errors;
// returns when the context is canceled or onEvent panics.
//
// The endpoint is not officially documented; it ships with the SDK because
// it's the lowest-latency delivery path. Treat it as best-effort.
type WSClient struct {
	APIKey         string
	URL            string        // defaults to DefaultWSURL
	ReconnectDelay time.Duration // default 90s
	HandshakeTime  time.Duration // default 15s

	// Logger is invoked on dial errors and read errors. Optional.
	Logger func(format string, v ...any)
}

// NewWSClient is a convenience constructor pulling APIKey + URL off a Client.
func NewWSClient(c *Client) *WSClient {
	return &WSClient{APIKey: c.APIKey(), URL: c.WSURL()}
}

// ConnectAndRead connects to the stream and invokes onEvent for every
// message until ctx is canceled. Auto-reconnects after ReconnectDelay on
// transient errors.
func (w *WSClient) ConnectAndRead(ctx context.Context, onEvent func(WSEvent)) error {
	if w.APIKey == "" {
		return ErrMissingAPIKey
	}
	u := w.URL
	if u == "" {
		u = DefaultWSURL
	}
	delay := w.ReconnectDelay
	if delay <= 0 {
		delay = 90 * time.Second
	}
	hs := w.HandshakeTime
	if hs <= 0 {
		hs = 15 * time.Second
	}

	d := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: hs,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		h := http.Header{}
		h.Set("x-api-key", w.APIKey)
		h.Set("X-API-Key", w.APIKey)

		conn, _, err := d.DialContext(ctx, u, h)
		if err != nil {
			w.logf("dial: %v", err)
			if waitErr := sleepCtx(ctx, delay); waitErr != nil {
				return waitErr
			}
			continue
		}

		readErr := w.readLoop(ctx, conn, onEvent)
		if readErr == nil || errors.Is(readErr, context.Canceled) {
			return readErr
		}
		w.logf("read: %v", readErr)
		if waitErr := sleepCtx(ctx, delay); waitErr != nil {
			return waitErr
		}
	}
}

func (w *WSClient) readLoop(ctx context.Context, conn *websocket.Conn, onEvent func(WSEvent)) error {
	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var ev WSEvent
		ev.Raw = append([]byte(nil), msg...)
		_ = json.Unmarshal(msg, &ev)
		if onEvent != nil {
			onEvent(ev)
		}
	}
}

func (w *WSClient) logf(format string, v ...any) {
	if w.Logger != nil {
		w.Logger(format, v...)
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
