package models

import (
	"sort"
	"time"
)

// PublicPeer is one player inside a lobby, as shown on the public site. The
// whole point of the LAN dashboard is that nothing is hidden: aliases, profile
// ids and the IPs the server sees are all public.
type PublicPeer struct {
	UserID      int32  `json:"userId"`
	ProfileID   int32  `json:"profileId"`
	StatGroupID int32  `json:"statGroupId"`
	Alias       string `json:"alias"`
	IP          string `json:"ip,omitempty"`
	Party       int32  `json:"party"`
	TeamID      int32  `json:"teamId"`
	RaceID      int32  `json:"raceId"`
	RaceName    string `json:"raceName"`
	RaceIcon    string `json:"raceIcon"`
}

// PublicLobby is a snapshot of one advertisement.
type PublicLobby struct {
	ID                int32        `json:"id"`
	Game              string       `json:"game"`
	IP                string       `json:"ip"`
	HostID            int32        `json:"hostId"`
	HostProfileID     int32        `json:"hostProfileId"`
	HostAlias         string       `json:"hostAlias"`
	MapName           string       `json:"mapName"`
	DisplayMapName    string       `json:"displayMapName"`
	MapIcon           string       `json:"mapIcon"`
	Description       string       `json:"description"`
	MaxPlayers        int          `json:"maxPlayers"`
	PlayerCount       int          `json:"playerCount"`
	Visible           bool         `json:"visible"`
	Joinable          bool         `json:"joinable"`
	Passworded        bool         `json:"passworded"`
	Automatch         bool         `json:"automatch"`
	ObserversEnabled  bool         `json:"observersEnabled"`
	ObserverDelay     uint32       `json:"observerDelay"`
	Lan               bool         `json:"lan"`
	State             int8         `json:"state"`
	MatchType         uint8        `json:"matchType"`
	RelayRegion       string       `json:"relayRegion"`
	Party             int32        `json:"party"`
	PlatformSessionID uint64       `json:"platformSessionId,omitempty"`
	XboxSessionID     string       `json:"xboxSessionId,omitempty"`
	StartedAt         int64        `json:"startedAt"`
	Peers             []PublicPeer `json:"players"`
}

// QueuePlayer is one player waiting in the automatch queue.
type QueuePlayer struct {
	UserID         int32     `json:"userId"`
	ProfileID      int32     `json:"profileId"`
	Alias          string    `json:"alias"`
	PollID         int32     `json:"pollId"`
	IP             string    `json:"ip,omitempty"`
	MatchTypes     []int32   `json:"matchTypes,omitempty"`
	VetoMapIDs     []int32   `json:"vetoMapIds,omitempty"`
	VetoMaps       []string  `json:"vetoMaps,omitempty"`
	PartySessionID int64     `json:"party,omitempty"`
	RelayRegion    string    `json:"relayRegion,omitempty"`
	MatchID        int32     `json:"matchId,omitempty"`
	WaitingSince   time.Time `json:"waitingSince"`
	WaitingSeconds int       `json:"waitingSeconds"`
}

