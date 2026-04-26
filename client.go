// Package twitterapi is a Go SDK for the twitterapi.io API.
//
// See https://docs.twitterapi.io/introduction and the README for usage.
package twitterapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Default endpoints. Override via Options.BaseURL / Options.WSURL.
const (
	DefaultBaseURL = "https://api.twitterapi.io"
	DefaultWSURL   = "wss://ws.twitterapi.io/twitter/tweet/websocket"
	defaultUA      = "go-twitterapi/0.1 (+https://github.com/bssth/go-twitterapi)"
)

// Client is the entry point to the SDK. Construct via New.
//
// All resource services are exposed as fields:
//
//	c.Users.GetByUsername(ctx, "elonmusk")
//	c.Tweets.AdvancedSearch(ctx, "from:elonmusk", nil)
//	c.Actions.CreateTweet(ctx, twitterapi.CreateTweetParams{TweetText: "hi"})
//
// A Client is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey  string
	baseURL string
	wsURL   string

	httpClient *http.Client
	userAgent  string

	freePlan bool
	rlMu     sync.Mutex
	lastReq  time.Time

	maxRetries int
	minBackoff time.Duration
	maxBackoff time.Duration

	tokenStore TokenStore
	tokenMu    sync.Mutex
	token      *LoginState

	defaultProxy string

	// Resource services. Populated by New.
	Users       *UsersService
	Tweets      *TweetsService
	Communities *CommunitiesService
	Spaces      *SpacesService
	Trends      *TrendsService
	Lists       *ListsService
	Account     *AccountService
	Actions     *ActionsService
	Media       *MediaService
	Monitor     *MonitorService
	Webhook     *WebhookService
	Legacy      *LegacyService
}

// Options configures a Client. APIKey is required (or set TWITTERAPIIO_API_KEY).
type Options struct {
	// APIKey is your twitterapi.io key (sent as x-api-key). If empty, falls
	// back to env TWITTERAPIIO_API_KEY.
	APIKey string

	// BaseURL overrides the API root. Defaults to DefaultBaseURL.
	BaseURL string

	// WSURL overrides the WebSocket endpoint for the experimental tweet stream.
	// Defaults to DefaultWSURL.
	WSURL string

	// HTTPClient lets you swap the underlying *http.Client (for proxies,
	// timeouts, transport tracing). A sensible default is used otherwise.
	HTTPClient *http.Client

	// UserAgent sent on every request. Defaults to a library identifier.
	UserAgent string

	// FreePlan inserts a 5s gap between requests to fit the free-tier rate limit.
	FreePlan bool

	// Retry tuning. Zero values use sane defaults.
	MaxRetries int
	MinBackoff time.Duration
	MaxBackoff time.Duration

	// TokenStore persists login_cookies for v2 write actions. If nil and
	// TokenFile != "", a FileTokenStore is created. Both empty disables
	// persistence (you must call Account.LoginV2 yourself or pass
	// LoginCookie).
	TokenStore TokenStore
	TokenFile  string

	// LoginCookie pre-populates the login_cookies used by v2 writes,
	// bypassing TokenStore on first call. Useful when you already have a
	// cookie from another process.
	LoginCookie string

	// DefaultProxy is sent in v2 write payloads when the per-call Proxy is
	// empty. Falls back to env TWITTERAPIIO_PROXY.
	DefaultProxy string
}

// ErrMissingAPIKey is returned by New if no API key was supplied.
var ErrMissingAPIKey = errors.New("twitterapi: missing API key")

