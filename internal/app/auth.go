package app

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	defaultLoginUsername     = "admin"
	defaultLoginPassword     = "password"
	defaultSessionCookieName = "ros_list_session"
	defaultSessionTTLMinutes = 720

	passwordHashPrefix = "pbkdf2-sha256"
	passwordIterations = 120000
	passwordSaltSize   = 16
	passwordKeyLen     = 32
)

type Session struct {
	ID                    string
	Username              string
	RequirePasswordChange bool
	ExpiresAt             time.Time
	CreatedAt             time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = time.Duration(defaultSessionTTLMinutes) * time.Minute
	}
	return &SessionManager{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

func (m *SessionManager) Create(username string, requirePasswordChange bool) (Session, error) {
	id, err := randomToken(32)
	if err != nil {
		return Session{}, err
	}

	now := time.Now()
	sess := Session{
		ID:                    id,
		Username:              username,
		RequirePasswordChange: requirePasswordChange,
		CreatedAt:             now,
		ExpiresAt:             now.Add(m.ttl),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[id] = sess
	return sess, nil
}

func (m *SessionManager) Get(id string) (Session, bool) {
	if id == "" {
		return Session{}, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return Session{}, false
	}
	if time.Now().After(sess.ExpiresAt) {
		delete(m.sessions, id)
		return Session{}, false
	}
	return sess, true
}

func (m *SessionManager) Delete(id string) {
	if id == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

func (m *SessionManager) Touch(id string) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return Session{}, false
	}
	if time.Now().After(sess.ExpiresAt) {
		delete(m.sessions, id)
		return Session{}, false
	}
	sess.ExpiresAt = time.Now().Add(m.ttl)
	m.sessions[id] = sess
	return sess, true
}

func (m *SessionManager) UpdateRequirePasswordChange(id string, required bool) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return Session{}, false
	}
	if time.Now().After(sess.ExpiresAt) {
		delete(m.sessions, id)
		return Session{}, false
	}
	sess.RequirePasswordChange = required
	sess.ExpiresAt = time.Now().Add(m.ttl)
	m.sessions[id] = sess
	return sess, true
}

func resolveLoginUsername(cfg AppConfig) string {
	username := strings.TrimSpace(cfg.Server.LoginUsername)
	if username == "" {
		return defaultLoginUsername
	}
	return username
}

func verifyLoginPassword(cfg AppConfig, password string) bool {
	password = strings.TrimSpace(password)
	if password == "" {
		return false
	}

	if hash := strings.TrimSpace(cfg.Server.LoginPasswordHash); hash != "" {
		return verifyPasswordHash(password, hash)
	}

	expected := strings.TrimSpace(cfg.Server.LoginPassword)
	if expected == "" {
		expected = defaultLoginPassword
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(expected)) == 1
}

func hashNewLoginPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := pbkdf2SHA256([]byte(password), salt, passwordIterations, passwordKeyLen)
	return fmt.Sprintf("%s$%d$%s$%s", passwordHashPrefix, passwordIterations, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

func verifyPasswordHash(password, encoded string) bool {
	parts := strings.Split(strings.TrimSpace(encoded), "$")
	if len(parts) != 4 {
		return false
	}
	if parts[0] != passwordHashPrefix {
		return false
	}

	iterations, err := parsePositiveInt(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return false
	}
	expected, err := hex.DecodeString(parts[3])
	if err != nil || len(expected) == 0 {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(got, expected) == 1
}

func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	hLen := sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	var out []byte

	for block := 1; block <= numBlocks; block++ {
		u := pbkdf2PRF(password, salt, block)
		t := make([]byte, len(u))
		copy(t, u)
		for i := 1; i < iter; i++ {
			u = pbkdf2PRF(password, u, 0)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}

	return out[:keyLen]
}

func pbkdf2PRF(password, data []byte, blockIndex int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(data)
	if blockIndex > 0 {
		mac.Write([]byte{
			byte(blockIndex >> 24),
			byte(blockIndex >> 16),
			byte(blockIndex >> 8),
			byte(blockIndex),
		})
	}
	return mac.Sum(nil)
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func parsePositiveInt(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty int")
	}
	n := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid int")
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

func validateNewPassword(password string) error {
	password = strings.TrimSpace(password)
	switch {
	case password == "":
		return fmt.Errorf("新密码不能为空")
	case len(password) < 8:
		return fmt.Errorf("新密码长度不能小于 8 位")
	case password == defaultLoginPassword:
		return fmt.Errorf("新密码不能继续使用默认密码")
	default:
		return nil
	}
}
