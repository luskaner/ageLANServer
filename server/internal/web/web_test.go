package web

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/luskaner/ageServer/common/battleServer"
	"github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/models"
)

func TestPublicPagesAndAdminPasswordLifecycle(t *testing.T) {
	dir := t.TempDir()
	staticStore, err := models.InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = staticStore.Close() })
	if err = models.SaveStaticBattleServer(models.StaticBattleServer{
		GameID: "age2", Region: "remote", IPv4: "192.168.10.25", BsPort: 30301, WebSocketPort: 30302,
	}); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte("[Web]\nEnabled = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := New(Options{
		Config:     internal.Web{PublicTitle: "Test Server"},
		ConfigFile: configPath,
		DataDir:    dir,
		Games:      map[string]models.Game{},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close dashboard: %v", err)
		}
	})

	for _, path := range []string{
		"/", "/lobbies", "/queue", "/matches", "/leaderboard", "/players", "/maps", "/battle-servers",
		"/admin/setup", "/api/info", "/api/health", "/static/images/civs/aztecs.png",
		"/static/images/maps/arabia.png", "/static/images/map-placeholder.svg", "/static/images/civs/unknown.svg",
	} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200; body: %s", path, response.Code, response.Body.String())
		}
	}
	homeResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(homeResponse, httptest.NewRequest(http.MethodGet, "/", nil))
	if !bytes.Contains(homeResponse.Body.Bytes(), []byte("Game Content Usage Rules")) {
		t.Error("dashboard footer is missing the Microsoft game-content notice")
	}
	infoResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(infoResponse, httptest.NewRequest(http.MethodGet, "/api/info", nil))
	var info map[string]any
	if err = json.Unmarshal(infoResponse.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["name"] != "Test Server" {
		t.Errorf("public API name = %v, want the configured brand", info["name"])
	}
	for _, field := range []string{"version", "go", "os", "configFile"} {
		if _, exists := info[field]; exists {
			t.Errorf("public API exposed implementation detail %q", field)
		}
	}

	setupPage := httptest.NewRecorder()
	server.Handler().ServeHTTP(setupPage, httptest.NewRequest(http.MethodGet, "/admin/setup", nil))
	tokenMatch := regexp.MustCompile(`name="token" value="([^"]+)"`).FindSubmatch(setupPage.Body.Bytes())
	if len(tokenMatch) != 2 {
		t.Fatal("setup page did not render a setup token")
	}
	var setupCookie *http.Cookie
	for _, candidate := range setupPage.Result().Cookies() {
		if candidate.Name == setupCookieName {
			setupCookie = candidate
		}
	}
	if setupCookie == nil || setupCookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("setup page did not set a same-site setup cookie")
	}

	unauthorizedSetup := httptest.NewRecorder()
	server.Handler().ServeHTTP(unauthorizedSetup, httptest.NewRequest(http.MethodPost, "/admin/setup", bytes.NewBufferString(
		"password=attacker-passphrase&confirm=attacker-passphrase&token=guess",
	)))
	if unauthorizedSetup.Code != http.StatusForbidden {
		t.Fatalf("setup without the browser token = %d, want 403", unauthorizedSetup.Code)
	}

	setup := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/admin/setup", bytes.NewBufferString(url.Values{
		"siteTitle": {"My Community"},
		"password":  {"secure-passphrase"},
		"confirm":   {"secure-passphrase"},
		"token":     {string(tokenMatch[1])},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(setupCookie)
	server.Handler().ServeHTTP(setup, request)
	if setup.Code != http.StatusSeeOther {
		t.Fatalf("setup status = %d, want redirect; body: %s", setup.Code, setup.Body.String())
	}
	if server.opts.Config.PublicTitle != "My Community" {
		t.Fatalf("site title = %q, want custom setup title", server.opts.Config.PublicTitle)
	}
	if title, ok, getErr := server.store.Get("public_title"); getErr != nil || !ok || title != "My Community" {
		t.Fatalf("persisted site title = %q, found=%t, err=%v", title, ok, getErr)
	}
	accepted := func(password string) bool {
		t.Helper()
		ok, err := server.passwordAccepted(password)
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	if !accepted("secure-passphrase") {
		t.Fatal("configured password was not accepted")
	}
	if accepted("wrong-password") {
		t.Fatal("wrong password was accepted")
	}

	login := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewBufferString(url.Values{
		"password": {"secure-passphrase"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.Handler().ServeHTTP(login, request)
	if login.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, want redirect", login.Code)
	}
	var cookie *http.Cookie
	for _, candidate := range login.Result().Cookies() {
		if candidate.Name == adminCookieName {
			cookie = candidate
		}
	}
	if cookie == nil || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("admin login did not set a protected session cookie")
	}

	adminPage := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.AddCookie(cookie)
	server.Handler().ServeHTTP(adminPage, request)
	if adminPage.Code != http.StatusOK ||
		!bytes.Contains(adminPage.Body.Bytes(), []byte(`role="tab"`)) ||
		!bytes.Contains(adminPage.Body.Bytes(), []byte(`id="admin-battles"`)) ||
		!bytes.Contains(adminPage.Body.Bytes(), []byte(`id="admin-matchmaking"`)) ||
		!bytes.Contains(adminPage.Body.Bytes(), []byte(`data-admin-panel="battles" hidden`)) {
		t.Fatalf("admin page did not render its grouped sections: %d: %s", adminPage.Code, adminPage.Body.String())
	}
	appScript := httptest.NewRecorder()
	server.Handler().ServeHTTP(appScript, httptest.NewRequest(http.MethodGet, "/static/app.js", nil))
	if appScript.Code != http.StatusOK || !bytes.Contains(appScript.Body.Bytes(), []byte(`document.querySelector("[data-admin-tab]")`)) {
		t.Fatal("admin tab initialization must be based on the rendered tab controls")
	}

	stateResponse := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/admin/api/state", nil)
	request.AddCookie(cookie)
	server.Handler().ServeHTTP(stateResponse, request)
	if stateResponse.Code != http.StatusOK {
		t.Fatalf("admin state status = %d: %s", stateResponse.Code, stateResponse.Body.String())
	}
	var state struct {
		CSRF string `json:"csrf"`
	}
	if err = json.Unmarshal(stateResponse.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.CSRF == "" {
		t.Fatal("admin state did not return a CSRF token")
	}

	change := `{"current":"secure-passphrase","new":"another-secure-passphrase"}`
	withoutCSRF := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/admin/api/password", bytes.NewBufferString(change))
	request.AddCookie(cookie)
	server.Handler().ServeHTTP(withoutCSRF, request)
	if withoutCSRF.Code != http.StatusForbidden {
		t.Fatalf("password change without CSRF status = %d, want 403", withoutCSRF.Code)
	}

	changed := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/admin/api/password", bytes.NewBufferString(change))
	request.AddCookie(cookie)
	request.Header.Set("X-CSRF-Token", state.CSRF)
	server.Handler().ServeHTTP(changed, request)
	if changed.Code != http.StatusOK {
		t.Fatalf("password change status = %d: %s", changed.Code, changed.Body.String())
	}
	if accepted("secure-passphrase") || !accepted("another-secure-passphrase") {
		t.Fatal("password change did not replace the previous password")
	}

	configUpdate := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/admin/api/config", bytes.NewBufferString(
		`{"changes":{"Web.Enabled":false,"Web.SnapshotSeconds":3}}`,
	))
	request.AddCookie(cookie)
	request.Header.Set("X-CSRF-Token", state.CSRF)
	server.Handler().ServeHTTP(configUpdate, request)
	if configUpdate.Code != http.StatusOK {
		t.Fatalf("config update status = %d: %s", configUpdate.Code, configUpdate.Body.String())
	}
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(configContent, []byte("Enabled = false")) || !bytes.Contains(configContent, []byte("SnapshotSeconds = 3")) {
		t.Fatalf("config updates were not written: %s", configContent)
	}
}

