package twitterapi

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// TweetsService groups every read endpoint that returns tweets / users-of-tweets.
type TweetsService struct{ c *Client }

type TweetsResponse struct {
	Tweets []Tweet `json:"tweets"`
	APIStatus
}

type CursorTweetsResponse struct {
	Tweets      []Tweet `json:"tweets"`
	HasNextPage bool    `json:"has_next_page,omitempty"`
	NextCursor  string  `json:"next_cursor,omitempty"`
	APIStatus
}

type RetweetersResponse struct {
	Users       []User `json:"users"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type ArticleResponse struct {
	Article Article `json:"article"`
	APIStatus
}

type BulkSearchResponse struct {
	Results map[string]CursorTweetsResponse `json:"results"`
	APIStatus
}

// ByIDs fetches /twitter/tweets?tweet_ids=...
func (s *TweetsService) ByIDs(ctx context.Context, tweetIDs []string) ([]Tweet, error) {
	if len(tweetIDs) == 0 {
		return nil, errors.New("twitterapi: tweetIDs empty")
	}
	q := url.Values{}
	q.Set("tweet_ids", strings.Join(tweetIDs, ","))
	var r TweetsResponse
	if err := s.c.getJSON(ctx, "/twitter/tweets", q, &r); err != nil {
		return nil, err
	}
	return r.Tweets, nil
}

// RepliesOpts tunes /twitter/tweet/replies (v1).
type RepliesOpts struct {
	SinceTimeUnix int64
	UntilTimeUnix int64
	Cursor        string
}

// RepliesPage fetches one page of replies.
func (s *TweetsService) RepliesPage(ctx context.Context, tweetID string, opts *RepliesOpts) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("tweetId", tweetID)
	if opts != nil {
		if opts.SinceTimeUnix > 0 {
			q.Set("sinceTime", strconv.FormatInt(opts.SinceTimeUnix, 10))
		}
		if opts.UntilTimeUnix > 0 {
			q.Set("untilTime", strconv.FormatInt(opts.UntilTimeUnix, 10))
		}
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/tweet/replies", q, &r)
	return r, err
}

// Replies iterates every reply of a tweet.
func (s *TweetsService) Replies(ctx context.Context, tweetID string, opts *RepliesOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := RepliesOpts{Cursor: cursor}
		if base != nil {
			o.SinceTimeUnix = base.SinceTimeUnix
			o.UntilTimeUnix = base.UntilTimeUnix
		}
		r, err := s.RepliesPage(ctx, tweetID, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// RepliesV2Opts — sortable replies via /twitter/tweet/replies/v2.
type RepliesV2Opts struct {
	Cursor    string
	QueryType string // "Relevance" (default) | "Latest" | "Likes"
}

func (s *TweetsService) RepliesV2Page(ctx context.Context, tweetID string, opts *RepliesV2Opts) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("tweetId", tweetID)
	if opts != nil {
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
		if opts.QueryType != "" {
			q.Set("queryType", opts.QueryType)
		}
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/tweet/replies/v2", q, &r)
	return r, err
}

func (s *TweetsService) RepliesV2(ctx context.Context, tweetID string, opts *RepliesV2Opts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := RepliesV2Opts{Cursor: cursor}
		if base != nil {
			o.QueryType = base.QueryType
		}
		r, err := s.RepliesV2Page(ctx, tweetID, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// QuotesOpts tunes /twitter/tweet/quotes.
type QuotesOpts struct {
	SinceTimeUnix  int64
	UntilTimeUnix  int64
	IncludeReplies *bool
	Cursor         string
}

func (s *TweetsService) QuotesPage(ctx context.Context, tweetID string, opts *QuotesOpts) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("tweetId", tweetID)
	if opts != nil {
		if opts.SinceTimeUnix > 0 {
			q.Set("sinceTime", strconv.FormatInt(opts.SinceTimeUnix, 10))
		}
		if opts.UntilTimeUnix > 0 {
			q.Set("untilTime", strconv.FormatInt(opts.UntilTimeUnix, 10))
		}
		if opts.IncludeReplies != nil {
			q.Set("includeReplies", strconv.FormatBool(*opts.IncludeReplies))
		}
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/tweet/quotes", q, &r)
	return r, err
}

func (s *TweetsService) Quotes(ctx context.Context, tweetID string, opts *QuotesOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := QuotesOpts{Cursor: cursor}
		if base != nil {
			o.SinceTimeUnix = base.SinceTimeUnix
			o.UntilTimeUnix = base.UntilTimeUnix
			o.IncludeReplies = base.IncludeReplies
		}
		r, err := s.QuotesPage(ctx, tweetID, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// RetweetersPage — paginates /twitter/tweet/retweeters.
func (s *TweetsService) RetweetersPage(ctx context.Context, tweetID, cursor string) (RetweetersResponse, error) {
	q := url.Values{}
	q.Set("tweetId", tweetID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r RetweetersResponse
	err := s.c.getJSON(ctx, "/twitter/tweet/retweeters", q, &r)
	return r, err
}

// Retweeters iterates everyone who retweeted.
func (s *TweetsService) Retweeters(ctx context.Context, tweetID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.RetweetersPage(ctx, tweetID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Users, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// ThreadContextPage fetches one page of /twitter/tweet/thread_context.
func (s *TweetsService) ThreadContextPage(ctx context.Context, tweetID, cursor string) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("tweetId", tweetID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r CursorTweetsResponse
	// API ships replies under "replies" on this endpoint — alias decode.
	type aux struct {
		Replies     []Tweet `json:"replies"`
		Tweets      []Tweet `json:"tweets"`
		HasNextPage bool    `json:"has_next_page"`
		NextCursor  string  `json:"next_cursor"`
		APIStatus
	}
	var a aux
	err := s.c.getJSON(ctx, "/twitter/tweet/thread_context", q, &a)
	if err != nil {
		return r, err
	}
	r.Tweets = a.Tweets
	if len(r.Tweets) == 0 {
		r.Tweets = a.Replies
	}
	r.HasNextPage = a.HasNextPage
	r.NextCursor = a.NextCursor
	r.APIStatus = a.APIStatus
	return r, nil
}

func (s *TweetsService) ThreadContext(ctx context.Context, tweetID string) *Iterator[Tweet] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		r, err := s.ThreadContextPage(ctx, tweetID, cursor)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// AdvancedSearchOpts tunes /twitter/tweet/advanced_search.
type AdvancedSearchOpts struct {
	QueryType string // "Latest" or "Top"
	Cursor    string
}

func (s *TweetsService) AdvancedSearchPage(ctx context.Context, query string, opts *AdvancedSearchOpts) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("query", query)
	if opts != nil {
		if opts.QueryType != "" {
			q.Set("queryType", opts.QueryType)
		}
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/tweet/advanced_search", q, &r)
	return r, err
}

// AdvancedSearch iterates results of /twitter/tweet/advanced_search.
//
// Beware: docs warn against deep pagination; prefer narrow time windows on
// large queries.
func (s *TweetsService) AdvancedSearch(ctx context.Context, query string, opts *AdvancedSearchOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := AdvancedSearchOpts{Cursor: cursor}
		if base != nil {
			o.QueryType = base.QueryType
		}
		r, err := s.AdvancedSearchPage(ctx, query, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// BulkSearchQuery is one query in a bulk_advanced_search batch.
type BulkSearchQuery struct {
	Query     string `json:"query"`
	QueryType string `json:"queryType,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
}

