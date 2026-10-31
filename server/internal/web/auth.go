package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	adminCookieName    = "agelan_admin"
	setupCookieName    = "agelan_setup"
	adminSessionTTL    = 12 * time.Hour
	pbkdf2Iterations   = 210_000
	pbkdf2KeyLength    = 32
	pbkdf2SaltLength   = 16
	loginMaxFailures   = 10
	loginFailureWindow = 5 * time.Minute
)

// hashPassword derives a PBKDF2-HMAC-SHA256 hash. Only the standard library is
// used, so the dashboard has no extra dependency to ship.
func hashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password is empty")
	}
	salt := make([]byte, pbkdf2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := pbkdf2SHA256([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLength)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s",
		pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// verifyPassword compares a password with an encoded hash in constant time.
func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

// pbkdf2SHA256 is the RFC 8018 PBKDF2 with HMAC-SHA256.
func pbkdf2SHA256(password, salt []byte, iterations, keyLength int) []byte {
	hashLength := sha256.Size
	blocks := (keyLength + hashLength - 1) / hashLength
	derived := make([]byte, 0, blocks*hashLength)
	prf := hmac.New(sha256.New, password)
	for block := 1; block <= blocks; block++ {
		prf.Reset()
		_, _ = prf.Write(salt)
		_, _ = prf.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		sum := prf.Sum(nil)
		t := append([]byte(nil), sum...)
		for iteration := 1; iteration < iterations; iteration++ {
			prf.Reset()
			_, _ = prf.Write(sum)
			sum = prf.Sum(nil)
			for index := range t {
				t[index] ^= sum[index]
			}
		}
		derived = append(derived, t...)
	}
	return derived[:keyLength]
}

// adminSession is one logged-in browser.
type adminSession struct {
	token     string
	csrf      string
	expiresAt time.Time
	remoteIP  string
}

type loginAttempts struct {
	count     int
	firstSeen time.Time
}

// sessionManager keeps the admin sessions in memory: restarting the server
// logs everybody out, which is the right default for a LAN tool.
type sessionManager struct {
	mu       sync.Mutex
	sessions map[string]*adminSession
	attempts map[string]*loginAttempts
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		sessions: make(map[string]*adminSession),
		attempts: make(map[string]*loginAttempts),
	}
}

func randomToken(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func (m *sessionManager) create(remoteIP string) (*adminSession, error) {
	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	session := &adminSession{
		token:     token,
		csrf:      csrf,
		expiresAt: time.Now().Add(adminSessionTTL),
		remoteIP:  remoteIP,
	}
	m.mu.Lock()
	m.sessions[session.token] = session
	m.mu.Unlock()
	return session, nil
}

func (m *sessionManager) get(token string) (*adminSession, bool) {
	if token == "" {
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[token]
	if !ok {
		return nil, false
	}
	if time.Now().After(session.expiresAt) {
		delete(m.sessions, token)
		return nil, false
	}
	return session, true
}

func (m *sessionManager) delete(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

// allowAttempt reports whether an IP may try to log in. Successful logins
// reset the counter.
func (m *sessionManager) allowAttempt(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	attempts, ok := m.attempts[ip]
	if !ok {
		return true
	}
	if time.Since(attempts.firstSeen) > loginFailureWindow {
		delete(m.attempts, ip)
		return true
	}
	return attempts.count < loginMaxFailures
}

func (m *sessionManager) failAttempt(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	attempts, ok := m.attempts[ip]
	if !ok || time.Since(attempts.firstSeen) > loginFailureWindow {
		m.attempts[ip] = &loginAttempts{count: 1, firstSeen: time.Now()}
		return
	}
	attempts.count++
}

func (m *sessionManager) resetAttempts(ip string) {
	m.mu.Lock()
	delete(m.attempts, ip)
	m.mu.Unlock()
}

// effectivePasswordHash returns the hash stored during first-time setup or
// through the authenticated password-change form.
func (s *Server) effectivePasswordHash() (string, bool, error) {
	if s.store != nil {
		hash, ok, err := s.store.Get("admin_password_hash")
		if err != nil {
			return "", false, err
		}
		if ok && hash != "" {
			return hash, true, nil
		}
	}
	return "", false, nil
}

// passwordAccepted verifies a password against the stored password hash.
func (s *Server) passwordAccepted(password string) (bool, error) {
	if password == "" {
		return false, nil
	}
	hash, configured, err := s.effectivePasswordHash()
	if err != nil {
		return false, err
	}
	return configured && verifyPassword(hash, password), nil
}

// passwordConfigured reports whether the admin panel can be used at all.
func (s *Server) passwordConfigured() (bool, error) {
	_, configured, err := s.effectivePasswordHash()
	return configured, err
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// currentSession returns the session of the request, if any.
func (s *Server) currentSession(r *http.Request) (*adminSession, bool) {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil {
		return nil, false
	}
	return s.sessions.get(cookie.Value)
}

// requireAdmin is the guard of every admin API. It also enforces the CSRF
// token on mutating requests.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (*adminSession, bool) {
	session, ok := s.currentSession(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "admin session required")
		return nil, false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf")
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(session.csrf)) != 1 {
			writeJSONError(w, http.StatusForbidden, "invalid CSRF token")
			return nil, false
		}
	}
	return session, true
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	configured, err := s.passwordConfigured()
	if err != nil {
		http.Error(w, "could not read admin configuration", http.StatusInternalServerError)
		return
	}
	if !configured {
		http.Redirect(w, r, "/admin/setup", http.StatusSeeOther)
		return
	}
	if _, ok := s.currentSession(r); ok {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	s.render(w, r, "login", map[string]any{
		"Title": "Admin",
		"Error": r.URL.Query().Get("error"),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := remoteIP(r)
	if !s.sessions.allowAttempt(ip) {
		http.Redirect(w, r, "/admin/login?error=Too+many+attempts.+Wait+a+few+minutes.", http.StatusSeeOther)
		return
	}
	password := r.FormValue("password")
	accepted, err := s.passwordAccepted(password)
	if err != nil {
		http.Error(w, "could not verify admin password", http.StatusInternalServerError)
		return
	}
	if !accepted {
		s.sessions.failAttempt(ip)
		http.Redirect(w, r, "/admin/login?error=Invalid+password.", http.StatusSeeOther)
		return
	}
	s.sessions.resetAttempts(ip)
	session, err := s.sessions.create(ip)
	if err != nil {
		http.Error(w, "could not create admin session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    session.token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   int(adminSessionTTL.Seconds()),
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if session, ok := s.currentSession(r); ok {
		s.sessions.delete(session.token)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (s *Server) handleSetupPage(w http.ResponseWriter, r *http.Request) {
	configured, err := s.passwordConfigured()
	if err != nil {
		http.Error(w, "could not read admin configuration", http.StatusInternalServerError)
		return
	}
	if configured {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	token, err := randomToken(32)
	if err != nil {
		http.Error(w, "could not create setup token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     setupCookieName,
		Value:    token,
		Path:     "/admin/setup",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   10 * 60,
	})
	s.render(w, r, "setup", map[string]any{
		"Title":      "Set up admin",
		"Error":      r.URL.Query().Get("error"),
		"SetupToken": token,
		"SiteTitle":  s.opts.Config.PublicTitle,
	})
}

// handleSetup stores the first admin password. It is only reachable while no
// password is configured, so a fresh install can be claimed from the LAN.
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	configured, err := s.passwordConfigured()
	if err != nil {
		http.Error(w, "could not read admin configuration", http.StatusInternalServerError)
		return
	}
	if configured {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	cookie, cookieErr := r.Cookie(setupCookieName)
	token := r.FormValue("token")
	if cookieErr != nil || token == "" ||
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) != 1 {
		http.Error(w, "invalid setup token; reload the setup page", http.StatusForbidden)
		return
	}
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")
	publicTitle := strings.TrimSpace(r.FormValue("siteTitle"))
	if publicTitle == "" || utf8.RuneCountInString(publicTitle) > 64 {
		http.Redirect(w, r, "/admin/setup?error=Site+name+must+contain+1+to+64+characters.", http.StatusSeeOther)
		return
	}
	if len(password) < 8 {
		http.Redirect(w, r, "/admin/setup?error=Use+at+least+8+characters.", http.StatusSeeOther)
		return
	}
	if password != confirm {
		http.Redirect(w, r, "/admin/setup?error=The+passwords+do+not+match.", http.StatusSeeOther)
		return
	}
	hash, err := hashPassword(password)
	if err != nil {
		http.Redirect(w, r, "/admin/setup?error="+err.Error(), http.StatusSeeOther)
		return
	}
	created, err := s.store.CompleteSetup(hash, publicTitle)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !created {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	s.opts.Config.PublicTitle = publicTitle
	http.SetCookie(w, &http.Cookie{
		Name: setupCookieName, Value: "", Path: "/admin/setup", HttpOnly: true,
		Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
