package twitterapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTweetsService_ByIDs(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweets",
		body: map[string]any{"tweets": []any{map[string]any{"id": "1"}, map[string]any{"id": "2"}}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"tweet_ids": "1,2"})
		},
	}))
	got, err := c.Tweets.ByIDs(context.Background(), []string{"1", "2"})
	if err != nil || len(got) != 2 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestTweetsService_UnmarshalQuoted(t *testing.T) {
	body := `{"tweets":[{"id":"1","quoted_tweet":{"id":"q","text":"quoted"},"retweeted_tweet":"<unknown>"}]}`
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweets", body: json.RawMessage(body),
	}))
	got, err := c.Tweets.ByIDs(context.Background(), []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].QuotedTweet == nil || got[0].QuotedTweet.ID != "q" {
		t.Fatalf("quoted: %+v", got[0].QuotedTweet)
	}
	if got[0].RetweetedTweet != nil {
		t.Fatalf("'<unknown>' should yield nil RetweetedTweet")
	}
	if string(got[0].RetweetedRaw) == "" {
		t.Fatal("RetweetedRaw should be preserved")
	}
}

func TestTweetsService_Replies(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweet/replies",
		body: map[string]any{"tweets": []any{map[string]any{"id": "r"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"tweetId": "1", "sinceTime": "100", "untilTime": "200"})
		},
	}))
	resp, err := c.Tweets.RepliesPage(context.Background(), "1", &RepliesOpts{SinceTimeUnix: 100, UntilTimeUnix: 200})
	if err != nil || len(resp.Tweets) != 1 {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
}

func TestTweetsService_AdvancedSearch(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweet/advanced_search",
		body: map[string]any{"tweets": []any{map[string]any{"id": "1"}}, "has_next_page": false},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"query": "from:openai", "queryType": "Latest"})
		},
	}))
	it := c.Tweets.AdvancedSearch(context.Background(), "from:openai", &AdvancedSearchOpts{QueryType: "Latest"})
	count := 0
	for it.Next() {
		count++
	}
	if it.Err() != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, it.Err())
	}
}

func TestTweetsService_BulkAdvancedSearch(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/tweet/bulk_advanced_search",
		body: map[string]any{"results": map[string]any{
			"query_0": map[string]any{"tweets": []any{map[string]any{"id": "a"}}},
			"query_1": map[string]any{"tweets": []any{map[string]any{"id": "b"}}},
		}},
		check: func(t *testing.T, r *http.Request) {
			var body struct {
				Queries []BulkSearchQuery `json:"queries"`
			}
			decodeBody(t, r, &body)
			if len(body.Queries) != 2 {
				t.Errorf("got %d queries", len(body.Queries))
			}
		},
	}))
	resp, err := c.Tweets.BulkAdvancedSearch(context.Background(), []BulkSearchQuery{
		{Query: "x"}, {Query: "y"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 || len(resp.Results["query_0"].Tweets) != 1 {
		t.Fatalf("got %+v", resp.Results)
	}
}

func TestTweetsService_Article(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/article",
		body: map[string]any{"article": map[string]any{
			"title": "hello", "likeCount": 10,
			"contents": []any{map[string]any{"type": "text", "text": "body"}},
		}},
		check: func(t *testing.T, r *http.Request) {
			wantQuery(t, r, map[string]string{"tweet_id": "9"})
		},
	}))
	a, err := c.Tweets.Article(context.Background(), "9")
	if err != nil || a.Title != "hello" || a.LikeCount != 10 || len(a.Contents) != 1 {
		t.Fatalf("a=%+v err=%v", a, err)
	}
}

func TestTweetsService_ThreadContext_Aliases(t *testing.T) {
	// Server sometimes returns "replies" instead of "tweets" — make sure we handle both.
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweet/thread_context",
		body: map[string]any{"replies": []any{map[string]any{"id": "x"}}, "has_next_page": false},
	}))
	resp, err := c.Tweets.ThreadContextPage(context.Background(), "1", "")
	if err != nil || len(resp.Tweets) != 1 || resp.Tweets[0].ID != "x" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
}

func TestTweetsService_ReplyChainToRoot(t *testing.T) {
	// Build: A (root) <- B <- C (start)
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Path {
		case "/twitter/tweet/thread_context":
			// Return only B in context to force a /tweets fallback for A.
			_, _ = w.Write([]byte(`{"tweets":[{"id":"B","isReply":true,"inReplyToId":"A"}],"has_next_page":false}`))
		case "/twitter/tweets":
			ids := r.URL.Query().Get("tweet_ids")
			if ids == "A" {
				_, _ = w.Write([]byte(`{"tweets":[{"id":"A","isReply":false}]}`))
				return
			}
			t.Fatalf("unexpected tweet_ids=%q", ids)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	start := Tweet{ID: "C", IsReply: true, InReplyToID: "B"}
	chain, err := c.Tweets.ReplyChainToRoot(context.Background(), start, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain) != 3 || chain[0].ID != "A" || chain[2].ID != "C" {
		t.Fatalf("chain=%+v", chain)
	}
	if calls < 2 {
		t.Fatalf("expected calls to context+tweets, got %d", calls)
	}
}

func TestTweetsService_RetweetersIterator(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/twitter/tweet/retweeters",
		body: map[string]any{"users": []any{map[string]any{"id": "u"}}, "has_next_page": false},
	}))
	it := c.Tweets.Retweeters(context.Background(), "1")
	count := 0
	for it.Next() {
		count++
	}
	if it.Err() != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, it.Err())
	}
}
