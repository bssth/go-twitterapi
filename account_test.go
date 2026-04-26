package twitterapi

import (
	"context"
	"net/http"
	"testing"
)

func TestAccountService_Info(t *testing.T) {
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "GET", path: "/oapi/my/info",
		body: map[string]any{"recharge_credits": 12.5, "total_bonus_credits": 1.25},
	}))
	info, err := c.Account.Info(context.Background())
	if err != nil || info.RechargeCredits != 12.5 || info.TotalBonusCredits != 1.25 {
		t.Fatalf("info=%+v err=%v", info, err)
	}
}

func TestAccountService_LoginV2_Persists(t *testing.T) {
	store := &MemoryTokenStore{}
	c, _ := newTestClient(t, muxOf(t, jsonRoute{
		method: "POST", path: "/twitter/user_login_v2",
		body: map[string]any{"login_cookies": "COOKIE", "status": "success", "msg": "ok"},
		check: func(t *testing.T, r *http.Request) {
			var got map[string]any
			decodeBody(t, r, &got)
			if got["user_name"] != "u" || got["totp_secret"] != "t" {
				t.Errorf("body=%+v", got)
			}
		},
	}), func(o *Options) { o.TokenStore = store })

	cookie, err := c.Account.LoginV2(context.Background(), LoginV2Params{
		UserName: "u", Email: "e", Password: "p", TOTPSecret: "t",
	})
	if err != nil || cookie != "COOKIE" {
		t.Fatalf("cookie=%q err=%v", cookie, err)
	}
	st, _ := store.Load()
	if st == nil || st.LoginCookie != "COOKIE" {
		t.Fatalf("not persisted: %+v", st)
	}
}

func TestAccountService_LoginV2_RequiresProxy(t *testing.T) {
	c, _ := New(Options{APIKey: "k"})
	if _, err := c.Account.LoginV2(context.Background(), LoginV2Params{
		UserName: "u", Email: "e", Password: "p",
	}); err == nil {
		t.Fatal("expected error for missing proxy")
	}
}

func TestAccountService_LoginV2_ServerError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","msg":"login failed"}`))
	})
	_, err := c.Account.LoginV2(context.Background(), LoginV2Params{
		UserName: "u", Email: "e", Password: "p",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAccountService_EnsureLogin_FromCache(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "cached"})
	got, err := c.Account.EnsureLogin(context.Background())
	if err != nil || got != "cached" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestAccountService_EnsureLogin_NeedsEnv(t *testing.T) {
	t.Setenv("TWITTERAPIIO_USER_NAME", "")
	t.Setenv("TWITTERAPIIO_EMAIL", "")
	t.Setenv("TWITTERAPIIO_PASSWORD", "")
	c, _ := New(Options{APIKey: "k"})
	_, err := c.Account.EnsureLogin(context.Background())
	if err == nil {
		t.Fatal("expected ErrNeedLogin")
	}
}
