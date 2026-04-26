package twitterapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMediaService_Upload_Multipart(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/twitter/upload_media_v2" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Fatalf("ct=%q", ct)
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.MultipartForm.Value["login_cookies"]; len(got) == 0 || got[0] != "C" {
			t.Errorf("login_cookies=%v", got)
		}
		if got := r.MultipartForm.Value["proxy"]; len(got) == 0 || !strings.Contains(got[0], "proxy.example") {
			t.Errorf("proxy=%v", got)
		}
		fhs := r.MultipartForm.File["file"]
		if len(fhs) == 0 {
			t.Fatal("file part missing")
		}
		f, _ := fhs[0].Open()
		defer f.Close()
		body, _ := io.ReadAll(f)
		if string(body) != "PNGDATA" {
			t.Errorf("file body=%q", body)
		}
		_, _ = w.Write([]byte(`{"media_id":"m1","status":"success","msg":"ok"}`))
	}, func(o *Options) { o.LoginCookie = "C" })

	resp, err := c.Media.Upload(context.Background(), "img.png", bytes.NewReader([]byte("PNGDATA")), nil)
	if err != nil || resp.MediaID != "m1" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
}

func TestMediaService_Upload_Validation(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "C", DefaultProxy: "p"})
	if _, err := c.Media.Upload(context.Background(), "", bytes.NewReader([]byte("x")), nil); err == nil {
		t.Fatal("expected error: filename")
	}
	if _, err := c.Media.Upload(context.Background(), "f", nil, nil); err == nil {
		t.Fatal("expected error: nil body")
	}
}

func TestMediaService_UpdateProfile_PATCH(t *testing.T) {
	var got map[string]any
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/twitter/update_profile_v2" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		decodeBody(t, r, &got)
		_, _ = w.Write([]byte(`{"status":"success","msg":"ok"}`))
	}, func(o *Options) { o.LoginCookie = "C" })

	name := "Mr. Burns"
	desc := "Excellent"
	if _, err := c.Media.UpdateProfile(context.Background(), UpdateProfileParams{
		Name: &name, Description: &desc,
	}); err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Mr. Burns" || got["description"] != "Excellent" {
		t.Errorf("body=%+v", got)
	}
}

func TestMediaService_UpdateProfile_RequiresAField(t *testing.T) {
	c, _ := New(Options{APIKey: "k", LoginCookie: "C", DefaultProxy: "p"})
	if _, err := c.Media.UpdateProfile(context.Background(), UpdateProfileParams{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestMediaService_UpdateAvatar_PATCH_Multipart(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/twitter/update_avatar_v2" {
			t.Fatalf("method=%s path=%s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}, func(o *Options) { o.LoginCookie = "C" })

	if _, err := c.Media.UpdateAvatar(context.Background(), "a.png", bytes.NewReader([]byte("X")), ""); err != nil {
		t.Fatal(err)
	}
}
