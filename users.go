package twitterapi

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// UsersService groups every read endpoint under /twitter/user/*.
type UsersService struct{ c *Client }

type GetUserResponse struct {
	Data User `json:"data"`
	APIStatus
}

type BatchGetUsersResponse struct {
	Users []User `json:"users"`
	APIStatus
}

type SearchUsersResponse struct {
	Users       []User `json:"users"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type FollowersResponse struct {
	Followers   []User `json:"followers"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type FollowingsResponse struct {
	Followings  []User `json:"followings"`
	HasNextPage bool   `json:"has_next_page,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	APIStatus
}

type UserTweetsResponse struct {
	Tweets      []Tweet `json:"tweets"`
	HasNextPage bool    `json:"has_next_page,omitempty"`
	NextCursor  string  `json:"next_cursor,omitempty"`
	APIStatus
}

type UserMentionsResponse = UserTweetsResponse

type CheckFollowResponse struct {
	Data FollowRelationship `json:"data"`
	APIStatus
}

// GetByUsername fetches /twitter/user/info.
func (s *UsersService) GetByUsername(ctx context.Context, userName string) (User, error) {
	if strings.TrimSpace(userName) == "" {
		return User{}, errors.New("twitterapi: userName is required")
	}
	q := url.Values{}
	q.Set("userName", userName)
	var r GetUserResponse
	if err := s.c.getJSON(ctx, "/twitter/user/info", q, &r); err != nil {
		return User{}, err
	}
	return r.Data, nil
}

// About fetches the extended bio at /twitter/user_about.
func (s *UsersService) About(ctx context.Context, userName string) (User, error) {
	q := url.Values{}
	q.Set("userName", userName)
	var r GetUserResponse
	if err := s.c.getJSON(ctx, "/twitter/user_about", q, &r); err != nil {
		return User{}, err
	}
	return r.Data, nil
}

// BatchByIDs fetches up to N users by numeric ID via batch_info_by_ids.
func (s *UsersService) BatchByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("twitterapi: userIDs empty")
	}
	q := url.Values{}
	q.Set("userIds", strings.Join(userIDs, ","))
	var r BatchGetUsersResponse
	if err := s.c.getJSON(ctx, "/twitter/user/batch_info_by_ids", q, &r); err != nil {
		return nil, err
	}
	return r.Users, nil
}

// SearchOpts tunes user search.
type SearchOpts struct {
	Cursor string
}

// SearchPage fetches one page of /twitter/user/search.
func (s *UsersService) SearchPage(ctx context.Context, query string, opts *SearchOpts) (SearchUsersResponse, error) {
	q := url.Values{}
	q.Set("query", query)
	if opts != nil && opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	var r SearchUsersResponse
	err := s.c.getJSON(ctx, "/twitter/user/search", q, &r)
	return r, err
}

// Search returns an iterator over /twitter/user/search.
func (s *UsersService) Search(ctx context.Context, query string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.SearchPage(ctx, query, &SearchOpts{Cursor: cursor})
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Users, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// FollowersOpts tunes Followers / Followings.
type FollowersOpts struct {
	Cursor   string
	PageSize int // 20–200, default 200
}

// FollowersPage fetches one page of followers.
func (s *UsersService) FollowersPage(ctx context.Context, userName string, opts *FollowersOpts) (FollowersResponse, error) {
	q := url.Values{}
	q.Set("userName", userName)
	if opts != nil {
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
		if opts.PageSize > 0 {
			q.Set("pageSize", strconv.Itoa(opts.PageSize))
		}
	}
	var r FollowersResponse
	err := s.c.getJSON(ctx, "/twitter/user/followers", q, &r)
	return r, err
}

// Followers iterates every follower.
func (s *UsersService) Followers(ctx context.Context, userName string, opts *FollowersOpts) *Iterator[User] {
	page := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		o := FollowersOpts{Cursor: cursor}
		if page != nil {
			o.PageSize = page.PageSize
		}
		r, err := s.FollowersPage(ctx, userName, &o)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Followers, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// FollowingsPage fetches one page of followings.
func (s *UsersService) FollowingsPage(ctx context.Context, userName string, opts *FollowersOpts) (FollowingsResponse, error) {
	q := url.Values{}
	q.Set("userName", userName)
	if opts != nil {
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
		if opts.PageSize > 0 {
			q.Set("pageSize", strconv.Itoa(opts.PageSize))
		}
	}
	var r FollowingsResponse
	err := s.c.getJSON(ctx, "/twitter/user/followings", q, &r)
	return r, err
}

// Followings iterates everyone the user follows.
func (s *UsersService) Followings(ctx context.Context, userName string, opts *FollowersOpts) *Iterator[User] {
	page := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		o := FollowersOpts{Cursor: cursor}
		if page != nil {
			o.PageSize = page.PageSize
		}
		r, err := s.FollowingsPage(ctx, userName, &o)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Followings, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// VerifiedFollowers — paginates /twitter/user/verifiedFollowers (uses user_id).
func (s *UsersService) VerifiedFollowersPage(ctx context.Context, userID, cursor string) (FollowersResponse, error) {
	q := url.Values{}
	q.Set("user_id", userID)
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var r FollowersResponse
	err := s.c.getJSON(ctx, "/twitter/user/verifiedFollowers", q, &r)
	return r, err
}

func (s *UsersService) VerifiedFollowers(ctx context.Context, userID string) *Iterator[User] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[User], error) {
		r, err := s.VerifiedFollowersPage(ctx, userID, cursor)
		if err != nil {
			return Page[User]{}, err
		}
		return Page[User]{Items: r.Followers, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// MentionsOpts tunes /twitter/user/mentions.
type MentionsOpts struct {
	SinceTimeUnix int64
	UntilTimeUnix int64
	Cursor        string
}

// MentionsPage fetches one page of @-mentions for userName.
func (s *UsersService) MentionsPage(ctx context.Context, userName string, opts *MentionsOpts) (UserMentionsResponse, error) {
	q := url.Values{}
	q.Set("userName", userName)
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
	var r UserMentionsResponse
	err := s.c.getJSON(ctx, "/twitter/user/mentions", q, &r)
	return r, err
}

// Mentions iterates every mention of userName.
func (s *UsersService) Mentions(ctx context.Context, userName string, opts *MentionsOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := MentionsOpts{Cursor: cursor}
		if base != nil {
			o.SinceTimeUnix = base.SinceTimeUnix
			o.UntilTimeUnix = base.UntilTimeUnix
		}
		r, err := s.MentionsPage(ctx, userName, &o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// LastTweetsOpts tunes /twitter/user/last_tweets.
type LastTweetsOpts struct {
	UserID         string
	UserName       string
	Cursor         string
	IncludeReplies *bool
}

// LastTweetsPage fetches one page of /twitter/user/last_tweets.
func (s *UsersService) LastTweetsPage(ctx context.Context, opts LastTweetsOpts) (UserTweetsResponse, error) {
	if strings.TrimSpace(opts.UserID) == "" && strings.TrimSpace(opts.UserName) == "" {
		return UserTweetsResponse{}, errors.New("twitterapi: UserID or UserName is required")
	}
	q := url.Values{}
	if opts.UserID != "" {
		q.Set("userId", opts.UserID)
	}
	if opts.UserName != "" {
		q.Set("userName", opts.UserName)
	}
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	if opts.IncludeReplies != nil {
		q.Set("includeReplies", strconv.FormatBool(*opts.IncludeReplies))
	}
	var r UserTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/user/last_tweets", q, &r)
	return r, err
}

// LastTweets iterates a user's recent tweets.
func (s *UsersService) LastTweets(ctx context.Context, opts LastTweetsOpts) *Iterator[Tweet] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := opts
		o.Cursor = cursor
		r, err := s.LastTweetsPage(ctx, o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// TimelineOpts tunes /twitter/user/tweet_timeline.
type TimelineOpts struct {
	UserID             string
	Cursor             string
	IncludeReplies     *bool
	IncludeParentTweet *bool
}

// TimelinePage — full timeline (id-keyed).
func (s *UsersService) TimelinePage(ctx context.Context, opts TimelineOpts) (UserTweetsResponse, error) {
	if strings.TrimSpace(opts.UserID) == "" {
		return UserTweetsResponse{}, errors.New("twitterapi: UserID is required")
	}
	q := url.Values{}
	q.Set("userId", opts.UserID)
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	if opts.IncludeReplies != nil {
		q.Set("includeReplies", strconv.FormatBool(*opts.IncludeReplies))
	}
	if opts.IncludeParentTweet != nil {
		q.Set("includeParentTweet", strconv.FormatBool(*opts.IncludeParentTweet))
	}
	var r UserTweetsResponse
	err := s.c.getJSON(ctx, "/twitter/user/tweet_timeline", q, &r)
	return r, err
}

// Timeline iterates a user's full timeline.
func (s *UsersService) Timeline(ctx context.Context, opts TimelineOpts) *Iterator[Tweet] {
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := opts
		o.Cursor = cursor
		r, err := s.TimelinePage(ctx, o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// CheckFollow returns whether source follows target and vice versa.
func (s *UsersService) CheckFollow(ctx context.Context, sourceUserName, targetUserName string) (FollowRelationship, error) {
	q := url.Values{}
	q.Set("source_user_name", sourceUserName)
	q.Set("target_user_name", targetUserName)
	var r CheckFollowResponse
	if err := s.c.getJSON(ctx, "/twitter/user/check_follow_relationship", q, &r); err != nil {
		return FollowRelationship{}, err
	}
	return r.Data, nil
}
