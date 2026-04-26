package twitterapi

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

// LoginState is what a TokenStore persists. RawLoginReply keeps the original
// API payload for debugging / future fields.
type LoginState struct {
	LoginCookie   string          `json:"login_cookies"`
	SavedAtUnix   int64           `json:"saved_at_unix"`
	RawLoginReply json.RawMessage `json:"raw_login_reply,omitempty"`
}

// TokenStore persists and loads the v2 login cookie. Implementations must be
// safe for concurrent use.
type TokenStore interface {
	Load() (*LoginState, error)
	Save(*LoginState) error
}

// FileTokenStore stores LoginState as JSON at Path. If the file contains a
// bare cookie string (legacy), it is migrated transparently on Load.
type FileTokenStore struct {
	Path string
	mu   sync.Mutex
}

// Load reads the cookie file. Returns (nil, nil) if the file is missing or empty.
func (s *FileTokenStore) Load() (*LoginState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	txt := strings.TrimSpace(string(b))
	if txt == "" {
		return nil, nil
	}
	var st LoginState
	if err := json.Unmarshal([]byte(txt), &st); err != nil {
		// Legacy: file contains the raw cookie value.
		st = LoginState{LoginCookie: txt, SavedAtUnix: time.Now().Unix()}
	}
	if strings.TrimSpace(st.LoginCookie) == "" {
		return nil, nil
	}
	return &st, nil
}

// Save writes the cookie to disk with 0600 permissions.
func (s *FileTokenStore) Save(st *LoginState) error {
	if st == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if st.SavedAtUnix == 0 {
		st.SavedAtUnix = time.Now().Unix()
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, b, 0o600)
}

// MemoryTokenStore is a TokenStore backed by RAM. Useful for tests and
// short-lived processes.
type MemoryTokenStore struct {
	mu sync.Mutex
	st *LoginState
}

func (m *MemoryTokenStore) Load() (*LoginState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.st == nil {
		return nil, nil
	}
	cp := *m.st
	return &cp, nil
}

func (m *MemoryTokenStore) Save(st *LoginState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if st == nil {
		m.st = nil
		return nil
	}
	cp := *st
	m.st = &cp
	return nil
}