// New constructs a Client. It returns ErrMissingAPIKey if no key is provided
// and TWITTERAPIIO_API_KEY is not set in the environment.
func New(opts Options) (*Client, error) {
	apiKey := strings.TrimSpace(opts.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("TWITTERAPIIO_API_KEY"))
	}
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	wsURL := strings.TrimSpace(opts.WSURL)
	if wsURL == "" {
		wsURL = DefaultWSURL
	}

	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{
			Timeout: 45 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}

	ua := strings.TrimSpace(opts.UserAgent)
	if ua == "" {
		ua = defaultUA
	}

	maxRetries := opts.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}
	minBackoff := opts.MinBackoff
	if minBackoff <= 0 {
		minBackoff = 300 * time.Millisecond
	}
	maxBackoff := opts.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 8 * time.Second
	}

	store := opts.TokenStore
	if store == nil && strings.TrimSpace(opts.TokenFile) != "" {
		store = &FileTokenStore{Path: opts.TokenFile}
	}

	defaultProxy := strings.TrimSpace(opts.DefaultProxy)
	if defaultProxy == "" {
		defaultProxy = strings.TrimSpace(os.Getenv("TWITTERAPIIO_PROXY"))
	}

	c := &Client{
		apiKey:       apiKey,
		baseURL:      baseURL,
		wsURL:        wsURL,
		httpClient:   hc,
		userAgent:    ua,
		freePlan:     opts.FreePlan,
		maxRetries:   maxRetries,
		minBackoff:   minBackoff,
		maxBackoff:   maxBackoff,
		tokenStore:   store,
		defaultProxy: defaultProxy,
	}

	if cookie := strings.TrimSpace(opts.LoginCookie); cookie != "" {
		c.token = &LoginState{LoginCookie: cookie, SavedAtUnix: time.Now().Unix()}
	} else if store != nil {
		_ = c.loadToken()
	}

	c.Users = &UsersService{c: c}
	c.Tweets = &TweetsService{c: c}
	c.Communities = &CommunitiesService{c: c}
	c.Spaces = &SpacesService{c: c}
	c.Trends = &TrendsService{c: c}
	c.Lists = &ListsService{c: c}
	c.Account = &AccountService{c: c}
	c.Actions = &ActionsService{c: c}
	c.Media = &MediaService{c: c}
	c.Monitor = &MonitorService{c: c}
	c.Webhook = &WebhookService{c: c}
	c.Legacy = &LegacyService{c: c}

	return c, nil
}

// APIKey returns the configured API key. Useful for the WebSocket client
// (NewWSClient(c.APIKey(), c.WSURL())).
func (c *Client) APIKey() string { return c.apiKey }

// BaseURL returns the configured API base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// WSURL returns the configured WebSocket URL.
func (c *Client) WSURL() string { return c.wsURL }

// SetFreePlan toggles the 5s inter-request throttle.
func (c *Client) SetFreePlan(v bool) { c.freePlan = v }

// SetDefaultProxy updates the proxy injected into v2 write payloads when no
// per-call Proxy is set.
func (c *Client) SetDefaultProxy(p string) { c.defaultProxy = strings.TrimSpace(p) }

// LoginCookie returns the currently-cached login_cookies, or empty string.
func (c *Client) LoginCookie() string {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token == nil {
		return ""
	}
	return strings.TrimSpace(c.token.LoginCookie)
}

// SetLoginCookie installs a cookie manually (and persists via TokenStore if
// configured). Use after an out-of-band login.
func (c *Client) SetLoginCookie(cookie string) {
	cookie = strings.TrimSpace(cookie)
	if cookie == "" {
		return
	}
	st := &LoginState{LoginCookie: cookie, SavedAtUnix: time.Now().Unix()}
	c.tokenMu.Lock()
	c.token = st
	store := c.tokenStore
	c.tokenMu.Unlock()
	if store != nil {
		_ = store.Save(st)
	}
}

func (c *Client) loadToken() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.tokenStore == nil {
		return nil
	}
	st, err := c.tokenStore.Load()
	if err != nil {
		return err
	}
	c.token = st
	return nil
}

func (c *Client) clearToken() {
	c.tokenMu.Lock()
	c.token = nil
	c.tokenMu.Unlock()
}

func (c *Client) saveToken(st *LoginState) {
	c.tokenMu.Lock()
	c.token = st
	store := c.tokenStore
	c.tokenMu.Unlock()
	if store != nil {
		_ = store.Save(st)
	}
}

