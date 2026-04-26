package twitterapi

import (
	"context"
	"net/url"
)

// CommunitiesService groups /twitter/community/* read endpoints.
type CommunitiesService struct{ c *Client }

type CommunityInfoResponse struct {
	CommunityInfo Community `json:"community_info"`
	APIStatus
}

type CommunityMembersResponse struct {
	Members     []User `json:"members"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type CommunityModeratorsResponse struct {
	Moderators  []User `json:"moderators"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

// Info fetches /twitter/community/info.
func (s *CommunitiesService) Info(ctx context.Context, communityID string) (Community, error) {
	q := url.Values{}
	q.Set("community_id", communityID)
	var r CommunityInfoResponse
	if err := s.c.getJSON(ctx, "/twitter/community/info", q, &r); err != nil {
		return Community{}, err
	}
	return r.CommunityInfo, nil
}

func (s *CommunitiesService) MembersPage(ctx context.Context, communityID, cursor string) (CommunityMembersResponse, error) {
	q := url.Values{}
	q.Set("community_id", communityID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r CommunityMembersResponse
	err := s.c.getJSON(ctx, "/twitter/community/members", q, &r)
	return r, err
}

func (s *CommunitiesService) Members(ctx context.Context, communityID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.MembersPage(ctx, communityID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Members, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

func (s *CommunitiesService) ModeratorsPage(ctx context.Context, communityID, cursor string) (CommunityModeratorsResponse, error) {
	q := url.Values{}
	q.Set("community_id", communityID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r CommunityModeratorsResponse
	err := s.c.getJSON(ctx, "/twitter/community/moderators", q, &r)
	return r, err
}

func (s *CommunitiesService) Moderators(ctx context.Context, communityID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.ModeratorsPage(ctx, communityID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Moderators, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

func (s *CommunitiesService) TweetsPage(ctx context.Context, communityID, cursor string) (CursorTweetsResponse, error) {
	q := url.Values{}
	q.Set("community_id", communityID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r CursorTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/community/tweets", q, &r)
	return r, err
}

func (s *CommunitiesService) Tweets(ctx context.Context, communityID string) *Iterator[Tweet] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		r, err := s.TweetsPage(ctx, communityID, cursor)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// AllCommunitySearchOpts tunes /twitter/community/get_tweets_from_all_community.
type AllCommunitySearchOpts struct {
	QueryType string // "Latest" or "Top"
	Cursor    string
}

func (s *CommunitiesService) SearchAllPage(ctx context.Context, query string, opts *AllCommunitySearchOpts) (CursorTweetsResponse, error) {
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
	err := s.c.getJSON(ctx, "/twitter/community/get_tweets_from_all_community", q, &r)
	return r, err
}

func (s *CommunitiesService) SearchAll(ctx context.Context, query string, opts *AllCommunitySearchOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := AllCommunitySearchOpts{Cursor: cursor}
		if base != nil {
			o.QueryType = base.QueryType
		}
		r, err := s.SearchAllPage(ctx, query, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}
