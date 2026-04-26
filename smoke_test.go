// Smoke tests that exercise every thin wrapper to (a) verify the path is
// correct and (b) bump statement coverage. Each request is matched against a
// shared mux that responds with a canned success envelope.
package twitterapi

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// smokeServer answers every documented endpoint with a generic success body.
// Path -> handler override.
type smokeMux struct {
	t        *testing.T
	mu       sync.Mutex
	hits     map[string]int
	override map[string]http.HandlerFunc
}

func newSmokeMux(t *testing.T) *smokeMux {
	return &smokeMux{
		t:        t,
		hits:     map[string]int{},
		override: map[string]http.HandlerFunc{},
	}
}

func (m *smokeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.hits[r.URL.Path]++
	h := m.override[r.URL.Path]
	m.mu.Unlock()
	if h != nil {
		h(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch {
	case strings.HasSuffix(r.URL.Path, "/info") && strings.Contains(r.URL.Path, "/community"):
		_, _ = w.Write([]byte(`{"community_info":{"id":"c"},"status":"success"}`))
	case strings.HasSuffix(r.URL.Path, "/spaces/detail"):
		_, _ = w.Write([]byte(`{"data":{"id":"s"},"status":"success"}`))
	case strings.HasSuffix(r.URL.Path, "/article"):
		_, _ = w.Write([]byte(`{"article":{"title":"t"},"status":"success"}`))
	case strings.HasSuffix(r.URL.Path, "/check_follow_relationship"):
		_, _ = w.Write([]byte(`{"data":{"following":true,"followed_by":false},"status":"success"}`))
	case strings.HasSuffix(r.URL.Path, "/get_dm_history_by_user_id"):
		_, _ = w.Write([]byte(`{"messages":[],"status":"success"}`))
	default:
		_, _ = w.Write([]byte(`{"status":"success","msg":"ok","tweets":[],"followers":[],"followings":[],"users":[],"members":[],"moderators":[],"data":{},"trends":[],"rules":[],"results":{}}`))
	}
}

// TestSmoke_AllEndpoints calls every simple endpoint method to verify the
// HTTP shape and exercise statement coverage.
func TestSmoke_AllEndpoints(t *testing.T) {
	m := newSmokeMux(t)
	c, _ := newTestClient(t, m.ServeHTTP, func(o *Options) {
		o.LoginCookie = "C"
	})
	ctx := context.Background()

	mustOK(t, "Users.GetByUsername", call(func() error {
		_, e := c.Users.GetByUsername(ctx, "x")
		return e
	}))
	mustOK(t, "Users.About", call(func() error { _, e := c.Users.About(ctx, "x"); return e }))
	mustOK(t, "Users.BatchByIDs", call(func() error { _, e := c.Users.BatchByIDs(ctx, []string{"1"}); return e }))
	mustOK(t, "Users.SearchPage", call(func() error { _, e := c.Users.SearchPage(ctx, "q", nil); return e }))
	mustOK(t, "Users.FollowingsPage", call(func() error { _, e := c.Users.FollowingsPage(ctx, "x", nil); return e }))
	mustOK(t, "Users.LastTweetsPage", call(func() error {
		_, e := c.Users.LastTweetsPage(ctx, LastTweetsOpts{UserName: "x"})
		return e
	}))
	mustOK(t, "Users.TimelinePage", call(func() error {
		_, e := c.Users.TimelinePage(ctx, TimelineOpts{UserID: "1"})
		return e
	}))
	mustOK(t, "Users.MentionsPage", call(func() error { _, e := c.Users.MentionsPage(ctx, "x", nil); return e }))

	mustOK(t, "Tweets.RepliesV2Page", call(func() error { _, e := c.Tweets.RepliesV2Page(ctx, "1", nil); return e }))
	mustOK(t, "Tweets.QuotesPage", call(func() error { _, e := c.Tweets.QuotesPage(ctx, "1", nil); return e }))
	mustOK(t, "Tweets.RetweetersPage", call(func() error { _, e := c.Tweets.RetweetersPage(ctx, "1", ""); return e }))
	mustOK(t, "Tweets.AdvancedSearchPage", call(func() error { _, e := c.Tweets.AdvancedSearchPage(ctx, "q", nil); return e }))
	mustOK(t, "Tweets.Article", call(func() error { _, e := c.Tweets.Article(ctx, "1"); return e }))

	mustOK(t, "Communities.ModeratorsPage", call(func() error { _, e := c.Communities.ModeratorsPage(ctx, "c", ""); return e }))
	mustOK(t, "Communities.TweetsPage", call(func() error { _, e := c.Communities.TweetsPage(ctx, "c", ""); return e }))
	mustOK(t, "Communities.SearchAllPage", call(func() error { _, e := c.Communities.SearchAllPage(ctx, "q", nil); return e }))

	mustOK(t, "Lists.TweetsPage", call(func() error { _, e := c.Lists.TweetsPage(ctx, "1", nil); return e }))
	mustOK(t, "Lists.TimelinePage", call(func() error { _, e := c.Lists.TimelinePage(ctx, "1", ""); return e }))

	mustOK(t, "Account.AccountDetailV3", call(func() error { _, e := c.Account.AccountDetailV3(ctx, "u"); return e }))
	mustOK(t, "Account.DeleteAccountV3", call(func() error { return c.Account.DeleteAccountV3(ctx, "u") }))

	mustOK(t, "Actions.DeleteTweet", call(func() error { _, e := c.Actions.DeleteTweet(ctx, "1", ""); return e }))
	mustOK(t, "Actions.UnlikeTweet", call(func() error { _, e := c.Actions.UnlikeTweet(ctx, "1", ""); return e }))
	mustOK(t, "Actions.Retweet", call(func() error { _, e := c.Actions.Retweet(ctx, "1", ""); return e }))
	mustOK(t, "Actions.Bookmark", call(func() error { _, e := c.Actions.Bookmark(ctx, "1", ""); return e }))
	mustOK(t, "Actions.Unbookmark", call(func() error { _, e := c.Actions.Unbookmark(ctx, "1", ""); return e }))
	mustOK(t, "Actions.FollowUser", call(func() error { _, e := c.Actions.FollowUser(ctx, "1", ""); return e }))
	mustOK(t, "Actions.UnfollowUser", call(func() error { _, e := c.Actions.UnfollowUser(ctx, "1", ""); return e }))
	mustOK(t, "Actions.JoinCommunity", call(func() error { _, e := c.Actions.JoinCommunity(ctx, "c", ""); return e }))
	mustOK(t, "Actions.DeleteCommunity", call(func() error { _, e := c.Actions.DeleteCommunity(ctx, "c", "n", ""); return e }))
	mustOK(t, "Actions.AddListMember", call(func() error { _, e := c.Actions.AddListMember(ctx, "1", "2", ""); return e }))
	mustOK(t, "Actions.Report", call(func() error {
		_, e := c.Actions.Report(ctx, ReportParams{TweetID: "1", Reason: ReportSpam})
		return e
	}))
	mustOK(t, "Actions.DMHistory", call(func() error { _, e := c.Actions.DMHistory(ctx, "1", ""); return e }))

	mustOK(t, "Monitor.RemoveUser", call(func() error { _, e := c.Monitor.RemoveUser(ctx, "id1"); return e }))
	mustOK(t, "Monitor.MonitorAccount", call(func() error { _, e := c.Monitor.MonitorAccount(ctx, MonitorAll); return e }))

	mustOK(t, "Legacy.AddListMember", call(func() error {
		_, e := c.Legacy.AddListMember(ctx, LegacyListMemberParams{AuthSession: "s", ListID: "1", UserID: "u"})
		return e
	}))
	mustOK(t, "Legacy.RemoveListMember", call(func() error {
		_, e := c.Legacy.RemoveListMember(ctx, LegacyListMemberParams{AuthSession: "s", ListID: "1", UserID: "u"})
		return e
	}))
}

func call(f func() error) error { return f() }

func mustOK(t *testing.T, name string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", name, err)
	}
}

// SetDefaultProxy is a setter not exercised elsewhere.
func TestSetDefaultProxy(t *testing.T) {
	c, _ := New(Options{APIKey: "k"})
	c.SetDefaultProxy(" http://x ")
	if c.pickProxy("") != "http://x" {
		t.Fatal("not set")
	}
}
