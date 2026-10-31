package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/events"
	"github.com/luskaner/ageServer/server/internal/logger"
	"github.com/luskaner/ageServer/server/internal/models"
)

//go:embed templates/*.html static/*
var assets embed.FS

// Options configures the dashboard.
type Options struct {
	Config     internal.Web
	ConfigFile string
	Games      map[string]models.Game
	GameHosts  map[string][]string
	Version    string
	DataDir    string
}

// Server holds the dashboard handlers and its persistent administration store.
type Server struct {
	opts                Options
	store               *AdminStore
	sessions            *sessionManager
	startedAt           time.Time
	pages               map[string]*template.Template
	fragments           *template.Template
	admin               *template.Template
	addresses           []string
	staticBattleServers map[string]map[string]struct{}
	battleHealthMu      sync.Mutex
	battleHealth        map[string]battleServerHealth
}

// New builds the dashboard handlers. The game HTTPS listeners serve them on port 443.
// Admin and static battle-server database errors fail startup; an unreachable config
// file is reported in the panel, not at boot.
func New(opts Options) (*Server, error) {
	if opts.Config.SnapshotSeconds <= 0 {
		opts.Config.SnapshotSeconds = 2
	}
	if opts.Config.PublicTitle == "" {
		opts.Config.PublicTitle = "Community Server"
	}
	store, err := NewAdminStore(opts.DataDir)
	if err != nil {
		return nil, err
	}
	if title, ok, titleErr := store.Get("public_title"); titleErr != nil {
		_ = store.Close()
		return nil, fmt.Errorf("load public site title: %w", titleErr)
	} else if ok && strings.TrimSpace(title) != "" {
		opts.Config.PublicTitle = title
	}
	server := &Server{
		opts:                opts,
		store:               store,
		sessions:            newSessionManager(),
		startedAt:           time.Now(),
		staticBattleServers: make(map[string]map[string]struct{}),
		battleHealth:        make(map[string]battleServerHealth),
	}
	for gameID, game := range opts.Games {
		if game == nil || game.BattleServers() == nil {
			continue
		}
		records, loadErr := models.StaticBattleServers(gameID)
		if loadErr != nil {
			_ = store.Close()
			return nil, fmt.Errorf("load static battle-server sources for %s: %w", gameID, loadErr)
		}
		server.staticBattleServers[gameID] = make(map[string]struct{}, len(records))
		for _, record := range records {
			server.staticBattleServers[gameID][record.Region] = struct{}{}
		}
	}
	if err = server.parseTemplates(); err != nil {
		_ = store.Close()
		return nil, err
	}
	server.addresses = localAddresses(flattenHosts(opts.GameHosts), 443)
	return server, nil
}

// Store exposes the admin database (used by tests).
func (s *Server) Store() *AdminStore { return s.store }

// Addresses returns the URLs the dashboard answers on.
func (s *Server) Addresses() []string { return s.addresses }

// Close allows the dashboard to be included with other server resources.
func (s *Server) Close() error {
	return s.Shutdown(context.Background())
}

func (s *Server) parseTemplates() error {
	funcs := template.FuncMap{
		"time": func(value time.Time) string {
			if value.IsZero() {
				return "-"
			}
			return value.Local().Format("2006-01-02 15:04")
		},
		"unixTime": func(value int64) string {
			if value <= 0 {
				return "-"
			}
			return time.Unix(value, 0).Local().Format("15:04")
		},
		"duration": func(seconds int) string {
			if seconds < 60 {
				return fmt.Sprintf("%ds", seconds)
			}
			minutes := seconds / 60
			if minutes < 60 {
				return fmt.Sprintf("%dm %ds", minutes, seconds%60)
			}
			return fmt.Sprintf("%dh %dm", minutes/60, minutes%60)
		},
		"join": func(values []string) string { return strings.Join(values, ", ") },
		"initial": func(value string) string {
			if value == "" {
				return "?"
			}
			return strings.ToUpper(value[:1])
		},
		"percent": func(part, total int) int {
			if total <= 0 {
				return 0
			}
			return part * 100 / total
		},
		"level": func(value uint16) string {
			if value == 0 || value == 9_999 {
				return "—"
			}
			return strconv.FormatUint(uint64(value), 10)
		},
		"activeBattleServers": func(servers []BattleServerView) int {
			count := 0
			for _, server := range servers {
				if server.Active {
					count++
				}
			}
			return count
		},
		"staticBattleServers": func(servers []BattleServerView) int {
			count := 0
			for _, server := range servers {
				if server.Source == "Static" {
					count++
				}
			}
			return count
		},
	}
	s.pages = make(map[string]*template.Template)
	for _, page := range []string{"index", "lobbies", "queue", "matches", "leaderboard", "players", "profile", "maps", "battle-servers", "lobby"} {
		parsed, err := template.New("layout").Funcs(funcs).ParseFS(assets,
			"templates/layout.html", "templates/fragments.html", "templates/"+page+".html")
		if err != nil {
			return fmt.Errorf("parse page %s: %w", page, err)
		}
		s.pages[page] = parsed
	}
	fragments, err := template.New("fragments").Funcs(funcs).ParseFS(assets, "templates/fragments.html")
	if err != nil {
		return fmt.Errorf("parse fragments: %w", err)
	}
	s.fragments = fragments
	admin, err := template.New("admin").Funcs(funcs).ParseFS(assets,
		"templates/layout.html", "templates/admin.html")
	if err != nil {
		return fmt.Errorf("parse admin page: %w", err)
	}
	s.admin = admin
	return nil
}

