package twitterapi

import (
	"context"
	"errors"
	"strings"
)

// ActionsService groups every v2 write endpoint. They share an auth contract:
// each request body must include login_cookies + proxy. The service handles
// both transparently — you only pass action-specific fields.
//
// On a "cookie expired/invalid" error from the API, the service refreshes the
// cookie via Account.EnsureLogin and retries once.
type ActionsService struct{ c *Client }

// ReportReason is the enum accepted by /twitter/report_v2.
type ReportReason string

const (
	ReportSpam              ReportReason = "SpamSimpleOption"
	ReportHateOrAbuse       ReportReason = "HateOrAbuseSimpleOption"
	ReportChildSafety       ReportReason = "ChildSafetySimpleOption"
	ReportViolentSpeech     ReportReason = "ViolentSpeechSimpleOption"
	ReportViolentMedia      ReportReason = "ViolentMediaSimpleOption"
	ReportIRB               ReportReason = "IRBSimpleOption"
	ReportImpersonation     ReportReason = "ImpersonationSimpleOption"
	ReportAdultContent      ReportReason = "AdultContentSimpleOption"
	ReportPrivateContent    ReportReason = "PrivateContentSimpleOption"
	ReportSuicideOrSelfHarm ReportReason = "SuicideSelfHarmSimpleOption"
	ReportTerrorism         ReportReason = "TerrorismSimpleOption"
	ReportCivicIntegrity    ReportReason = "CivicIntegritySimpleOption"
)

// SimpleStatusResponse is the canonical {status, msg} envelope of
// most v2 actions.
type SimpleStatusResponse struct {
	APIStatus
}

// CreateTweetParams configures CreateTweet. Either ReplyToTweetID or
// QuoteTweetID may be set (or neither); not both.
type CreateTweetParams struct {
	TweetText string

	ReplyToTweetID string
	QuoteTweetID   string
	AttachmentURL  string // alternate way to quote (full URL)
	CommunityID    string

	// IsNoteTweet enables long-form posting (>280 chars). Requires Premium.
	IsNoteTweet *bool

	MediaIDs []string

	// ScheduleFor uses ISO-8601 in UTC, e.g. "2026-01-20T10:00:00.000Z".
	ScheduleFor string

	// SkipSanitize disables the SanitizeForTwitter pre-processing pass.
	SkipSanitize bool

	// Proxy overrides the per-call proxy. Falls back to client default.
	Proxy string
}

// CreateTweetResponse is /twitter/create_tweet_v2's payload.
type CreateTweetResponse struct {
	TweetID string `json:"tweet_id"`
	APIStatus
}

// CreateTweet posts a tweet, reply, quote, or scheduled tweet.
func (s *ActionsService) CreateTweet(ctx context.Context, p CreateTweetParams) (CreateTweetResponse, error) {
	text := strings.TrimSpace(p.TweetText)
	if text == "" {
		return CreateTweetResponse{}, errors.New("twitterapi: TweetText is required")
	}
	if !p.SkipSanitize {
		text = SanitizeForTwitter(text)
	}
	payload := map[string]any{"tweet_text": text}
	if p.ReplyToTweetID != "" {
		payload["reply_to_tweet_id"] = p.ReplyToTweetID
	}
	if p.QuoteTweetID != "" {
		payload["quote_tweet_id"] = p.QuoteTweetID
	}
	if p.AttachmentURL != "" {
		payload["attachment_url"] = p.AttachmentURL
	}
	if p.CommunityID != "" {
		payload["community_id"] = p.CommunityID
	}
	if p.IsNoteTweet != nil {
		payload["is_note_tweet"] = *p.IsNoteTweet
	}
	if len(p.MediaIDs) > 0 {
		payload["media_ids"] = p.MediaIDs
	}
	if p.ScheduleFor != "" {
		payload["schedule_for"] = p.ScheduleFor
	}
	var r CreateTweetResponse
	if err := s.postAuthed(ctx, "/twitter/create_tweet_v2", p.Proxy, payload, &r); err != nil {
		return r, err
	}
	return r, nil
}