func TestBattleServerHealthChecksBothTCPPorts(t *testing.T) {
	battleListener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer battleListener.Close()
	webSocketListener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer webSocketListener.Close()

	battlePort := battleListener.Addr().(*net.TCPAddr).Port
	webSocketPort := webSocketListener.Addr().(*net.TCPAddr).Port
	server := &Server{battleHealth: make(map[string]battleServerHealth)}
	active, status, _ := server.battleServerHealth("age2", "test", "127.0.0.1", battlePort, webSocketPort)
	if active || status != "Checking" {
		t.Fatalf("initial health = (%t, %q), want checking", active, status)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		active, status, _ = server.battleServerHealth("age2", "test", "127.0.0.1", battlePort, webSocketPort)
		if status != "Checking" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !active || status != "Online" {
		t.Fatalf("health = (%t, %q), want online when both TCP ports accept connections", active, status)
	}
}

func TestDashboardHidesSyntheticProfileLevel(t *testing.T) {
	if got := dashboardLevel(9_999); got != 0 {
		t.Fatalf("synthetic level = %d, want unknown (0)", got)
	}
	if got := dashboardLevel(42); got != 42 {
		t.Fatalf("real level = %d, want 42", got)
	}
}

func TestPublicBattleServerEndpointIncludesLoadedSourceAndPorts(t *testing.T) {
	dir := t.TempDir()
	store, err := models.InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err = models.SaveStaticBattleServer(models.StaticBattleServer{
		GameID: "age2", Region: "remote", Name: "Remote realm", IPv4: "192.168.10.25",
		BsPort: 30301, WebSocketPort: 30302, OutOfBandPort: 30303,
	}); err != nil {
		t.Fatal(err)
	}

	battleServers := &models.MainBattleServers{}
	battleServers.Initialize([]models.BattleServer{
		&models.MainBattleServer{Base: battleServer.Base{
			Region: "remote", Name: "Remote realm", IPv4: "192.168.10.25",
			BsPort: 30301, WebSocketPort: 30302, OutOfBandPort: 30303,
		}},
	}, nil)
	server, err := New(Options{
		Config:  internal.Web{},
		DataDir: dir,
		Games:   map[string]models.Game{"age2": emptyGame{battleServers: battleServers}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/battle-servers?game=age2", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/battle-servers = %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		BattleServers []BattleServerView `json:"battleServers"`
	}
	if err = json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.BattleServers) != 1 {
		t.Fatalf("battle servers = %d, want 1", len(payload.BattleServers))
	}
	entry := payload.BattleServers[0]
	if entry.Region != "remote" || entry.Name != "Remote realm" || entry.Source != "Static" ||
		entry.IPv4 != "192.168.10.25" || entry.BsPort != 30301 ||
		entry.WebSocketPort != 30302 || entry.OutOfBandPort != 30303 {
		t.Fatalf("public battle-server details = %+v", entry)
	}

	page := httptest.NewRecorder()
	server.Handler().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/battle-servers?game=age2", nil))
	if page.Code != http.StatusOK || !bytes.Contains(page.Body.Bytes(), []byte("Remote realm")) {
		t.Fatalf("battle-server page did not render loaded servers: %d: %s", page.Code, page.Body.String())
	}
}

func TestAdminAPIRejectsMissingSession(t *testing.T) {
	server, err := New(Options{Config: internal.Web{}, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })

	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin/api/state", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestPlayerProfileShowsCivilizationsMapsAndRecentResults(t *testing.T) {
	dir := t.TempDir()
	history := models.NewMatchHistoryStore(filepath.Join(dir, "players"))
	if _, _, err := history.GetProfilePerformance(101); err != nil {
		t.Fatalf("profile history database was not created: %v", err)
	}
	t.Cleanup(func() { _ = history.Close() })
	history.UpsertProfile(models.StoredProfile{ProfileID: 101, Alias: "TestPlayer"})
	history.ApplyResult(models.MatchResult{
		MatchID: 1001, MapName: "Arabia.rms", Players: []int32{101, 202},
		Teams: []int32{0, 1}, Results: []int32{1, 0}, Races: []int32{2, 5},
		StartedAt: time.Now().Add(-time.Hour), FinishedAt: time.Now(),
	})
	server, err := New(Options{
		Config:  internal.Web{PublicTitle: "Community"},
		DataDir: filepath.Join(dir, "web"),
		Games:   map[string]models.Game{"age2": emptyGame{history: history}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })

	for _, path := range []string{"/players/101?game=age2", "/api/players/101?game=age2"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s = %d: %s", path, response.Code, response.Body.String())
		}
		expectedValues := []string{"TestPlayer", "Arabia", "Britons"}
		if path == "/players/101?game=age2" {
			expectedValues = append(expectedValues, "Winner")
		}
		for _, expected := range expectedValues {
			if !bytes.Contains(response.Body.Bytes(), []byte(expected)) {
				t.Errorf("GET %s did not include %q: %s", path, expected, response.Body.String())
			}
		}
	}
}

func TestLocalAddressesUseDashboardScheme(t *testing.T) {
	addresses := localAddresses([]string{"127.0.0.1"}, 443)
	if len(addresses) != 1 || addresses[0] != "https://127.0.0.1:443" {
		t.Fatalf("addresses = %v, want only the configured HTTPS address", addresses)
	}
}

func TestDashboardSharesHTTPSHandlerWithGameRoutes(t *testing.T) {
	server, err := New(Options{
		Config:    internal.Web{},
		DataDir:   t.TempDir(),
		GameHosts: map[string][]string{"age2": {"127.0.0.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	gameHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("game route"))
	})
	tlsServer := httptest.NewTLSServer(server.HandlerWithFallback(gameHandler))
	defer tlsServer.Close()

	response, err := tlsServer.Client().Get(tlsServer.URL + "/api/health")
	if err != nil {
		t.Fatalf("dashboard health request failed: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d, want 200", response.StatusCode)
	}

	response, err = tlsServer.Client().Get(tlsServer.URL + "/api/game-route")
	if err != nil {
		t.Fatalf("game route request failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("game route status = %d, want 202", response.StatusCode)
	}
	if len(server.Addresses()) != 1 || server.Addresses()[0] != "https://127.0.0.1:443" {
		t.Fatalf("dashboard advertised URLs = %v, want HTTPS on port 443", server.Addresses())
	}
}

func TestAdminStaticBattleServerCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := models.InitializeStaticBattleServerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	server, err := New(Options{
		Config:  internal.Web{},
		DataDir: dir,
		Games:   map[string]models.Game{"age2": emptyGame{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	hash, err := hashPassword("secure-passphrase")
	if err != nil {
		t.Fatal(err)
	}
	if err = server.store.Set("admin_password_hash", hash); err != nil {
		t.Fatal(err)
	}
	session, err := server.sessions.create("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	body := `{"game":"age2","region":"static-east","name":"East","ipv4":"192.168.10.25","bsPort":30301,"webSocketPort":30302,"outOfBandPort":30303}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/admin/api/battle-servers", bytes.NewBufferString(body))
	request.AddCookie(&http.Cookie{Name: adminCookieName, Value: session.token})
	request.Header.Set("X-CSRF-Token", session.csrf)
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("create static server = %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/admin/api/battle-servers?game=age2", nil)
	request.AddCookie(&http.Cookie{Name: adminCookieName, Value: session.token})
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte("static-east")) {
		t.Fatalf("list static servers = %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodDelete, "/admin/api/battle-servers/age2/static-east", nil)
	request.AddCookie(&http.Cookie{Name: adminCookieName, Value: session.token})
	request.Header.Set("X-CSRF-Token", session.csrf)
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("delete static server = %d: %s", response.Code, response.Body.String())
	}
}

type emptyGame struct {
	battleServers models.BattleServers
	history       models.MatchHistory
}

func (emptyGame) Title() string                                         { return "age2" }
func (emptyGame) Resources() models.Resources                           { return nil }
func (emptyGame) Items() models.Items                                   { return nil }
func (emptyGame) LeaderboardDefinitions() models.LeaderboardDefinitions { return nil }
func (emptyGame) PresenceDefinitions() models.PresenceDefinitions       { return nil }
func (g emptyGame) BattleServers() models.BattleServers                 { return g.battleServers }
func (emptyGame) Users() models.Users                                   { return nil }
func (emptyGame) Advertisements() models.Advertisements                 { return nil }
func (emptyGame) ChatChannels() models.ChatChannels                     { return nil }
func (emptyGame) Sessions() models.Sessions                             { return nil }
func (emptyGame) CustomFiles() models.CustomFiles                       { return nil }
func (emptyGame) Matchmaking() models.Matchmaking                       { return nil }
func (g emptyGame) MatchHistory() models.MatchHistory                   { return g.history }