// Handler returns the dashboard router. It is exported for tests.
func (s *Server) Handler() http.Handler {
	return s.HandlerWithFallback(http.NotFoundHandler())
}

// HandlerWithFallback serves dashboard paths and delegates all other paths to the
// existing game handler, allowing both services to share the same HTTPS listener.
func (s *Server) HandlerWithFallback(fallback http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Public API.
	mux.HandleFunc("GET /api/info", s.HandleInfo)
	mux.HandleFunc("GET /api/health", s.HandleHealth)
	mux.HandleFunc("GET /api/lobbies", s.HandleLobbies)
	mux.HandleFunc("GET /api/lobbies/{id}", s.HandleLobby)
	mux.HandleFunc("GET /api/queue", s.HandleQueue)
	mux.HandleFunc("GET /api/matches", s.HandleMatches)
	mux.HandleFunc("GET /api/matches/{id}", s.HandleMatch)
	mux.HandleFunc("GET /api/leaderboard", s.HandleLeaderboard)
	mux.HandleFunc("GET /api/players", s.HandlePlayers)
	mux.HandleFunc("GET /api/players/{id}", s.HandlePlayerProfile)
	mux.HandleFunc("GET /api/maps", s.HandleMaps)
	mux.HandleFunc("GET /api/battle-servers", s.HandleBattleServers)
	mux.HandleFunc("GET /api/events", s.HandleEvents)

	// Server-rendered pages.
	mux.HandleFunc("GET /{$}", s.handlePage("index"))
	mux.HandleFunc("GET /lobbies", s.handlePage("lobbies"))
	mux.HandleFunc("GET /queue", s.handlePage("queue"))
	mux.HandleFunc("GET /matches", s.handlePage("matches"))
	mux.HandleFunc("GET /leaderboard", s.handlePage("leaderboard"))
	mux.HandleFunc("GET /players", s.handlePage("players"))
	mux.HandleFunc("GET /players/{id}", s.handlePlayerProfilePage)
	mux.HandleFunc("GET /maps", s.handlePage("maps"))
	mux.HandleFunc("GET /battle-servers", s.handlePage("battle-servers"))
	mux.HandleFunc("GET /lobby/{id}", s.handleLobbyPage)

	// HTML fragments used by the SSE-driven live updates.
	mux.HandleFunc("GET /fragments/lobbies", s.handleFragment("lobbies"))
	mux.HandleFunc("GET /fragments/matches", s.handleFragment("matches"))
	mux.HandleFunc("GET /fragments/queue", s.handleFragment("queue"))
	mux.HandleFunc("GET /fragments/leaderboard", s.handleFragment("leaderboard"))
	mux.HandleFunc("GET /fragments/players", s.handleFragment("players"))
	mux.HandleFunc("GET /fragments/maps", s.handleFragment("maps"))

	// Static assets.
	static, _ := fs.Sub(assets, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	// Admin.
	mux.HandleFunc("GET /admin/login", s.handleLoginPage)
	mux.HandleFunc("POST /admin/login", s.handleLogin)
	mux.HandleFunc("GET /admin/logout", s.handleLogout)
	mux.HandleFunc("POST /admin/logout", s.handleLogout)
	mux.HandleFunc("GET /admin/setup", s.handleSetupPage)
	mux.HandleFunc("POST /admin/setup", s.handleSetup)
	mux.HandleFunc("GET /admin", s.handleAdminPage)
	mux.HandleFunc("GET /admin/api/state", s.HandleAdminState)
	mux.HandleFunc("GET /admin/api/pool", s.HandleAdminPool)
	mux.HandleFunc("PUT /admin/api/pool", s.HandleAdminPool)
	mux.HandleFunc("POST /admin/api/pool/reset", s.HandleAdminPoolReset)
	mux.HandleFunc("POST /admin/api/pool/cursor", s.HandleAdminPoolCursor)
	mux.HandleFunc("GET /admin/api/config", s.HandleAdminConfig)
	mux.HandleFunc("POST /admin/api/config", s.HandleAdminConfig)
	mux.HandleFunc("POST /admin/api/config/raw", s.HandleAdminConfigRaw)
	mux.HandleFunc("POST /admin/api/password", s.HandleAdminPassword)
	mux.HandleFunc("GET /admin/api/battle-servers", s.HandleAdminBattleServers)
	mux.HandleFunc("POST /admin/api/battle-servers", s.HandleAdminBattleServers)
	mux.HandleFunc("DELETE /admin/api/battle-servers/{game}/{region}", s.HandleAdminDeleteBattleServer)
	mux.HandleFunc("GET /admin/api/matches", s.HandleAdminRecentMatches)
	mux.HandleFunc("POST /admin/api/lobbies/{id}/close", s.HandleAdminCloseLobby)
	mux.HandleFunc("POST /admin/api/queue/{user}/kick", s.HandleAdminKick)
	mux.Handle("/", fallback)

	return loggingMiddleware(mux)
}

// loggingMiddleware logs dashboard requests under the dashboard prefix.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isDashboardPath(r.URL.Path) || r.URL.Path == "/static/style.css" || r.URL.Path == "/static/app.js" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/api/events" {
			logger.Printf("web: %s %s (%s)\n", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}
	})
}

