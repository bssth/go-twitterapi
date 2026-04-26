package twitterapi

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestActionsService_CreateTweet_AttachesAuth(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/create_tweet_v2",
		body: map[string]any{"tweet_id": "100", "status": "success", "msg": "ok"},
		check: func(t *testing.T, r *http.Request) {
			decodeBody(t, r, &got)
		},
	}), func(o *Options) { o.LoginCookie = "COOKIE" })

	resp, err := c.Actions.CreateTweet(context.Background(), CreateTweetParams{
		TweetText: "hello — there",
	})
	if err != nil || resp.TweetID != "100" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	if got["login_cookies"] != "COOKIE" {
		t.Errorf("missing login_cookies: %+v", got)
	}
	if got["proxy"] != "http://proxy.example:8080" {
		t.Errorf("missing proxy: %+v", got)
	}
	if got["tweet_text"] != "hello - there" {
		t.Errorf("sanitize not applied: %q", got["tweet_text"])
	}
}

func TestActionsService_CreateTweet_SkipSanitize(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/create_tweet_v2",
		body: map[string]any{"tweet_id": "1", "status": "success"},
		check: func(t *testing.T, r *http.Request) {
			decodeBody(t, r, &got)
		},
	}), func(o *Options) { o.LoginCookie = "C" })

	_, err := c.Actions.CreateTweet(context.Background(), CreateTweetParams{
		TweetText:    "raw — text",
		SkipSanitize: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got["tweet_text"].(string), "—") {
		t.Errorf("sanitize applied even with SkipSanitize: %q", got["tweet_text"])
	}
}

func TestActionsService_CreateTweet_RequiresText(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "C", DefaultProxy: "p"})
	if _, err := c.Actions.CreateTweet(context.Background(), CreateTweetParams{}); err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestActionsService_CookieRefresh_OnExpiry(t *testing.T) {
	var calls int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/twitter/like_tweet_v2":
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				_, _ = w.Write([]byte(`{"status":"error","msg":"login_cookies invalid"}`))
				return
			}
			// On retry, expect the *fresh* cookie.
			var b map[string]any
			decodeBody(t, r, &b)
			if b["login_cookies"] != "FRESH" {
				t.Errorf("retry used stale cookie: %+v", b)
			}
			_, _ = w.Write([]byte(`{"status":"success","msg":"ok"}`))
		case "/twitter/user_login_v2":
			_, _ = w.Write([]byte(`{"login_cookies":"FRESH","status":"success"}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}, func(o *Options) { o.LoginCookie = "STALE" })

	t.Setenv("TWITTERAPIIO_USER_NAME", "u")
	t.Setenv("TWITTERAPIIO_EMAIL", "e")
	t.Setenv("TWITTERAPIIO_PASSWORD", "p")

	if _, err := c.Actions.LikeTweet(context.Background(), "1", "http://p"); err != nil {
		t.Fatalf("LikeTweet: %v", err)
	}
	if c.LoginCookie() != "FRESH" {
		t.Fatalf("cookie not refreshed: %q", c.LoginCookie())
	}
}

func TestActionsService_RequiresProxy(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "C"}) // no proxy
	t.Setenv("TWITTERAPIIO_PROXY", "")
	if _, err := c.Actions.LikeTweet(context.Background(), "1", ""); err == nil {
		t.Fatal("expected error for missing proxy")
	}
}

func TestActionsService_BookmarksIterator(t *testing.T) {
	page := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twitter/bookmarks_v2" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		page++
		if page == 1 {
			_, _ = w.Write([]byte(`{"tweets":[{"id":"1"}],"has_next_page":true,"next_cursor":"x"}`))
			return
		}
		_, _ = w.Write([]byte(`{"tweets":[{"id":"2"}],"has_next_page":false}`))
	}, func(o *Options) { o.LoginCookie = "C" })
	it := c.Actions.Bookmarks(context.Background(), BookmarksOpts{})
	got, err := it.All(0)
	if err != nil || len(got) != 2 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestActionsService_Report_Validates(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "C", DefaultProxy: "p"})
	if _, err := c.Actions.Report(context.Background(), ReportParams{Reason: ReportSpam}); err == nil {
		t.Fatal("expected error: needs tweet/user id")
	}
	if _, err := c.Actions.Report(context.Background(), ReportParams{TweetID: "1", UserID: "2", Reason: ReportSpam}); err == nil {
		t.Fatal("expected error: both ids set")
	}
	if _, err := c.Actions.Report(context.Background(), ReportParams{TweetID: "1"}); err == nil {
		t.Fatal("expected error: missing reason")
	}
}

func TestActionsService_SendDM(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/send_dm_to_user",
		body: map[string]any{"message_id": "M", "status": "success"},
		check: func(t *testing.T, r *http.Request) {
			decodeBody(t, r, &got)
		},
	}), func(o *Options) { o.LoginCookie = "C" })
	resp, err := c.Actions.SendDM(context.Background(), SendDMParams{
		UserID: "1", Text: "hi", MediaID: "m1",
	})
	if err != nil || resp.MessageID != "M" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	if got["media_id"] != "m1" {
		t.Errorf("media_id not in body: %+v", got)
	}
}

func TestActionsService_CommunityCreateAndLeave(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t,
		jsonRoute{
			method: "POST", path: "/twitter/create_community_v2",
			body: map[string]any{"community_id": "c1", "status": "success"},
		},
		jsonRoute{
			method: "POST", path: "/twitter/leave_community_v2",
			body: map[string]any{"community_id": "c1", "community_name": "n", "status": "success"},
		},
	), func(o *Options) { o.LoginCookie = "C" })

	cr, err := c.Actions.CreateCommunity(context.Background(), "n", "d", "")
	if err != nil || cr.CommunityID != "c1" {
		t.Fatalf("cr=%+v err=%v", cr, err)
	}
	lv, err := c.Actions.LeaveCommunity(context.Background(), "c1", "")
	if err != nil || lv.CommunityName != "n" {
		t.Fatalf("lv=%+v err=%v", lv, err)
	}
}
