package twitterapi

import (
	"context"
	"net/http"
	"testing"
)

func TestUsersService_GetByUsername(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user/info",
		body: map[string]any{"data": map[string]any{"userName": "elon", "name": "Elon", "followers": 12345}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"userName": "elon"})
		},
	}))
	u, err := c.Users.GetByUsername(context.Background(), "elon")
	if err != nil {
		t.Fatal(err)
	}
	if u.UserName != "elon" || u.Followers != 12345 {
		t.Fatalf("got %+v", u)
	}
}

func TestUsersService_GetByUsername_RequiresArg(t *testing.T) {
	c, _ := New(Options{APIKey: "k"})
	if _, err := c.Users.GetByUsername(context.Background(), "  "); err == nil {
		t.Fatal("expected error for empty userName")
	}
}

func TestUsersService_About(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user_about",
		body: map[string]any{"data": map[string]any{"id": "1", "userName": "x"}},
	}))
	u, err := c.Users.About(context.Background(), "x")
	if err != nil || u.ID != "1" {
		t.Fatalf("u=%+v err=%v", u, err)
	}
}

func TestUsersService_BatchByIDs(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user/batch_info_by_ids",
		body: map[string]any{"users": []any{
			map[string]any{"id": "1"}, map[string]any{"id": "2"},
		}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"userIds": "1,2,3"})
		},
	}))
	users, err := c.Users.BatchByIDs(context.Background(), []string{"1", "2", "3"})
	if err != nil || len(users) != 2 {
		t.Fatalf("err=%v users=%v", err, users)
	}
}

func TestUsersService_BatchByIDs_Empty(t *testing.T) {
	c, _ := New(Options{APIKey: "k"})
	if _, err := c.Users.BatchByIDs(context.Background(), nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestUsersService_FollowersIterator_TwoPages(t *testing.T) {
	page := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twitter/user/followers" {
			t.Fatalf("path %q", r.URL.Path)
		}
		page++
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{"followers":[{"id":"1"},{"id":"2"}],"has_next_page":true,"next_cursor":"c1"}`))
		case 2:
			if r.URL.Query().Get("cursor") != "c1" {
				t.Errorf("missing cursor: %q", r.URL.Query().Get("cursor"))
			}
			_, _ = w.Write([]byte(`{"followers":[{"id":"3"}],"has_next_page":false}`))
		default:
			t.Fatalf("unexpected page %d", page)
		}
	})
	got := []string{}
	it := c.Users.Followers(context.Background(), "elon", &FollowersOpts{PageSize: 200})
	for it.Next() {
		got = append(got, it.Item().ID)
	}
	if it.Err() != nil {
		t.Fatal(it.Err())
	}
	if len(got) != 3 || got[2] != "3" {
		t.Fatalf("got %v", got)
	}
}

func TestUsersService_Search(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user/search",
		body: map[string]any{"users": []any{map[string]any{"id": "1"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"query": "go developers"})
		},
	}))
	resp, err := c.Users.SearchPage(context.Background(), "go developers", nil)
	if err != nil || len(resp.Users) != 1 {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}

func TestUsersService_Mentions_Pagination(t *testing.T) {
	page := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		if r.URL.Path != "/twitter/user/mentions" {
			t.Fatalf("path %q", r.URL.Path)
		}
		if page == 1 {
			_, _ = w.Write([]byte(`{"tweets":[{"id":"a"}],"has_next_page":true,"next_cursor":"x"}`))
			return
		}
		_, _ = w.Write([]byte(`{"tweets":[{"id":"b"}],"has_next_page":false}`))
	})
	since := int64(1700000000)
	until := int64(1700001000)
	it := c.Users.Mentions(context.Background(), "elon", &MentionsOpts{SinceTimeUnix: since, UntilTimeUnix: until})
	got := []string{}
	for it.Next() {
		got = append(got, it.Item().ID)
	}
	if it.Err() != nil || len(got) != 2 {
		t.Fatalf("got=%v err=%v", got, it.Err())
	}
}

func TestUsersService_CheckFollow(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user/check_follow_relationship",
		body: map[string]any{"data": map[string]any{"following": true, "followed_by": false}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"source_user_name": "a", "target_user_name": "b"})
		},
	}))
	rel, err := c.Users.CheckFollow(context.Background(), "a", "b")
	if err != nil || !rel.Following || rel.FollowedBy {
		t.Fatalf("rel=%+v err=%v", rel, err)
	}
}

func TestUsersService_LastTweets_RequiresIDOrName(t *testing.T) {
	c, _ := New(Options{APIKey: "k"})
	if _, err := c.Users.LastTweetsPage(context.Background(), LastTweetsOpts{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUsersService_VerifiedFollowers(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/user/verifiedFollowers",
		body: map[string]any{"followers": []any{map[string]any{"id": "1"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"user_id": "12345"})
		},
	}))
	resp, err := c.Users.VerifiedFollowersPage(context.Background(), "12345", "")
	if err != nil || len(resp.Followers) != 1 {
		t.Fatalf("err=%v resp=%+v", err, resp)
	}
}