func isDashboardPath(path string) bool {
	if path == "/" || path == "/lobbies" || path == "/queue" || path == "/matches" ||
		path == "/leaderboard" || path == "/players" || path == "/maps" || path == "/battle-servers" ||
		strings.HasPrefix(path, "/players/") || strings.HasPrefix(path, "/lobby/") || strings.HasPrefix(path, "/fragments/") ||
		strings.HasPrefix(path, "/static/") || path == "/admin" || strings.HasPrefix(path, "/admin/") {
		return true
	}
	for _, prefix := range []string{
		"/api/info", "/api/health", "/api/lobbies", "/api/queue", "/api/matches",
		"/api/leaderboard", "/api/players", "/api/maps", "/api/events",
		"/api/battle-servers",
	} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

// Shutdown closes the dashboard's persistent store.
func (s *Server) Shutdown(_ context.Context) error {
	if s.store != nil {
		return s.store.Close()
	}
	return nil
}

// localAddresses lists the URLs a LAN client can use, so the console shows
// every reachable IP instead of just the bound one.
func localAddresses(hosts []string, port int) []string {
	seen := make(map[string]struct{})
	var addresses []string
	add := func(host string) {
		if host == "" || host == "0.0.0.0" || host == "::" {
			return
		}
		address := fmt.Sprintf("https://%s:%d", host, port)
		if _, ok := seen[address]; ok {
			return
		}
		seen[address] = struct{}{}
		addresses = append(addresses, address)
	}
	wildcard := len(hosts) == 0
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" || host == "0.0.0.0" {
			wildcard = true
			continue
		}
		add(host)
	}
	if wildcard {
		interfaces, err := net.Interfaces()
		if err == nil {
			for _, iface := range interfaces {
				if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
					continue
				}
				addrs, addrErr := iface.Addrs()
				if addrErr != nil {
					continue
				}
				for _, addr := range addrs {
					ipNet, ok := addr.(*net.IPNet)
					if !ok || ipNet.IP.To4() == nil {
						continue
					}
					add(ipNet.IP.String())
				}
			}
		}
		add("127.0.0.1")
	}
	sort.Strings(addresses)
	return addresses
}

func flattenHosts(hostsByGame map[string][]string) []string {
	var hosts []string
	for _, hostsForGame := range hostsByGame {
		hosts = append(hosts, hostsForGame...)
	}
	return hosts
}

// lastEventID is the id of the newest event on the bus.
func (s *Server) lastEventID() uint64 { return events.LastID() }

// pageData is what every rendered page receives.
type pageData struct {
	Title         string
	Path          string
	Config        internal.Web
	Snapshot      Snapshot
	Game          GameSnapshot
	LadderPage    []LeaderboardRow
	LadderTot     int64
	GameID        string
	Games         []string
	Lobby         *models.PublicLobby
	PlayerProfile *PlayerProfileView
	Recent        []MatchView
	Error         string
	SetupToken    string
	SiteTitle     string
	Admin         bool
}