// LiveMatch is an automatch pairing in progress.
type LiveMatch struct {
	ID                int32         `json:"id"`
	QueueID           int32         `json:"queueId"`
	HostID            int32         `json:"hostId"`
	HostAlias         string        `json:"hostAlias"`
	MapName           string        `json:"mapName"`
	DisplayMapName    string        `json:"displayMapName"`
	MapIcon           string        `json:"mapIcon"`
	MapID             int32         `json:"mapId"`
	Players           []QueuePlayer `json:"players"`
	State             string        `json:"state"`
	AdvertisementID   int32         `json:"advertisementId,omitempty"`
	PlatformSessionID uint64        `json:"platformSessionId,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
}

// MatchmakingSnapshot is everything the matchmaker knows right now.
type MatchmakingSnapshot struct {
	Queue   []QueuePlayer `json:"queue"`
	Matches []LiveMatch   `json:"matches"`
}

// PublicPlayer is a connected (or recently known) player.
type PublicPlayer struct {
	UserID          int32  `json:"userId"`
	ProfileID       int32  `json:"profileId"`
	StatGroupID     int32  `json:"statGroupId"`
	Alias           string `json:"alias"`
	Xbox            bool   `json:"xbox"`
	Level           uint16 `json:"level,omitempty"`
	XP              uint32 `json:"xp"`
	Rating          int32  `json:"rating"`
	Online          bool   `json:"online"`
	IP              string `json:"ip,omitempty"`
	AdvertisementID int32  `json:"advertisementId,omitempty"`
	InQueue         bool   `json:"inQueue"`
}

// Snapshot returns every advertisement of the game. It takes the per-lobby
// read lock so the values it copies are internally consistent.
func (advs *MainAdvertisements) Snapshot(gameId string) []PublicLobby {
	if advs == nil || advs.store == nil {
		return nil
	}
	count, sequence := advs.store.Iter()
	lobbies := make([]PublicLobby, 0, count)
	for _, adv := range sequence {
		if adv == nil {
			continue
		}
		id := adv.GetId()
		advs.locks.RLock(id)
		lobby := advs.snapshotLocked(gameId, adv)
		advs.locks.RUnlock(id)
		lobbies = append(lobbies, lobby)
	}
	return lobbies
}

func (advs *MainAdvertisements) snapshotLocked(gameId string, adv Advertisement) PublicLobby {
	lobby := PublicLobby{
		ID:                adv.GetId(),
		Game:              gameId,
		IP:                adv.GetIp(),
		HostID:            adv.UnsafeGetHostId(),
		MapName:           adv.UnsafeGetMapName(),
		DisplayMapName:    DisplayMapName(adv.UnsafeGetMapName()),
		MapIcon:           MapIconPath(adv.UnsafeGetMapName()),
		Description:       adv.UnsafeGetDescription(),
		MaxPlayers:        adv.UnsafeGetMaxPlayers(),
		Visible:           adv.UnsafeGetVisible(),
		Joinable:          adv.UnsafeGetJoinable(),
		Passworded:        adv.UnsafeGetPasswordValue() != "",
		Automatch:         adv.UnsafeIsAutomatch(),
		ObserversEnabled:  adv.UnsafeGetObserversEnabled(),
		ObserverDelay:     adv.UnsafeGetObserversDelay(),
		State:             adv.UnsafeGetState(),
		MatchType:         adv.UnsafeGetMatchType(),
		RelayRegion:       adv.GetRelayRegion(),
		Party:             adv.GetParty(),
		PlatformSessionID: adv.UnsafeGetPlatformSessionId(),
		XboxSessionID:     adv.GetXboxSessionId(),
		StartedAt:         adv.UnsafeGetStartTime(),
	}

	main, ok := adv.(*MainAdvertisement)
	if !ok {
		return lobby
	}
	lobby.Lan = main.lan
	lobby.Passworded = main.password.enabled
	if peers := main.peers; peers != nil {
		count, sequence := peers.Iter()
		lobby.Peers = make([]PublicPeer, 0, count)
		for _, peer := range sequence {
			entry := PublicPeer{}
			if mainPeer, ok := peer.(*MainPeer); ok {
				entry.UserID = mainPeer.userId
				entry.StatGroupID = mainPeer.statGroupID
				entry.IP = mainPeer.advertisementIp
				if mutable := mainPeer.mutable.Load(); mutable != nil {
					entry.TeamID = mutable.TeamID
					entry.RaceID = mutable.RaceID
					entry.RaceName = CivilizationName(gameId, mutable.RaceID)
					entry.RaceIcon = CivilizationIconPath(gameId, mutable.RaceID)
				}
			} else {
				entry.UserID = peer.GetUserId()
			}
			entry.Party = peer.GetParty()
			if advs.users != nil {
				if user, found := advs.users.GetUserById(entry.UserID); found {
					entry.Alias = user.GetAlias()
					entry.ProfileID = user.GetProfileId()
				}
			}
			lobby.Peers = append(lobby.Peers, entry)
		}
	}
	lobby.PlayerCount = len(lobby.Peers)
	if advs.users != nil {
		if host, found := advs.users.GetUserById(lobby.HostID); found {
			lobby.HostAlias = host.GetAlias()
			lobby.HostProfileID = host.GetProfileId()
		}
	}
	return lobby
}

// Snapshot copies the queue and the live matches.
func (mm *MainMatchmaking) Snapshot() MatchmakingSnapshot {
	if mm == nil || mm.queue == nil {
		return MatchmakingSnapshot{}
	}
	mm.mu.Lock()
	defer mm.mu.Unlock()

	snapshot := MatchmakingSnapshot{
		Queue:   make([]QueuePlayer, 0, len(mm.queue)),
		Matches: make([]LiveMatch, 0, len(mm.matches)),
	}
	now := time.Now()
	for _, entry := range mm.queue {
		snapshot.Queue = append(snapshot.Queue, mm.queuePlayerLocked(entry, now))
	}
	sort.Slice(snapshot.Queue, func(a, b int) bool {
		return snapshot.Queue[a].WaitingSince.Before(snapshot.Queue[b].WaitingSince)
	})
	for _, match := range mm.matches {
		live := LiveMatch{
			ID:                match.ID,
			QueueID:           match.Queue.ID,
			HostID:            match.HostID,
			MapName:           match.MapName,
			DisplayMapName:    DisplayMapName(match.MapName),
			MapIcon:           MapIconPath(match.MapName),
			MapID:             match.MapID,
			State:             match.State.String(),
			AdvertisementID:   match.AdvertisementID,
			PlatformSessionID: match.PlatformSessionID,
			CreatedAt:         match.CreatedAt,
			Players:           make([]QueuePlayer, 0, len(match.Players)),
		}
		for _, player := range match.Players {
			live.Players = append(live.Players, mm.queuePlayerLocked(player, now))
		}
		for _, player := range live.Players {
			if player.UserID == match.HostID {
				live.HostAlias = player.Alias
			}
		}
		snapshot.Matches = append(snapshot.Matches, live)
	}
	sort.Slice(snapshot.Matches, func(a, b int) bool {
		return snapshot.Matches[a].CreatedAt.Before(snapshot.Matches[b].CreatedAt)
	})
	return snapshot
}

func (mm *MainMatchmaking) queuePlayerLocked(entry *automatchQueueEntry, now time.Time) QueuePlayer {
	player := QueuePlayer{
		UserID:         entry.UserID,
		PollID:         entry.PollID,
		IP:             entry.LocalIP,
		MatchTypes:     append([]int32(nil), entry.MatchTypes...),
		VetoMapIDs:     append([]int32(nil), entry.VetoMapKeys...),
		VetoMaps:       append([]string(nil), entry.VetoMapNames...),
		PartySessionID: entry.PartySessionID,
		RelayRegion:    entry.RelayRegion,
		MatchID:        entry.MatchID,
		WaitingSince:   entry.CreatedAt,
	}
	if !entry.CreatedAt.IsZero() {
		player.WaitingSeconds = int(now.Sub(entry.CreatedAt).Seconds())
	}
	if mm.game != nil && mm.game.Users() != nil {
		if user, found := mm.game.Users().GetUserById(entry.UserID); found {
			player.Alias = user.GetAlias()
			player.ProfileID = user.GetProfileId()
		}
	}
	return player
}

// String names a match state for the dashboard.
func (s automatchMatchState) String() string {
	switch s {
	case automatchWaitingHost:
		return "waiting_host"
	case automatchHosting:
		return "hosting"
	case automatchReady:
		return "ready"
	default:
		return "unknown"
	}
}

func dashboardLevel(level uint16) uint16 {
	if level == 9_999 {
		return 0
	}
	return level
}

// OnlinePlayers returns every player the server knows about, marking the ones
// with an active session and annotating where they are playing.
func OnlinePlayers(game Game) []PublicPlayer {
	if game == nil || game.Users() == nil {
		return nil
	}
	users := game.Users()

	ipByUser := make(map[int32]string)
	lobbyByUser := make(map[int32]int32)
	if advs := game.Advertisements(); advs != nil {
		for _, lobby := range advs.Snapshot(game.Title()) {
			ipByUser[lobby.HostID] = lobby.IP
			lobbyByUser[lobby.HostID] = lobby.ID
			for _, peer := range lobby.Peers {
				if peer.IP != "" {
					ipByUser[peer.UserID] = peer.IP
				}
				lobbyByUser[peer.UserID] = lobby.ID
			}
		}
	}

	inQueue := make(map[int32]bool)
	if matchmaking := game.Matchmaking(); matchmaking != nil {
		for _, player := range matchmaking.Snapshot().Queue {
			inQueue[player.UserID] = true
			if player.IP != "" {
				ipByUser[player.UserID] = player.IP
			}
		}
	}

	sessions := game.Sessions()
	players := make([]PublicPlayer, 0, 16)
	for userID := range users.GetUserIds() {
		user, found := users.GetUserById(userID)
		if !found {
			continue
		}
		player := PublicPlayer{
			UserID:          user.GetId(),
			ProfileID:       user.GetProfileId(),
			StatGroupID:     user.GetStatGroupID(),
			Alias:           user.GetAlias(),
			Xbox:            user.GetXbox(),
			Level:           dashboardLevel(user.GetProfileLevel()),
			XP:              user.GetProfileExperience(),
			Rating:          user.GetRating(),
			IP:              ipByUser[user.GetId()],
			AdvertisementID: lobbyByUser[user.GetId()],
			InQueue:         inQueue[user.GetId()],
		}
		if sessions != nil {
			_, player.Online = sessions.GetByUserId(user.GetId())
		}
		players = append(players, player)
	}
	sort.Slice(players, func(a, b int) bool {
		if players[a].Online != players[b].Online {
			return players[a].Online
		}
		return players[a].Alias < players[b].Alias
	})
	return players
}
