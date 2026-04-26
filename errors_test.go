package twitterapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAPIStatus_IsError(t *testing.T) {
	if (APIStatus{Status: StatusError}).IsError() != true {
		t.Fatal("error not detected")
	}
	if (APIStatus{Status: StatusSuccess}).IsError() != false {
		t.Fatal("success flagged as error")
	}
}

func TestAPIStatus_Message_PrefersMsg(t *testing.T) {
	got := APIStatus{Msg: "a", MsgAlt: "b"}.Message()
	if got != "a" {
		t.Fatalf("got %q", got)
	}
	got = APIStatus{MsgAlt: "b"}.Message()
	if got != "b" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractStatus(t *testing.T) {
	cases := []struct {
		name  string
		body  string
		want  bool
		check func(APIStatus) bool
	}{
		{"empty", "", false, nil},
		{"none", `{"data":{}}`, false, nil},
		{"msg", `{"status":"error","msg":"x"}`, true, func(s APIStatus) bool { return s.Message() == "x" }},
		{"message field", `{"status":"error","message":"y"}`, true, func(s APIStatus) bool { return s.Message() == "y" }},
		{"only msg w/o status", `{"msg":"informational"}`, true, func(s APIStatus) bool { return s.Msg == "informational" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, ok := extractStatus([]byte(c.body))
			if ok != c.want {
				t.Fatalf("ok=%v", ok)
			}
			if c.check != nil && !c.check(s) {
				t.Fatalf("check failed: %+v", s)
			}
		})
	}
}

func TestAPIError_Format(t *testing.T) {
	e := &APIError{StatusCode: 500, Message: "boom"}
	if !strings.Contains(e.Error(), "http 500") {
		t.Fatal(e.Error())
	}
	e2 := &APIError{StatusCode: 200, Message: "semantic"}
	if !strings.Contains(e2.Error(), "api error") {
		t.Fatal(e2.Error())
	}
	e3 := &APIError{StatusCode: 500, Body: []byte("plain text")}
	if !strings.Contains(e3.Error(), "plain text") {
		t.Fatal(e3.Error())
	}
}

func TestAPIError_IsSentinels(t *testing.T) {
	if !errors.Is(&APIError{StatusCode: http.StatusPaymentRequired}, ErrInsufficientCredits) {
		t.Fatal("402 should map to ErrInsufficientCredits")
	}
	expired := &APIError{
		StatusCode: 200,
		Message:    "login_cookies are invalid",
		Body:       []byte(`{"msg":"login_cookies invalid"}`),
	}
	if !errors.Is(expired, ErrCookieExpired) {
		t.Fatal("should detect expired cookie")
	}
	if errors.Is(&APIError{StatusCode: 400, Message: "bad req"}, ErrCookieExpired) {
		t.Fatal("false positive on cookie expiry")
	}
}

func TestNewAPIError_FromDetail(t *testing.T) {
	resp := &http.Response{StatusCode: 422, Header: http.Header{}}
	body := []byte(`{"detail":"validation failed"}`)
	e := newAPIError(resp, body)
	if e.Message != "validation failed" {
		t.Fatalf("detail not surfaced: %q", e.Message)
	}
}

func TestDecodeAndCheck_SemanticError(t *testing.T) {
	body := []byte(`{"status":"error","msg":"nope"}`)
	var dst struct {
		APIStatus
	}
	err := decodeAndCheck(body, &dst)
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *APIError
	if !errors.As(err, &ae) || ae.StatusCode != 200 {
		t.Fatalf("not APIError 200: %v", err)
	}
}
