package twitterapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors. Use errors.Is to test.
var (
	// ErrNeedLogin is returned when a v2 write action is invoked without a
	// cached login_cookies and the auto-login env vars are absent.
	ErrNeedLogin = errors.New("twitterapi: login_cookies required (set Options.LoginCookie, configure TokenStore + env, or call Account.LoginV2)")

	// ErrCookieExpired is returned when the API rejects login_cookies. The
	// SDK retries once with a fresh login when possible.
	ErrCookieExpired = errors.New("twitterapi: login_cookies expired or invalid")

	// ErrInsufficientCredits maps HTTP 402 Payment Required.
	ErrInsufficientCredits = errors.New("twitterapi: insufficient credits (HTTP 402)")
)

// Status mirrors the API's "status" field on response envelopes.
type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

// APIStatus is the common envelope embedded in most responses. The API is
// inconsistent about field names (msg vs message), so both are captured.
type APIStatus struct {
	Status  Status `json:"status,omitempty"`
	Msg     string `json:"msg,omitempty"`
	MsgAlt  string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
	HTTPErr int    `json:"error,omitempty"`
}

// IsError reports whether the envelope explicitly signals an error.
func (s APIStatus) IsError() bool { return s.Status == StatusError }

// Message returns the populated error message regardless of which field the
// API used.
func (s APIStatus) Message() string {
	if s.Msg != "" {
		return s.Msg
	}
	return s.MsgAlt
}

// extractStatus pulls the envelope from raw JSON without forcing the caller's
// type to embed APIStatus.
func extractStatus(body []byte) (APIStatus, bool) {
	if len(body) == 0 {
		return APIStatus{}, false
	}
	var s APIStatus
	if err := json.Unmarshal(body, &s); err != nil {
		return APIStatus{}, false
	}
	if s.Status == "" && s.Msg == "" && s.MsgAlt == "" {
		return APIStatus{}, false
	}
	return s, true
}

// APIError is the error type returned for non-2xx HTTP responses and for
// successful HTTP responses whose body says {"status":"error"}.
type APIError struct {
	StatusCode int         // HTTP status code (200 for semantic errors)
	Status     Status      // envelope "status"
	Message    string      // envelope "msg" or "message"
	Code       int         // envelope "code" if any
	Header     http.Header // response headers (nil for semantic 200 errors)
	Body       []byte      // raw response body
}

func (e *APIError) Error() string {
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		// Fall back to a body excerpt.
		msg = truncate(e.Body, 200)
	}
	if e.StatusCode == http.StatusOK {
		return fmt.Sprintf("twitterapi: api error: %s", msg)
	}
	return fmt.Sprintf("twitterapi: http %d: %s", e.StatusCode, msg)
}

// Is supports errors.Is for sentinel mapping.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrInsufficientCredits:
		return e.StatusCode == http.StatusPaymentRequired
	case ErrCookieExpired:
		return looksLikeCookieExpired(e)
	}
	return false
}

func newAPIError(resp *http.Response, body []byte) *APIError {
	ae := &APIError{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	}
	if st, ok := extractStatus(body); ok {
		ae.Status = st.Status
		ae.Message = st.Message()
		ae.Code = st.Code
	}
	if ae.Message == "" {
		// FastAPI returns {"detail": "..."}.
		var d struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(body, &d) == nil && d.Detail != "" {
			ae.Message = d.Detail
		}
	}
	return ae
}

func looksLikeCookieExpired(err error) bool {
	if err == nil {
		return false
	}
	var ae *APIError
	if errors.As(err, &ae) {
		s := strings.ToLower(ae.Message + " " + string(ae.Body))
		return strings.Contains(s, "login_cookies") &&
			(strings.Contains(s, "invalid") || strings.Contains(s, "expired") || strings.Contains(s, "faulty"))
	}
	return false
}
