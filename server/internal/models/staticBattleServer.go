package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/luskaner/ageServer/common/battleServer"
	"github.com/luskaner/ageServer/common/game"
	_ "modernc.org/sqlite"
)

type StaticBattleServer struct {
	GameID        string `json:"game"`
	Region        string `json:"region"`
	Name          string `json:"name"`
	IPv4          string `json:"ipv4"`
	BsPort        int    `json:"bsPort"`
	WebSocketPort int    `json:"webSocketPort"`
	OutOfBandPort int    `json:"outOfBandPort"`
}

type StaticBattleServerStore struct {
	mu sync.Mutex
	db *sql.DB
}

var staticBattleServers *StaticBattleServerStore
var staticBattleServersMu sync.RWMutex

func InitializeStaticBattleServerStore(dataDir string) (*StaticBattleServerStore, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create static battle-server data directory: %w", err)
	}
	path := filepath.Join(dataDir, "web.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open static battle-server database: %w", err)
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=NORMAL", "PRAGMA busy_timeout=5000"} {
		if _, err = db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%s: %w", pragma, err)
		}
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS static_battle_servers (
		game_id          TEXT    NOT NULL,
		region           TEXT    NOT NULL,
		name             TEXT    NOT NULL DEFAULT '',
		ipv4             TEXT    NOT NULL,
		bs_port          INTEGER NOT NULL,
		websocket_port   INTEGER NOT NULL,
		out_of_band_port INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (game_id, region)
	)`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create static battle-server table: %w", err)
	}
	store := &StaticBattleServerStore{db: db}
	staticBattleServersMu.Lock()
	previous := staticBattleServers
	staticBattleServers = store
	staticBattleServersMu.Unlock()
	if previous != nil {
		_ = previous.Close()
	}
	return store, nil
}

func (s *StaticBattleServerStore) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func StaticBattleServers(gameID string) ([]StaticBattleServer, error) {
	store := currentStaticBattleServerStore()
	if store == nil {
		return nil, errors.New("static battle-server database is not initialized")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.db == nil {
		return nil, errors.New("static battle-server database is closed")
	}
	rows, err := store.db.Query(`SELECT game_id, region, name, ipv4, bs_port, websocket_port, out_of_band_port
		FROM static_battle_servers WHERE game_id = ? ORDER BY region`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var servers []StaticBattleServer
	for rows.Next() {
		var server StaticBattleServer
		if err = rows.Scan(&server.GameID, &server.Region, &server.Name, &server.IPv4,
			&server.BsPort, &server.WebSocketPort, &server.OutOfBandPort); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, rows.Err()
}

func SaveStaticBattleServer(server StaticBattleServer) error {
	if err := validateStaticBattleServer(&server); err != nil {
		return err
	}
	store := currentStaticBattleServerStore()
	if store == nil {
		return errors.New("static battle-server database is not initialized")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.db == nil {
		return errors.New("static battle-server database is closed")
	}
	_, err := store.db.Exec(`INSERT INTO static_battle_servers
		(game_id, region, name, ipv4, bs_port, websocket_port, out_of_band_port)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(game_id, region) DO UPDATE SET
			name = excluded.name, ipv4 = excluded.ipv4, bs_port = excluded.bs_port,
			websocket_port = excluded.websocket_port, out_of_band_port = excluded.out_of_band_port`,
		server.GameID, server.Region, server.Name, server.IPv4, server.BsPort,
		server.WebSocketPort, server.OutOfBandPort)
	return err
}

func DeleteStaticBattleServer(gameID, region string) error {
	store := currentStaticBattleServerStore()
	if store == nil {
		return errors.New("static battle-server database is not initialized")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.db == nil {
		return errors.New("static battle-server database is closed")
	}
	result, err := store.db.Exec(`DELETE FROM static_battle_servers WHERE game_id = ? AND region = ?`, gameID, region)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func LoadStaticBattleServers(gameID string) ([]BattleServer, error) {
	records, err := StaticBattleServers(gameID)
	if err != nil {
		return nil, err
	}
	servers := make([]BattleServer, 0, len(records))
	for _, record := range records {
		servers = append(servers, &MainBattleServer{Base: battleServer.Base{
			Region: record.Region, Name: record.Name, IPv4: record.IPv4, BsPort: record.BsPort,
			WebSocketPort: record.WebSocketPort, OutOfBandPort: record.OutOfBandPort,
		}})
	}
	return servers, nil
}

func validateStaticBattleServer(server *StaticBattleServer) error {
	server.GameID = strings.TrimSpace(server.GameID)
	server.Region = strings.TrimSpace(server.Region)
	server.Name = strings.TrimSpace(server.Name)
	server.IPv4 = strings.TrimSpace(server.IPv4)
	if !game.SupportedGames.ContainsOne(server.GameID) {
		return fmt.Errorf("unsupported game %q", server.GameID)
	}
	if server.Region == "" {
		return errors.New("region is required")
	}
	if server.Name == "" {
		server.Name = server.Region
	}
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return errors.New("a valid IPv4 address is required")
	}
	if server.BsPort < 1 || server.BsPort > 65535 {
		return errors.New("battle server port must be between 1 and 65535")
	}
	if server.WebSocketPort < 1 || server.WebSocketPort > 65535 {
		return errors.New("websocket port must be between 1 and 65535")
	}
	if server.OutOfBandPort < 0 || server.OutOfBandPort > 65535 {
		return errors.New("out-of-band port must be between 0 and 65535")
	}
	ports := []int{server.BsPort, server.WebSocketPort}
	if server.OutOfBandPort != 0 {
		ports = append(ports, server.OutOfBandPort)
	}
	sort.Ints(ports)
	for index := 1; index < len(ports); index++ {
		if ports[index] == ports[index-1] {
			return errors.New("battle server ports must be different")
		}
	}
	return nil
}

func currentStaticBattleServerStore() *StaticBattleServerStore {
	staticBattleServersMu.RLock()
	defer staticBattleServersMu.RUnlock()
	return staticBattleServers
}