func (s *Server) basePage(r *http.Request, title, path string) pageData {
	snapshot := s.BuildSnapshot()
	page := pageData{
		Title:     title,
		Path:      path,
		Config:    s.opts.Config,
		Snapshot:  snapshot,
		Games:     s.sortedGames(),
		SiteTitle: s.opts.Config.PublicTitle,
	}
	if len(snapshot.Games) > 0 {
		page.Game = snapshot.Games[0]
		page.LadderPage = page.Game.Ladder
		page.LadderTot = page.Game.LadderTotal
		page.GameID = page.Game.Game
	}
	if requested := r.URL.Query().Get("game"); requested != "" {
		for _, game := range snapshot.Games {
			if game.Game == requested {
				page.Game = game
				page.GameID = game.Game
				break
			}
		}
	}
	if page.GameID == "" && len(page.Games) > 0 {
		page.GameID = page.Games[0]
	}
	if game := s.opts.Games[page.GameID]; game != nil && game.MatchHistory() != nil {
		page.Recent = s.enrichMatches(game, game.MatchHistory().RecentMatches(20))
	}
	return page
}

func (s *Server) handlePage(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusMovedPermanently)
			return
		}
		page := s.basePage(r, pageTitle(name), "/"+name)
		if name == "leaderboard" {
			page.LadderPage, page.LadderTot = leaderboardRowsOrEmpty(s.opts.Games[page.GameID], page.Game)
		}
		s.executePage(w, name, page)
	}
}

func leaderboardRowsOrEmpty(game models.Game, snapshot GameSnapshot) ([]LeaderboardRow, int64) {
	if game == nil || game.MatchHistory() == nil {
		return nil, 0
	}
	return leaderboardRows(game, snapshot.LadderID, 0, 100)
}

func (s *Server) handleLobbyPage(w http.ResponseWriter, r *http.Request) {
	page := s.basePage(r, "Lobby", "/lobbies")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	requestedGame := strings.TrimSpace(r.URL.Query().Get("game"))
	for _, game := range page.Snapshot.Games {
		if requestedGame != "" && game.Game != requestedGame {
			continue
		}
		for index := range game.Lobbies {
			if game.Lobbies[index].ID == int32(id) {
				page.Lobby = &game.Lobbies[index]
				page.Game = game
				page.GameID = game.Game
				break
			}
		}
		if page.Lobby != nil {
			break
		}
	}
	if page.Lobby == nil {
		http.NotFound(w, r)
		return
	}
	s.executePage(w, "lobby", page)
}

func (s *Server) executePage(w http.ResponseWriter, name string, data pageData) {
	parsed, ok := s.pages[name]
	if !ok {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := parsed.ExecuteTemplate(w, "layout", data); err != nil {
		logger.Printf("web: render %s: %s\n", name, err)
	}
}

func (s *Server) handleFragment(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := s.basePage(r, name, "/"+name)
		if name == "leaderboard" {
			page.LadderPage, page.LadderTot = leaderboardRowsOrEmpty(s.opts.Games[page.GameID], page.Game)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if err := s.fragments.ExecuteTemplate(w, "frag-"+name, page); err != nil {
			logger.Printf("web: render fragment %s: %s\n", name, err)
		}
	}
}

func (s *Server) handleAdminPage(w http.ResponseWriter, r *http.Request) {
	configured, err := s.passwordConfigured()
	if err != nil {
		http.Error(w, "could not read admin configuration", http.StatusInternalServerError)
		return
	}
	if !configured {
		http.Redirect(w, r, "/admin/setup", http.StatusSeeOther)
		return
	}
	if _, ok := s.currentSession(r); !ok {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}
	page := s.basePage(r, "Admin", "/admin")
	page.Admin = true
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.admin.ExecuteTemplate(w, "layout", page); err != nil {
		logger.Printf("web: render admin: %s\n", err)
	}
}

// render is used by the login/setup pages, which have no snapshot to build.
func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	parsed, err := template.New(name).Funcs(template.FuncMap{
		"join": strings.Join,
	}).ParseFS(assets, "templates/layout.html", "templates/"+name+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := pageData{
		Title:  fmt.Sprint(data["Title"]),
		Config: s.opts.Config,
		Path:   "/" + name,
	}
	if message, ok := data["Error"].(string); ok {
		page.Error = message
	}
	if token, ok := data["SetupToken"].(string); ok {
		page.SetupToken = token
	}
	page.SiteTitle = s.opts.Config.PublicTitle
	if title, ok := data["SiteTitle"].(string); ok {
		page.SiteTitle = title
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = parsed.ExecuteTemplate(w, "layout", page); err != nil {
		logger.Printf("web: render %s: %s\n", name, err)
	}
}

func pageTitle(name string) string {
	switch name {
	case "index":
		return "Live"
	case "lobbies":
		return "Lobbies"
	case "queue":
		return "Automatch queue"
	case "matches":
		return "Matches"
	case "leaderboard":
		return "Ladder"
	case "players":
		return "Players"
	case "maps":
		return "Maps"
	case "battle-servers":
		return "Battle servers"
	default:
		if name == "" {
			return ""
		}
		return strings.ToUpper(name[:1]) + name[1:]
	}
}
