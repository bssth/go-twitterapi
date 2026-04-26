package twitterapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
)

// AccountService covers account/auth endpoints: balance, login flows, cookie
// lifecycle. Read methods that need only the API key are also here.
type AccountService struct{ c *Client }

// Info returns the account credit balance.
func (s *AccountService) Info(ctx context.Context) (AccountInfo, error) {
	var r struct {
		AccountInfo
		APIStatus
	}
	if err := s.c.getJSON(ctx, "/oapi/my/info", nil, &r); err != nil {
		return AccountInfo{}, err
	}
	return r.AccountInfo, nil
}

// LoginV2Params are the inputs of /twitter/user_login_v2.
type LoginV2Params struct {
	UserName   string
	Email      string
	Password   string
	Proxy      string // residential preferred
	TOTPSecret string // optional, recommended
}

// LoginV2Response carries the cookie returned by user_login_v2.
type LoginV2Response struct {
	LoginCookie string `json:"login_cookies"`
	APIStatus
}

// LoginV2 authenticates and caches the resulting login_cookies via the
// configured TokenStore. Returns the cookie string.
func (s *AccountService) LoginV2(ctx context.Context, p LoginV2Params) (string, error) {
	if p.UserName == "" || p.Email == "" || p.Password == "" {
		return "", errors.New("twitterapi: LoginV2 requires UserName, Email and Password")
	}
	proxy := s.c.pickProxy(p.Proxy)
	if proxy == "" {
		return "", errors.New("twitterapi: LoginV2 requires Proxy (or DefaultProxy / TWITTERAPIIO_PROXY)")
	}
	payload := map[string]any{
		"user_name": p.UserName,
		"email":     p.Email,
		"password":  p.Password,
		"proxy":     proxy,
	}
	if p.TOTPSecret != "" {
		payload["totp_secret"] = p.TOTPSecret
	}

	// We want the raw body too — keep it on LoginState for debug.
	raw, err := s.c.rawPost(ctx, "/twitter/user_login_v2", payload)
	if err != nil {
		return "", err
	}
	var r LoginV2Response
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	if st, ok := extractStatus(raw); ok && st.IsError() {
		return "", &APIError{StatusCode: 200, Status: st.Status, Message: st.Message(), Code: st.Code, Body: raw}
	}
	cookie := strings.TrimSpace(r.LoginCookie)
	if cookie == "" {
		return "", errors.New("twitterapi: user_login_v2 returned empty login_cookies")
	}
	s.c.saveToken(&LoginState{
		LoginCookie:   cookie,
		SavedAtUnix:   time.Now().Unix(),
		RawLoginReply: raw,
	})
	return cookie, nil
}

// LoginV3Params — async login flow. Poll AccountDetailV3 until status=Active.
type LoginV3Params struct {
	UserName string
	Email    string
	Password string
	TOTPCode string
	Cookie   string
	Proxy    string
}

// LoginV3 starts the v3 async login. Returns the (eventual) login_cookies.
func (s *AccountService) LoginV3(ctx context.Context, p LoginV3Params) (string, error) {
	if p.UserName == "" {
		return "", errors.New("twitterapi: LoginV3 requires UserName")
	}
	proxy := s.c.pickProxy(p.Proxy)
	if proxy == "" {
		return "", errors.New("twitterapi: LoginV3 requires Proxy")
	}
	payload := map[string]any{
		"user_name": p.UserName,
		"proxy":     proxy,
	}
	if p.Email != "" {
		payload["email"] = p.Email
	}
	if p.Password != "" {
		payload["password"] = p.Password
	}
	if p.TOTPCode != "" {
		payload["totp_code"] = p.TOTPCode
	}
	if p.Cookie != "" {
		payload["cookie"] = p.Cookie
	}
	var r LoginV2Response
	if err := s.c.postJSON(ctx, "/twitter/user_login_v3", payload, &r); err != nil {
		return "", err
	}
	cookie := strings.TrimSpace(r.LoginCookie)
	if cookie != "" {
		s.c.saveToken(&LoginState{LoginCookie: cookie, SavedAtUnix: time.Now().Unix()})
	}
	return cookie, nil
}

// AccountDetailV3 polls /twitter/get_my_x_account_detail_v3.
func (s *AccountService) AccountDetailV3(ctx context.Context, userName string) (json.RawMessage, error) {
	q := url.Values{}
	q.Set("user_name", userName)
	return s.c.rawGet(ctx, "/twitter/get_my_x_account_detail_v3", q)
}

// DeleteAccountV3 removes the stored cookie for userName on the server.
func (s *AccountService) DeleteAccountV3(ctx context.Context, userName string) error {
	return s.c.deleteJSON(ctx, "/twitter/delete_my_x_account_v3", map[string]any{"user_name": userName}, nil)
}

// EnsureLogin resolves a cookie via, in order: cached -> LoginV2 with env
// (TWITTERAPIIO_USER_NAME, _EMAIL, _PASSWORD, _PROXY, optional _TOTP_SECRET).
// Most callers should not need to invoke this directly — every v2 write uses
// it internally.
func (s *AccountService) EnsureLogin(ctx context.Context) (string, error) {
	if cookie := s.c.LoginCookie(); cookie != "" {
		return cookie, nil
	}
	user := strings.TrimSpace(os.Getenv("TWITTERAPIIO_USER_NAME"))
	email := strings.TrimSpace(os.Getenv("TWITTERAPIIO_EMAIL"))
	pass := os.Getenv("TWITTERAPIIO_PASSWORD")
	totp := strings.TrimSpace(os.Getenv("TWITTERAPIIO_TOTP_SECRET"))
	if user == "" || email == "" || pass == "" {
		return "", ErrNeedLogin
	}
	return s.LoginV2(ctx, LoginV2Params{
		UserName:   user,
		Email:      email,
		Password:   pass,
		TOTPSecret: totp,
	})
}

func (c *Client) rawGet(ctx context.Context, path string, q url.Values) (json.RawMessage, error) {
	req, err := c.newRequest(ctx, "GET", path, q, nil, "")
	if err != nil {
		return nil, err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return body, err
	}
	if st, ok := extractStatus(body); ok && st.IsError() {
		return body, &APIError{StatusCode: 200, Status: st.Status, Message: st.Message(), Code: st.Code, Body: body}
	}
	return body, nil
}

func (c *Client) rawPost(ctx context.Context, path string, payload any) (json.RawMessage, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, "POST", path, nil, bytes.NewReader(b), "application/json")
	if err != nil {
		return nil, err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return body, err
	}
	return body, nil
}