// DeleteTweet removes one of the authenticated user's tweets.
func (s *ActionsService) DeleteTweet(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/delete_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// LikeTweet likes a tweet.
func (s *ActionsService) LikeTweet(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/like_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// UnlikeTweet removes a like.
func (s *ActionsService) UnlikeTweet(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/unlike_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// Retweet retweets a tweet.
func (s *ActionsService) Retweet(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/retweet_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// Bookmark adds a tweet to the user's bookmarks.
func (s *ActionsService) Bookmark(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/bookmark_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// Unbookmark removes a bookmark.
func (s *ActionsService) Unbookmark(ctx context.Context, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/unbookmark_tweet_v2", proxy, map[string]any{"tweet_id": tweetID}, &r)
	return r, err
}

// BookmarksOpts tunes /twitter/bookmarks_v2.
type BookmarksOpts struct {
	Count  int // server default 20
	Cursor string
	Proxy  string
}

// BookmarksPage returns one page of the authed user's bookmarks.
func (s *ActionsService) BookmarksPage(ctx context.Context, opts BookmarksOpts) (CursorTweetsResponse, error) {
	payload := map[string]any{}
	if opts.Count > 0 {
		payload["count"] = opts.Count
	}
	if opts.Cursor != "" {
		payload["cursor"] = opts.Cursor
	}
	var r CursorTweetsResponse
	err := s.postAuthed(ctx, "/twitter/bookmarks_v2", opts.Proxy, payload, &r)
	return r, err
}

// Bookmarks iterates the authed user's bookmarks.
func (s *ActionsService) Bookmarks(ctx context.Context, opts BookmarksOpts) *Iterator[Tweet] {
	base := opts
	return NewIterator(ctx, func(ctx context.Context, cursor string) (Page[Tweet], error) {
		o := base
		o.Cursor = cursor
		r, err := s.BookmarksPage(ctx, o)
		if err != nil {
			return Page[Tweet]{}, err
		}
		return Page[Tweet]{Items: r.Tweets, NextCursor: r.NextCursor, HasNextPage: r.HasNextPage}, nil
	})
}

// FollowUser follows a user by numeric ID.
func (s *ActionsService) FollowUser(ctx context.Context, userID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/follow_user_v2", proxy, map[string]any{"user_id": userID}, &r)
	return r, err
}

// UnfollowUser unfollows a user by numeric ID.
func (s *ActionsService) UnfollowUser(ctx context.Context, userID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/unfollow_user_v2", proxy, map[string]any{"user_id": userID}, &r)
	return r, err
}

// SendDMParams configures SendDM.
type SendDMParams struct {
	UserID           string // recipient
	Text             string
	MediaID          string // singular — DM endpoint uses media_id, not media_ids
	ReplyToMessageID string
	Proxy            string
}

// SendDMResponse is /twitter/send_dm_to_user's payload.
type SendDMResponse struct {
	MessageID string `json:"message_id"`
	APIStatus
}

// SendDM sends a direct message.
func (s *ActionsService) SendDM(ctx context.Context, p SendDMParams) (SendDMResponse, error) {
	if p.UserID == "" || p.Text == "" {
		return SendDMResponse{}, errors.New("twitterapi: SendDM requires UserID and Text")
	}
	payload := map[string]any{
		"user_id": p.UserID,
		"text":    p.Text,
	}
	if p.MediaID != "" {
		payload["media_id"] = p.MediaID
	}
	if p.ReplyToMessageID != "" {
		payload["reply_to_message_id"] = p.ReplyToMessageID
	}
	var r SendDMResponse
	err := s.postAuthed(ctx, "/twitter/send_dm_to_user", p.Proxy, payload, &r)
	return r, err
}

// DMHistory fetches the DM thread with a user. The server takes the auth
// material (login_cookies, proxy) on the querystring for this endpoint.
func (s *ActionsService) DMHistory(ctx context.Context, userID, proxy string) ([]byte, error) {
	cookie, err := s.c.Account.EnsureLogin(ctx)
	if err != nil {
		return nil, err
	}
	q := map[string]any{
		"login_cookies": cookie,
		"user_id":       userID,
		"proxy":         s.c.pickProxy(proxy),
	}
	// Use POST-as-JSON to keep the implementation simple; the server accepts both.
	return s.c.rawPost(ctx, "/twitter/get_dm_history_by_user_id", q)
}

// ReportParams targets a tweet OR a user (exactly one of TweetID / UserID).
type ReportParams struct {
	TweetID string
	UserID  string
	Reason  ReportReason
	Proxy   string
}

// Report files an abuse report.
func (s *ActionsService) Report(ctx context.Context, p ReportParams) (SimpleStatusResponse, error) {
	if (p.TweetID == "" && p.UserID == "") || (p.TweetID != "" && p.UserID != "") {
		return SimpleStatusResponse{}, errors.New("twitterapi: Report requires exactly one of TweetID or UserID")
	}
	if p.Reason == "" {
		return SimpleStatusResponse{}, errors.New("twitterapi: Reason is required")
	}
	payload := map[string]any{"reason": string(p.Reason)}
	if p.TweetID != "" {
		payload["tweet_id"] = p.TweetID
	}
	if p.UserID != "" {
		payload["user_id"] = p.UserID
	}
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/report_v2", p.Proxy, payload, &r)
	return r, err
}

// CreateCommunityResponse is the payload of /twitter/create_community_v2.
type CreateCommunityResponse struct {
	CommunityID string `json:"community_id"`
	APIStatus
}

// CreateCommunity creates a Twitter Community.
func (s *ActionsService) CreateCommunity(ctx context.Context, name, description, proxy string) (CreateCommunityResponse, error) {
	if name == "" || description == "" {
		return CreateCommunityResponse{}, errors.New("twitterapi: name and description are required")
	}
	payload := map[string]any{
		"name":        name,
		"description": description,
	}
	var r CreateCommunityResponse
	err := s.postAuthed(ctx, "/twitter/create_community_v2", proxy, payload, &r)
	return r, err
}

// JoinCommunityResponse is the payload of join/leave community.
type JoinCommunityResponse struct {
	CommunityID   string `json:"community_id"`
	CommunityName string `json:"community_name"`
	APIStatus
}

func (s *ActionsService) JoinCommunity(ctx context.Context, communityID, proxy string) (JoinCommunityResponse, error) {
	var r JoinCommunityResponse
	err := s.postAuthed(ctx, "/twitter/join_community_v2", proxy, map[string]any{"community_id": communityID}, &r)
	return r, err
}

func (s *ActionsService) LeaveCommunity(ctx context.Context, communityID, proxy string) (JoinCommunityResponse, error) {
	var r JoinCommunityResponse
	err := s.postAuthed(ctx, "/twitter/leave_community_v2", proxy, map[string]any{"community_id": communityID}, &r)
	return r, err
}

func (s *ActionsService) DeleteCommunity(ctx context.Context, communityID, communityName, proxy string) (SimpleStatusResponse, error) {
	if communityID == "" || communityName == "" {
		return SimpleStatusResponse{}, errors.New("twitterapi: communityID and communityName are both required")
	}
	payload := map[string]any{
		"community_id":   communityID,
		"community_name": communityName,
	}
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/delete_community_v2", proxy, payload, &r)
	return r, err
}

// AddListMember (v2 endpoint exists for cookie-based auth). LegacyService.AddListMember
// is the auth_session variant.
func (s *ActionsService) AddListMember(ctx context.Context, listID, userID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.postAuthed(ctx, "/twitter/list/add_member_v2", proxy, map[string]any{
		"list_id": listID,
		"user_id": userID,
	}, &r)
	return r, err
}

func (s *ActionsService) postAuthed(ctx context.Context, path, proxyOverride string, payload map[string]any, out any) error {
	cookie, err := s.c.Account.EnsureLogin(ctx)
	if err != nil {
		return err
	}
	proxy := s.c.pickProxy(proxyOverride)
	if proxy == "" {
		return errors.New("twitterapi: v2 write requires a proxy (set Options.DefaultProxy or pass per-call)")
	}
	payload["login_cookies"] = cookie
	payload["proxy"] = proxy

	err = s.c.postJSON(ctx, path, payload, out)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrCookieExpired) {
		return err
	}

	// Force a fresh login and retry once.
	s.c.clearToken()
	cookie2, err2 := s.c.Account.EnsureLogin(ctx)
	if err2 != nil {
		return err
	}
	payload["login_cookies"] = cookie2
	return s.c.postJSON(ctx, path, payload, out)
}