// BulkAdvancedSearch runs many queries in parallel server-side. Results are
// keyed by "query_0", "query_1", ... in the order of input.
func (s *TweetsService) BulkAdvancedSearch(ctx context.Context, queries []BulkSearchQuery) (BulkSearchResponse, error) {
	var r BulkSearchResponse
	err := s.c.postJSON(ctx, "/twitter/tweet/bulk_advanced_search", map[string]any{"queries": queries}, &r)
	return r, err
}

// Article fetches /twitter/article.
func (s *TweetsService) Article(ctx context.Context, tweetID string) (Article, error) {
	q := url.Values{}
	q.Set("tweet_id", tweetID)
	var r ArticleResponse
	if err := s.c.getJSON(ctx, "/twitter/article", q, &r); err != nil {
		return Article{}, err
	}
	return r.Article, nil
}

// ReplyChainOpts tunes GetReplyChainToRoot.
type ReplyChainOpts struct {
	MaxContextPages int // default 3
}

// ReplyChainToRoot walks parent tweets up from start using thread_context, then
// /twitter/tweets fallbacks. Returned slice goes root -> ... -> start.
func (s *TweetsService) ReplyChainToRoot(ctx context.Context, start Tweet, opts *ReplyChainOpts) ([]Tweet, error) {
	if !start.IsReply {
		return []Tweet{start}, nil
	}
	if start.ID == "" {
		return nil, errors.New("twitterapi: start tweet has empty ID")
	}

	maxPages := 3
	if opts != nil && opts.MaxContextPages > 0 {
		maxPages = opts.MaxContextPages
	}

	all := make(map[string]Tweet, 256)
	cursor := ""
	for page := 0; page < maxPages; page++ {
		resp, err := s.ThreadContextPage(ctx, start.ID, cursor)
		if err != nil {
			break
		}
		for _, tw := range resp.Tweets {
			all[tw.ID] = tw
		}
		if !resp.HasNextPage || resp.NextCursor == "" {
			break
		}
		cursor = resp.NextCursor
	}
	if _, ok := all[start.ID]; !ok {
		all[start.ID] = start
	}

	chain := make([]Tweet, 0, 16)
	cur := start
	seen := map[string]bool{}
	for {
		if cur.ID == "" || seen[cur.ID] {
			break
		}
		seen[cur.ID] = true
		chain = append(chain, cur)

		if cur.InReplyToId == "" || (cur.ConversationId != "" && cur.ID == cur.ConversationId) {
			break
		}
		parentID := cur.InReplyToId
		if p, ok := all[parentID]; ok {
			cur = p
			continue
		}
		tweets, err := s.ByIDs(ctx, []string{parentID})
		if err != nil || len(tweets) == 0 {
			return reverseTweets(chain), nil
		}
		all[parentID] = tweets[0]
		cur = tweets[0]
	}
	return reverseTweets(chain), nil
}

func reverseTweets(in []Tweet) []Tweet {
	out := make([]Tweet, len(in))
	for i := range in {
		out[i] = in[len(in)-1-i]
	}
	return out
}
