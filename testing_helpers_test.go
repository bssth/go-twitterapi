package twitterapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient builds a Client whose BaseURL points at a freshly-spun
// httptest.Server using the supplied handler. Retries are disabled to keep
// test runs fast; bump MaxRetries via the `tweak` callback if needed.
func newTestClient(t *testing.T, h http.HandlerFunc, tweak ...func(*Options)) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	opts := Options{
		APIKey:       "test-key",
		BaseURL:      srv.URL,
		WSURL:        strings.Replace(srv.URL, "http", "ws", 1) + "/ws",
		MaxRetries:   0,
		MinBackoff:   1 * time.Millisecond,
		MaxBackoff:   2 * time.Millisecond,
		DefaultProxy: "http://proxy.example:8080",
	}
	for _, f := range tweak {
		f(&opts)
	}
	c, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

// jsonRoute is a tiny mux: matches by path then writes the JSON payload.
type jsonRoute struct {
	method string
	path   string
	status int
	body   any
	check  func(t *testing.T, r *http.Request)
}

// muxOf returns a handler that dispatches by (method, path) to the first
// matching route. If nothing matches the test fails — every request must be
// expected.
func muxOf(t *testing.T, routes ...jsonRoute) http.HandlerFunc {
	t.Helper()
	hits := make([]int32, len(routes))
	t.Cleanup(func() {
		for i, r := range routes {
			if atomic.LoadInt32(&hits[i]) == 0 {
				t.Errorf("route %s %s never hit", r.method, r.path)
			}
		}
	})
	return func(w http.ResponseWriter, r *http.Request) {
		for i := range routes {
			rt := routes[i]
			if (rt.method == "" || rt.method == r.Method) && rt.path == r.URL.Path {
				atomic.AddInt32(&hits[i], 1)
				if got := r.Header.Get("x-api-key"); got != "test-key" {
					t.Errorf("missing/wrong x-api-key: %q", got)
				}
				if rt.check != nil {
					rt.check(t, r)
				}
				status := rt.status
				if status == 0 {
					status = http.StatusOK
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if rt.body != nil {
					_ = json.NewEncoder(w).Encode(rt.body)
				}
				return
			}
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "no route", http.StatusNotFound)
	}
}

// decodeBody unmarshals the request body into v.
func decodeBody(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Errorf("decode body: %v", err)
	}
}

// wantQuery asserts that r has each given query value.
func wantQuery(t *testing.T, r *http.Request, kv map[string]string) {
	t.Helper()
	q := r.URL.Query()
	for k, v := range kv {
		if got := q.Get(k); got != v {
			t.Errorf("query %q = %q, want %q", k, got, v)
		}
	}
}
