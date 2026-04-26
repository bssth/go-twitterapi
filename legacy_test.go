package twitterapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLegacyService_CreateTweet(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/create_tweet",
		body: map[string]any{"data": map[string]any{"create_tweet": map[string]any{}}, "status": "success", "msg": "ok"},
		check: func(t *testing.T, r *http.Request) {
			decodeBody(t, r, &got)
		},
	}))
	resp, err := c.Legacy.CreateTweet(context.Background(), LegacyTweetParams{
		AuthSession: "s", TweetText: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["auth_session"] != "s" {
		t.Errorf("body=%+v", got)
	}
	if resp.Status != StatusSuccess {
		t.Errorf("status=%v", resp.Status)
	}
}

func TestLegacyService_LikeRetweetUpload(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t,
		jsonRoute{method: "POST", path: "/twitter/like_tweet", body: map[string]any{"status": "success", "msg": "ok"}},
		jsonRoute{method: "POST", path: "/twitter/retweet_tweet", body: map[string]any{"status": "success", "msg": "ok"}},
		jsonRoute{method: "POST", path: "/twitter/upload_image", body: map[string]any{"media_id": "m1", "status": "success"}},
	))
	if _, err := c.Legacy.LikeTweet(context.Background(), "s", "1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Legacy.RetweetTweet(context.Background(), "s", "1", ""); err != nil {
		t.Fatal(err)
	}
	resp, err := c.Legacy.UploadImage(context.Background(), "s", "https://example.com/x.png", "")
	if err != nil || resp.MediaID != "m1" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
}

func TestLegacyService_LoginAnd2FA(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t,
		jsonRoute{method: "POST", path: "/twitter/login_by_email_or_username",
			body: map[string]any{"hint": "h", "login_data": map[string]any{"x": 1}, "status": "success"}},
		jsonRoute{method: "POST", path: "/twitter/login_by_2fa",
			body: map[string]any{"session": "S", "user": map[string]any{"id_str": "1", "screen_name": "u", "name": "U"}, "status": "success"}},
	))
	r1, err := c.Legacy.LoginV1(context.Background(), "u", "p", "")
	if err != nil || r1.Hint != "h" {
		t.Fatalf("r1=%+v err=%v", r1, err)
	}
	r2, err := c.Legacy.Login2FA(context.Background(), json.RawMessage(`{"x":1}`), "123456", "")
	if err != nil || r2.Session != "S" || r2.User.ScreenName != "u" {
		t.Fatalf("r2=%+v err=%v", r2, err)
	}
}