// pickProxy resolves the proxy string per priority: explicit override -> client
// default -> env TWITTERAPIIO_PROXY -> "".
func (c *Client) pickProxy(override string) string {
	if v := strings.TrimSpace(override); v != "" {
		return v
	}
	if c.defaultProxy != "" {
		return c.defaultProxy
	}
	if p := strings.TrimSpace(os.Getenv("TWITTERAPIIO_PROXY")); p != "" {
		return p
	}
	return ""
}

func (c *Client) throttleIfNeeded(ctx context.Context) error {
	if !c.freePlan {
		return nil
	}
	c.rlMu.Lock()
	defer c.rlMu.Unlock()

	wait := time.Until(c.lastReq.Add(5 * time.Second))
	if wait <= 0 {
		c.lastReq = time.Now()
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		c.lastReq = time.Now()
		return nil
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, q url.Values, body io.Reader, contentType string) (*http.Request, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	if err := c.throttleIfNeeded(ctx); err != nil {
		return nil, nil, err
	}

	// Snapshot body for retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := c.sleepBackoff(ctx, attempt, lastErr); err != nil {
				return nil, nil, err
			}
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if isRetryableNetErr(err) && attempt < c.maxRetries {
				continue
			}
			return nil, nil, err
		}

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, body, nil
		}

		apiErr := newAPIError(resp, body)
		lastErr = apiErr

		if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < c.maxRetries {
			continue
		}
		return resp, body, apiErr
	}
	return nil, nil, lastErr
}

func (c *Client) sleepBackoff(ctx context.Context, attempt int, lastErr error) error {
	// Honor Retry-After when the server sets it.
	if ae, ok := lastErr.(*APIError); ok {
		if d := parseRetryAfter(ae.Header); d > 0 {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				return nil
			}
		}
	}

	backoff := c.minBackoff * time.Duration(1<<uint(attempt-1))
	if backoff > c.maxBackoff {
		backoff = c.maxBackoff
	}
	jitter := time.Duration(rand.Int63n(int64(backoff/3 + 1)))
	t := time.NewTimer(backoff + jitter)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func isRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection reset") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "eof") ||
		strings.Contains(s, "tls handshake timeout") ||
		strings.Contains(s, "i/o timeout")
}

func parseRetryAfter(h http.Header) time.Duration {
	ra := strings.TrimSpace(h.Get("Retry-After"))
	if ra == "" {
		return 0
	}
	if n, err := strconv.Atoi(ra); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(ra); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func (c *Client) getJSON(ctx context.Context, path string, q url.Values, out any) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, q, nil, "")
	if err != nil {
		return err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	return decodeAndCheck(body, out)
}

func (c *Client) postJSON(ctx context.Context, path string, payload, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, nil, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	return decodeAndCheck(body, out)
}

func (c *Client) patchJSON(ctx context.Context, path string, payload, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPatch, path, nil, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	return decodeAndCheck(body, out)
}

// deleteJSON sends DELETE with a JSON body (twitterapi.io's delete_rule needs this).
func (c *Client) deleteJSON(ctx context.Context, path string, payload, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	return decodeAndCheck(body, out)
}

type multipartFunc func(w *multipart.Writer) error

func (c *Client) postMultipart(ctx context.Context, method, path string, form multipartFunc, out any) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := form(w); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := c.newRequest(ctx, method, path, nil, &buf, w.FormDataContentType())
	if err != nil {
		return err
	}
	_, body, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	return decodeAndCheck(body, out)
}

// decodeAndCheck unmarshals into out and surfaces semantic {status:"error"} as
// an APIError with HTTP 200 + the parsed message.
func decodeAndCheck(body []byte, out any) error {
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("twitterapi: decode response: %w (body=%s)", err, truncate(body, 256))
		}
	}
	if st, ok := extractStatus(body); ok && st.IsError() {
		return &APIError{
			StatusCode: http.StatusOK,
			Status:     st.Status,
			Message:    st.Message(),
			Code:       st.Code,
			Body:       body,
		}
	}
	return nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...(truncated)"
}
