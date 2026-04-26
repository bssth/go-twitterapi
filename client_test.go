package twitterapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("TWITTERAPIIO_API_KEY", "")
	if _, err := New(Options{}); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("want ErrMissingAPIKey, got %v", err)
	}
}

func TestNew_EnvFallback(t *testing.T) {
	t.Setenv("TWITTERAPIIO_API_KEY", "from-env")
	c, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if c.APIKey() != "from-env" {
		t.Fatalf("got %q", c.APIKey())
	}
}

func TestNew_DefaultsAndOverrides(t *testing.T) {
	c, err := New(Options{APIKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if c.BaseURL() != DefaultBaseURL {
		t.Errorf("BaseURL default: %q", c.BaseURL())
	}
	if c.WSURL() != DefaultWSURL {
		t.Errorf("WSURL default: %q", c.WSURL())
	}
	if c.maxRetries != 5 || c.minBackoff <= 0 || c.maxBackoff <= 0 {
		t.Errorf("retry defaults wrong: %+v", c)
	}
	c2, _ := New(Options{
		APIKey:     "k",
		BaseURL:    "https://x.example",
		WSURL:      "wss://y.example",
		MaxRetries: 1,
		MinBackoff: 5 * time.Millisecond,
		MaxBackoff: 10 * time.Millisecond,
	})
	if c2.BaseURL() != "https://x.example" || c2.WSURL() != "wss://y.example" || c2.maxRetries != 1 {
		t.Errorf("overrides not applied: %+v", c2)
	}
}

func TestSetLoginCookie_PersistsToStore(t *testing.T) {
	store := &MemoryTokenStore{}
	c, err := New(Options{APIKey: "k", TokenStore: store})
	if err != nil {
		t.Fatal(err)
	}
	c.SetLoginCookie("hello")
	if c.LoginCookie() != "hello" {
		t.Fatalf("LoginCookie() = %q", c.LoginCookie())
	}
	st, _ := store.Load()
	if st == nil || st.LoginCookie != "hello" {
		t.Fatalf("not persisted: %+v", st)
	}
}

func TestPickProxy_Priority(t *testing.T) {
	t.Setenv("TWITTERAPIIO_PROXY", "from-env")
	c, _ := New(Options{APIKey: "k", DefaultProxy: "from-default"})
	if got := c.pickProxy("override"); got != "override" {
		t.Errorf("override lost: %q", got)
	}
	if got := c.pickProxy(""); got != "from-default" {
		t.Errorf("default lost: %q", got)
	}
	c2, _ := New(Options{APIKey: "k"})
	if got := c2.pickProxy(""); got != "from-env" {
		t.Errorf("env lost: %q", got)
	}
}

func TestDo_RetriesOn500(t *testing.T) {
	var hits int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			http.Error(w, `{"status":"error","msg":"transient"}`, 500)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"userName":"x"}}`))
	}, func(o *Options) { o.MaxRetries = 5 })

	if _, err := c.Users.GetByUsername(context.Background(), "x"); err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if hits != 3 {
		t.Fatalf("expected 3 hits, got %d", hits)
	}
}

func TestDo_NoRetryOn400(t *testing.T) {
	var hits int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		http.Error(w, `{"status":"error","msg":"bad request"}`, 400)
	}, func(o *Options) { o.MaxRetries = 5 })

	_, err := c.Users.GetByUsername(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if hits != 1 {
		t.Fatalf("400 should not retry, got %d hits", hits)
	}
}

func TestDo_HonorsRetryAfter(t *testing.T) {
	var hits int32
	var first time.Time
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			first = time.Now()
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"userName":"x"}}`))
	}, func(o *Options) { o.MaxRetries = 2 })

	start := time.Now()
	if _, err := c.Users.GetByUsername(context.Background(), "x"); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("Retry-After 1s not honored: elapsed=%v (first hit at %v)", elapsed, first)
	}
}

func TestDo_NetworkErrorRetried(t *testing.T) {
	// Use a closed listener URL to force connection failure on the first try,
	// then point at a working server. Easier to test by spinning a server that
	// closes the connection mid-request.
	var hits int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			// hijack and slam the connection
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Skip("hijacker unavailable")
				return
			}
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte(`{"data":{"userName":"x"}}`))
	}, func(o *Options) { o.MaxRetries = 3 })

	if _, err := c.Users.GetByUsername(context.Background(), "x"); err != nil {
		t.Fatalf("after retry: %v", err)
	}
	if hits < 2 {
		t.Fatalf("expected retry, hits=%d", hits)
	}
}

func TestDo_HeadersInjected(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing x-api-key")
		}
		if !strings.Contains(r.Header.Get("User-Agent"), "go-twitterapi") {
			t.Errorf("User-Agent: %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	_, _ = c.Users.GetByUsername(context.Background(), "x")
}

func TestThrottle_FreePlan(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{}}`))
	}, func(o *Options) { o.FreePlan = false })

	// Force a near-now last request and check throttle waits.
	c.SetFreePlan(true)
	c.lastReq = time.Now().Add(-100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Users.GetByUsername(ctx, "x")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}

func TestContextCancelDuringDo(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := c.Users.GetByUsername(ctx, "x")
	if err == nil {
		t.Fatal("expected error")
	}
}
