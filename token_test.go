package twitterapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileTokenStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &FileTokenStore{Path: filepath.Join(dir, "tok.json")}
	if got, err := s.Load(); err != nil || got != nil {
		t.Fatalf("missing file: got=%v err=%v", got, err)
	}
	if err := s.Save(&LoginState{LoginCookie: "abc"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.LoginCookie != "abc" || got.SavedAtUnix == 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestFileTokenStore_LegacyRawCookie(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tok.txt")
	// Pre-populate with a bare cookie string (legacy format).
	if err := writeFile(path, []byte("bare-cookie-value")); err != nil {
		t.Fatal(err)
	}
	s := &FileTokenStore{Path: path}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.LoginCookie != "bare-cookie-value" {
		t.Fatalf("legacy migration failed: %+v", got)
	}
}

func TestFileTokenStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tok")
	if err := writeFile(path, []byte("   \n")); err != nil {
		t.Fatal(err)
	}
	s := &FileTokenStore{Path: path}
	got, err := s.Load()
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestMemoryTokenStore(t *testing.T) {
	s := &MemoryTokenStore{}
	got, _ := s.Load()
	if got != nil {
		t.Fatal("empty store should load nil")
	}
	_ = s.Save(&LoginState{LoginCookie: "x"})
	got, _ = s.Load()
	if got.LoginCookie != "x" {
		t.Fatal("not saved")
	}
	// Mutating the returned copy must not affect the store.
	got.LoginCookie = "mutated"
	again, _ := s.Load()
	if again.LoginCookie != "x" {
		t.Fatalf("aliasing leak: %+v", again)
	}
	// Save(nil) clears.
	_ = s.Save(nil)
	if v, _ := s.Load(); v != nil {
		t.Fatal("nil save should clear")
	}
}

func writeFile(path string, b []byte) error {
	return os.WriteFile(path, b, 0o600)
}
