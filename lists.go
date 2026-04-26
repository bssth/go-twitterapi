package twitterapi

import (
	"context"
	"net/url"
	"strconv"
)

// ListsService groups /twitter/list/* read endpoints. Write endpoints (add /
// remove member) live on LegacyService — twitterapi.io exposes them only via
// auth_session, not v2 cookies.
type ListsService struct{ c *Client }

type ListMembersResponse struct {
	Members     []User `json:"members"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type ListFollowersResponse struct {
	Followers   []User `json:"followers"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

// ListTweetsOpts tunes /twitter/list/tweets.
type ListTweetsOpts struct {
	SinceTimeUnix  int64
	UntilTimeUnix  int64
	IncludeReplies *bool
	QueryType      string
	Cursor         string
}

func (s *ListsService) TweetsPage(ctx context.Context, listID string, opts *ListTweetsOpts) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("listId", listID)
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
		if opts.QueryType != "" {
			q.Set("queryType", opts.QueryType)
		}
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/list/tweets", q, &r)
	return r, err
}

func (s *ListsService) Tweets(ctx context.Context, listID string, opts *ListTweetsOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := ListTweetsOpts{Cursor: cursor}
		if base != nil {
			o.SinceTimeUnix = base.SinceTimeUnix
			o.UntilTimeUnix = base.UntilTimeUnix
			o.IncludeReplies = base.IncludeReplies
			o.QueryType = base.QueryType
		}
		r, err := s.TweetsPage(ctx, listID, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// TimelinePage hits /twitter/list/tweets_timeline (camelCase listId).
func (s *ListsService) TimelinePage(ctx context.Context, listID, cursor string) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("listId", listID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/list/tweets_timeline", q, &r)
	return r, err
}

func (s *ListsService) Timeline(ctx context.Context, listID string) *Iterator[Tweet] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		r, err := s.TimelinePage(ctx, listID, cursor)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

func (s *ListsService) MembersPage(ctx context.Context, listID, cursor string) (ListMembersResponse, error) {
	q := url.Values{}
	q.Set("list_id", listID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r ListMembersResponse
	err := s.c.getJSON(ctx, "/twitter/list/members", q, &r)
	return r, err
}

func (s *ListsService) Members(ctx context.Context, listID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.MembersPage(ctx, listID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Members, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

func (s *ListsService) FollowersPage(ctx context.Context, listID, cursor string) (ListFollowersResponse, error) {
	q := url.Values{}
	q.Set("list_id", listID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r ListFollowersResponse
	err := s.c.getJSON(ctx, "/twitter/list/followers", q, &r)
	return r, err
}

func (s *ListsService) Followers(ctx context.Context, listID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.FollowersPage(ctx, listID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Followers, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}
