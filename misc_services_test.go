package twitterapi

import (
	"context"
	"net/http"
	"testing"
)

func TestCommunitiesService_Info(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/community/info",
		body: map[string]any{"community_info": map[string]any{"id": "c1", "name": "go", "member_count": 100}},
	}))
	info, err := c.Communities.Info(context.Background(), "c1")
	if err != nil || info.ID != "c1" || info.MemberCount != 100 {
		t.Fatalf("info=%+v err=%v", info, err)
	}
}

func TestCommunitiesService_MembersIterator(t *testing.T) {
	page := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twitter/community/members" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		page++
		if page == 1 {
			_, _ = w.Write([]byte(`{"members":[{"id":"a"},{"id":"b"}],"has_next_page":true,"next_cursor":"x"}`))
			return
		}
		_, _ = w.Write([]byte(`{"members":[{"id":"c"}],"has_next_page":false}`))
	})
	it := c.Communities.Members(context.Background(), "c1")
	got, err := it.All(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
}

func TestCommunitiesService_SearchAll(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/community/get_tweets_from_all_community",
		body: map[string]any{"tweets": []any{map[string]any{"id": "t"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"query": "go", "queryType": "Top"})
		},
	}))
	it := c.Communities.SearchAll(context.Background(), "go", &AllCommunitySearchOpts{QueryType: "Top"})
	got, err := it.All(0)
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestSpacesService_Detail(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/spaces/detail",
		body: map[string]any{"data": map[string]any{"id": "s1", "title": "talks", "state": "live"}},
	}))
	s, err := c.Spaces.Detail(context.Background(), "s1")
	if err != nil || s.ID != "s1" || s.Title != "talks" {
		t.Fatalf("s=%+v err=%v", s, err)
	}
}

func TestTrendsService_Get(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/trends",
		body: map[string]any{"trends": []any{
			map[string]any{"name": "Go", "tweet_volume": 12345},
			map[string]any{"name": "Rust"},
		}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"woeid": "1", "count": "30"})
		},
	}))
	got, err := c.Trends.Get(context.Background(), 1, 30)
	if err != nil || len(got) != 2 || got[0].Name != "Go" || got[0].TweetCount != 12345 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestListsService_Tweets(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/list/tweets",
		body: map[string]any{"tweets": []any{map[string]any{"id": "1"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"listId": "777", "queryType": "Latest"})
		},
	}))
	it := c.Lists.Tweets(context.Background(), "777", &ListTweetsOpts{QueryType: "Latest"})
	got, err := it.All(0)
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestListsService_MembersAndFollowers(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t,
		jsonRoute{
			method: "GET", path: "/twitter/list/members",
			body: map[string]any{"members": []any{map[string]any{"id": "m"}}, "has_next_page": false},
		},
		jsonRoute{
			method: "GET", path: "/twitter/list/followers",
			body: map[string]any{"followers": []any{map[string]any{"id": "f"}}, "has_next_page": false},
		},
	))
	mem, err := c.Lists.MembersPage(context.Background(), "1", "")
	if err != nil || len(mem.Members) != 1 {
		t.Fatalf("members: %+v err=%v", mem, err)
	}
	fol, err := c.Lists.FollowersPage(context.Background(), "1", "")
	if err != nil || len(fol.Followers) != 1 {
		t.Fatalf("followers: %+v err=%v", fol, err)
	}
}

func TestMonitorService_Add_RemoveAtSign(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/oapi/x_user_stream/add_user_to_monitor_tweet",
		body: map[string]any{"status": "success", "msg": "ok"},
		check: func(t *testing.T, r *http.Request) {
			decodeBody(t, r, &got)
		},
	}))
	if _, err := c.Monitor.AddUser(context.Background(), "@elon"); err != nil {
		t.Fatal(err)
	}
	if got["x_user_name"] != "elon" {
		t.Fatalf("@ not stripped: %+v", got)
	}
}

func TestMonitorService_List(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/oapi/x_user_stream/get_user_to_monitor_tweet",
		body: map[string]any{"data": []any{
			map[string]any{"id_for_user": "1", "x_user_name": "elon"},
		}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"query_type": "1"})
		},
	}))
	got, err := c.Monitor.List(context.Background(), MonitorTweets)
	if err != nil || len(got) != 1 || got[0].XUserName != "elon" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestWebhookService_Rules(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t,
		jsonRoute{
			method: "POST", path: "/oapi/tweet_filter/add_rule",
			body: map[string]any{"rule_id": "r1", "status": "success", "msg": "ok"},
			check: func(t *testing.T, r *http.Request) {
				var b map[string]any
				decodeBody(t, r, &b)
				if b["interval_seconds"].(float64) != 60 {
					t.Errorf("interval=%v", b["interval_seconds"])
				}
			},
		},
		jsonRoute{
			method: "POST", path: "/oapi/tweet_filter/update_rule",
			body: map[string]any{"status": "success", "msg": "ok"},
			check: func(t *testing.T, r *http.Request) {
				var b map[string]any
				decodeBody(t, r, &b)
				if b["is_effect"].(float64) != 1 {
					t.Errorf("is_effect=%v", b["is_effect"])
				}
			},
		},
		jsonRoute{
			method: "DELETE", path: "/oapi/tweet_filter/delete_rule",
			body: map[string]any{"status": "success", "msg": "ok"},
		},
		jsonRoute{
			method: "GET", path: "/oapi/tweet_filter/get_rules",
			body: map[string]any{"rules": []any{
				map[string]any{"rule_id": "r1", "tag": "t", "value": "from:x"},
			}},
		},
	))
	add, err := c.Webhook.AddRule(context.Background(), "tag", "from:x", 0)
	if err != nil || add.RuleID != "r1" {
		t.Fatalf("add=%+v err=%v", add, err)
	}
	if _, err := c.Webhook.UpdateRule(context.Background(), "r1", "tag", "from:x", 0, true); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Webhook.DeleteRule(context.Background(), "r1"); err != nil {
		t.Fatal(err)
	}
	rules, err := c.Webhook.ListRules(context.Background())
	if err != nil || len(rules) != 1 {
		t.Fatalf("rules=%+v err=%v", rules, err)
	}
}
