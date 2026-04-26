package twitterapi

import (
	"context"
	"encoding/json"
)

// LegacyService wraps the deprecated v1 write endpoints that authenticate via
// auth_session instead of login_cookies. New code should prefer
// ActionsService — these are kept only for callers with existing sessions.
//
// Deprecated: use ActionsService whenever possible.
type LegacyService struct{ c *Client }

// LegacyTweetParams configures /twitter/create_tweet.
//
// Deprecated.
type LegacyTweetParams struct {
	AuthSession      string
	TweetText        string
	QuoteTweetID     string
	InReplyToTweetID string
	MediaID          string
	Proxy            string
}

// LegacyTweetResponse is the (deeply nested) payload of /twitter/create_tweet.
type LegacyTweetResponse struct {
	Data json.RawMessage `json:"data"`
	APIStatus
}

// CreateTweet calls the deprecated v1 endpoint.
//
// Deprecated: use ActionsService.CreateTweet.
func (s *LegacyService) CreateTweet(ctx context.Context, p LegacyTweetParams) (LegacyTweetResponse, error) {
	payload := map[string]any{
		"auth_session": p.AuthSession,
		"tweet_text":   p.TweetText,
		"proxy":        s.c.pickProxy(p.Proxy),
	}
	if p.QuoteTweetID != "" {
		payload["quote_tweet_id"] = p.QuoteTweetID
	}
	if p.InReplyToTweetID != "" {
		payload["in_reply_to_tweet_id"] = p.InReplyToTweetID
	}
	if p.MediaID != "" {
		payload["media_id"] = p.MediaID
	}
	var r LegacyTweetResponse
	err := s.c.postJSON(ctx, "/twitter/create_tweet", payload, &r)
	return r, err
}

// LikeTweet calls /twitter/like_tweet.
//
// Deprecated: use ActionsService.LikeTweet.
func (s *LegacyService) LikeTweet(ctx context.Context, authSession, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/twitter/like_tweet", map[string]any{
		"auth_session": authSession,
		"tweet_id":     tweetID,
		"proxy":        s.c.pickProxy(proxy),
	}, &r)
	return r, err
}

// RetweetTweet calls /twitter/retweet_tweet.
//
// Deprecated: use ActionsService.Retweet.
func (s *LegacyService) RetweetTweet(ctx context.Context, authSession, tweetID, proxy string) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/twitter/retweet_tweet", map[string]any{
		"auth_session": authSession,
		"tweet_id":     tweetID,
		"proxy":        s.c.pickProxy(proxy),
	}, &r)
	return r, err
}

// UploadImageResponse is the payload of /twitter/upload_image.
//
// Deprecated.
type UploadImageResponse struct {
	MediaID string `json:"media_id"`
	APIStatus
}

// UploadImage calls /twitter/upload_image (server fetches the URL itself).
//
// Deprecated: use MediaService.Upload.
func (s *LegacyService) UploadImage(ctx context.Context, authSession, imageURL, proxy string) (UploadImageResponse, error) {
	var r UploadImageResponse
	err := s.c.postJSON(ctx, "/twitter/upload_image", map[string]any{
		"auth_session": authSession,
		"image_url":    imageURL,
		"proxy":        s.c.pickProxy(proxy),
	}, &r)
	return r, err
}

// LegacyListMemberParams configures /twitter/list/add_member and remove_member.
type LegacyListMemberParams struct {
	AuthSession string
	ListID      string
	UserID      string // either UserID or UserName
	UserName    string
	Proxy       string
}

func (s *LegacyService) buildListMember(p LegacyListMemberParams) map[string]any {
	payload := map[string]any{
		"auth_session": p.AuthSession,
		"list_id":      p.ListID,
		"proxy":        s.c.pickProxy(p.Proxy),
	}
	if p.UserID != "" {
		payload["user_id"] = p.UserID
	}
	if p.UserName != "" {
		payload["user_name"] = p.UserName
	}
	return payload
}

// AddListMember adds a member to a list (legacy).
//
// Deprecated: use ActionsService.AddListMember.
func (s *LegacyService) AddListMember(ctx context.Context, p LegacyListMemberParams) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/twitter/list/add_member", s.buildListMember(p), &r)
	return r, err
}

// RemoveListMember removes a member from a list (legacy).
//
// Deprecated.
func (s *LegacyService) RemoveListMember(ctx context.Context, p LegacyListMemberParams) (SimpleStatusResponse, error) {
	var r SimpleStatusResponse
	err := s.c.postJSON(ctx, "/twitter/list/remove_member", s.buildListMember(p), &r)
	return r, err
}

// LoginV1Response is the payload of the legacy login.
type LoginV1Response struct {
	Hint      string          `json:"hint,omitempty"`
	LoginData json.RawMessage `json:"login_data,omitempty"`
	APIStatus
}

// LoginV1 calls /twitter/login_by_email_or_username (legacy).
//
// Deprecated: use AccountService.LoginV2.
func (s *LegacyService) LoginV1(ctx context.Context, usernameOrEmail, password, proxy string) (LoginV1Response, error) {
	var r LoginV1Response
	err := s.c.postJSON(ctx, "/twitter/login_by_email_or_username", map[string]any{
		"username_or_email": usernameOrEmail,
		"password":          password,
		"proxy":             s.c.pickProxy(proxy),
	}, &r)
	return r, err
}

// Login2FAResponse is the payload of /twitter/login_by_2fa.
type Login2FAResponse struct {
	Session string `json:"session"`
	User    struct {
		IDStr      string `json:"id_str"`
		ScreenName string `json:"screen_name"`
		Name       string `json:"name"`
	} `json:"user"`
	APIStatus
}

// Login2FA completes the legacy 2FA flow.
//
// Deprecated.
func (s *LegacyService) Login2FA(ctx context.Context, loginData json.RawMessage, twoFactorCode, proxy string) (Login2FAResponse, error) {
	var r Login2FAResponse
	err := s.c.postJSON(ctx, "/twitter/login_by_2fa", map[string]any{
		"login_data": loginData,
		"2fa_code":   twoFactorCode,
		"proxy":      s.c.pickProxy(proxy),
	}, &r)
	return r, err
}
