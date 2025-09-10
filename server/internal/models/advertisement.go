package models

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
)

type ModDll struct {
	file     string
	checksum int32
}

type Observers struct {
	enabled  bool
	delay    uint32
	password string
	userIds  *i.SafeSet[int32]
}

type Password struct {
	value   string
	enabled bool
}

type MainAdvertisement struct {
	id                int32
	ip                string
	automatchPollId   int32
	relayRegion       string
	appBinaryChecksum int32
	mapName           string
	description       string
	dataChecksum      int32
	hostId            int32
	modDll            ModDll
	modName           string
	modVersion        string
	observers         Observers
	password          Password
	visible           bool
	party             int32
	race              int32
	team              int32
	statGroup         int32
	versionFlags      uint32
	joinable          bool
	matchType         uint8
	maxPlayers        uint8
	options           string
	slotInfo          string
	platformSessionId uint64
	state             int8
	startTime         int64
	peers             *i.SafeOrderedMap[int32, *MainPeer]
	metadata          string
}

type MainAdvertisements struct {
	store         *i.SafeOrderedMap[int32, *MainAdvertisement]
	locks         *i.KeyRWMutex[int32]
	users         *MainUsers
	battleServers *MainBattleServers
}

func (advs *MainAdvertisements) Initialize(users *MainUsers, battleServers *MainBattleServers) {
	advs.store = i.NewSafeOrderedMap[int32, *MainAdvertisement]()
	advs.locks = i.NewKeyRWMutex[int32]()
	advs.users = users
	advs.battleServers = battleServers
}

func (adv *MainAdvertisement) GetMetadata() string {
	return adv.metadata
}

// UnsafeGetModDllChecksum requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetModDllChecksum() int32 {
	return adv.modDll.checksum
}

// UnsafeGetModDllFile requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetModDllFile() string {
	return adv.modDll.file
}

// UnsafeGetPasswordValue requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetPasswordValue() string {
	return adv.password.value
}

// UnsafeGetStartTime requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetStartTime() int64 {
	return adv.startTime
}

// UnsafeGetState requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetState() int8 {
	return adv.state
}

func (adv *MainAdvertisement) GetId() int32 {
	return adv.id
}

// UnsafeGetDescription requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetDescription() string {
	return adv.description
}

func (adv *MainAdvertisement) GetRelayRegion() string {
	return adv.relayRegion
}

// UnsafeGetJoinable requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetJoinable() bool {
	return adv.joinable
}

// UnsafeGetVisible requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetVisible() bool {
	return adv.visible
}

func (adv *MainAdvertisement) GetHostId() int32 {
	return adv.hostId
}

// UnsafeGetAppBinaryChecksum requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetAppBinaryChecksum() int32 {
	return adv.appBinaryChecksum
}

// UnsafeGetDataChecksum requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetDataChecksum() int32 {
	return adv.dataChecksum
}

// UnsafeGetMatchType requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetMatchType() uint8 {
	return adv.matchType
}

// UnsafeGetModName requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetModName() string {
	return adv.modName
}

// UnsafeGetModVersion requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetModVersion() string {
	return adv.modVersion
}

func (adv *MainAdvertisement) GetIp() string {
	return adv.ip
}

// UnsafeGetVersionFlags requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetVersionFlags() uint32 {
	return adv.versionFlags
}

// UnsafeGetPlatformSessionId requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetPlatformSessionId() uint64 {
	return adv.platformSessionId
}

// UnsafeGetObserversDelay requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetObserversDelay() uint32 {
	return adv.observers.delay
}

// UnsafeGetObserversEnabled requires advertisement read lock
func (adv *MainAdvertisement) UnsafeGetObserversEnabled() bool {
	return adv.observers.enabled
}

func (adv *MainAdvertisement) GetPeers() *i.SafeOrderedMap[int32, *MainPeer] {
	return adv.peers
}

func (advs *MainAdvertisements) Store(advFrom *shared.AdvertisementHostRequest, generateMetadata bool) *MainAdvertisement {
	adv := &MainAdvertisement{}
	i.WithRng(func(rand *rand.Rand) {
		adv.ip = fmt.Sprintf("/10.0.11.%d", rand.IntN(254)+1)
	})
	adv.relayRegion = advFrom.RelayRegion
	if generateMetadata {
		adv.metadata = fmt.Sprintf(
			`{"templateName":"GameSession","name":"%s","scid":"00000000-0000-0000-0000-000068a451d4"}`,
			uuid.New().String(),
		)
	} else {
		adv.metadata = "0"
	}
	adv.hostId = advFrom.HostId
	adv.party = advFrom.Party
	adv.race = advFrom.Race
	adv.team = advFrom.Team
	adv.statGroup = advFrom.StatGroup
	adv.peers = i.NewSafeOrderedMap[int32, *MainPeer]()
	advs.UpdateUnsafe(adv, &shared.AdvertisementUpdateRequest{
		AppBinaryChecksum: advFrom.AppBinaryChecksum,
		DataChecksum:      advFrom.DataChecksum,
		ModDllChecksum:    advFrom.ModDllChecksum,
		ModDllFile:        advFrom.ModDllFile,
		ModName:           advFrom.ModName,
		ModVersion:        advFrom.ModVersion,
		VersionFlags:      advFrom.VersionFlags,
		Description:       advFrom.Description,
		AutomatchPollId:   advFrom.AutomatchPollId,
		MapName:           advFrom.MapName,
		Observable:        advFrom.Observable,
		ObserverPassword:  advFrom.ObserverPassword,
		ObserverDelay:     advFrom.ObserverDelay,
		Password:          advFrom.Password,
		Passworded:        advFrom.Passworded,
		Visible:           advFrom.Visible,
		Joinable:          advFrom.Joinable,
		MatchType:         advFrom.MatchType,
		MaxPlayers:        advFrom.MaxPlayers,
		Options:           advFrom.Options,
		SlotInfo:          advFrom.SlotInfo,
		State:             advFrom.State,
	})
	exists := true
	var storedAdv *MainAdvertisement
	for exists {
		i.WithRng(func(rand *rand.Rand) {
			adv.id = rand.Int32()
		})
		exists, storedAdv = advs.store.Store(adv.id, adv, func(_ *MainAdvertisement) bool {
			return false
		})
	}
	return storedAdv
}

func (adv *MainAdvertisement) MakeMessage(broadcast bool, content string, typeId uint8, sender *MainUser, receivers []*MainUser) *MainMessage {
	return &MainMessage{
		advertisementId: adv.GetId(),
		time:            time.Now().UTC().Unix(),
		broadcast:       broadcast,
		content:         content,
		typ:             typeId,
		sender:          sender,
		receivers:       receivers,
	}
}

func (advs *MainAdvertisements) WithReadLock(id int32, action func()) {
	advs.locks.RLock(id)
	defer advs.locks.RUnlock(id)
	action()
}

func (advs *MainAdvertisements) WithWriteLock(id int32, action func()) {
	advs.locks.Lock(id)
	defer advs.locks.Unlock(id)
	action()
}

// UpdateUnsafe is safe only if adv has not been stored yet
func (advs *MainAdvertisements) UpdateUnsafe(adv *MainAdvertisement, advFrom *shared.AdvertisementUpdateRequest) {
	adv.hostId = advFrom.HostId
	adv.automatchPollId = advFrom.AutomatchPollId
	adv.appBinaryChecksum = advFrom.AppBinaryChecksum
	adv.mapName = advFrom.MapName
	adv.description = advFrom.Description
	adv.dataChecksum = advFrom.DataChecksum
	adv.modDll.checksum = advFrom.ModDllChecksum
	adv.modDll.file = advFrom.ModDllFile
	adv.modName = advFrom.ModName
	adv.modVersion = advFrom.ModVersion
	adv.observers.delay = advFrom.ObserverDelay
	adv.observers.enabled = advFrom.Observable
	adv.observers.password = advFrom.ObserverPassword
	adv.password.enabled = advFrom.Passworded
	adv.password.value = advFrom.Password
	adv.visible = advFrom.Visible
	adv.versionFlags = advFrom.VersionFlags
	adv.joinable = advFrom.Joinable
	adv.matchType = advFrom.MatchType
	adv.maxPlayers = advFrom.MaxPlayers
	adv.options = advFrom.Options
	adv.slotInfo = advFrom.SlotInfo
	adv.UnsafeUpdateState(advFrom.State)
}

func (advs *MainAdvertisements) GetAdvertisement(id int32) (*MainAdvertisement, bool) {
	return advs.store.Load(id)
}

// UnsafeNewPeer requires advertisement write lock
func (advs *MainAdvertisements) UnsafeNewPeer(advertisementId int32, advertisementIp string, userId int32, userStatId int32, race int32, team int32) *MainPeer {
	adv, exists := advs.GetAdvertisement(advertisementId)
	if !exists {
		return nil
	}
	peer := NewPeer(advertisementId, advertisementIp, userId, userStatId, race, team)
	_, storedPeer := adv.peers.Store(peer.userId, peer, func(_ *MainPeer) bool {
		return false
	})
	return storedPeer
}

// UnsafeRemovePeer requires advertisement write lock
func (advs *MainAdvertisements) UnsafeRemovePeer(advertisementId int32, userId int32) bool {
	adv, exists := advs.GetAdvertisement(advertisementId)
	if !exists {
		return false
	}
	if !adv.peers.Delete(userId) {
		return false
	}
	if adv.hostId == userId {
		advs.UnsafeDelete(adv)
	}
	return true
}

// UnsafeDelete requires advertisement write lock
func (advs *MainAdvertisements) UnsafeDelete(adv *MainAdvertisement) {
	advs.store.Delete(adv.id)
}

// UnsafeUpdateState is only safe if advertisement has not been added yet
func (adv *MainAdvertisement) UnsafeUpdateState(state int8) {
	previousState := adv.state
	adv.state = state
	if adv.state == 1 && previousState != 1 {
		adv.startTime = time.Now().UTC().Unix()
		adv.visible = false
		adv.joinable = false
		adv.observers.userIds = i.NewSafeSet[int32]()
	}
}

// UnsafeUpdatePlatformSessionId requires advertisement write lock
func (adv *MainAdvertisement) UnsafeUpdatePlatformSessionId(sessionId uint64) {
	adv.platformSessionId = sessionId
}

func (adv *MainAdvertisement) StartObserving(userId int32) {
	adv.observers.userIds.Store(userId)
}

func (adv *MainAdvertisement) StopObserving(userId int32) {
	adv.observers.userIds.Delete(userId)
}

func (adv *MainAdvertisement) EncodePeers() []i.A {
	peersLen, peers := adv.peers.Values()
	encodedPeers := make([]i.A, peersLen)
	j := 0
	for p := range peers {
		encodedPeers[j] = p.Encode()
		j++
	}
	return encodedPeers
}

// UnsafeEncode requires advertisement read lock
func (adv *MainAdvertisement) UnsafeEncode(gameId string, battleServers *MainBattleServers) i.A {
	var visible uint8
	if adv.visible {
		visible = 1
	} else {
		visible = 0
	}
	var passworded uint8
	if adv.password.enabled {
		passworded = 1
	} else {
		passworded = 0
	}
	var startTime *int64
	if adv.startTime != 0 {
		startTime = &adv.startTime
	} else {
		startTime = nil
	}
	var started uint8
	if startTime != nil {
		started = 1
	} else {
		started = 0
	}
	response := i.A{
		adv.id,
		adv.platformSessionId,
	}
	if gameId == common.GameAoE2 {
		response = append(
			response,
			"0",
			"",
			"",
		)
	}
	response = append(
		response,
		adv.GetMetadata(),
		adv.hostId,
		started,
		adv.description,
	)
	if gameId == common.GameAoE2 {
		response = append(response, adv.description)
	}
	lan := 1
	var battleServer *MainBattleServer
	var battleServerExists bool
	if battleServer, battleServerExists = battleServers.Get(adv.relayRegion); battleServerExists {
		lan = 0
	}
	response = append(
		response,
		visible,
		adv.mapName,
		adv.options,
		passworded,
		adv.maxPlayers,
		adv.slotInfo,
		adv.matchType,
		adv.EncodePeers(),
		adv.observers.userIds.Len(),
		0,
		0,
		adv.observers.delay,
		1,
		lan,
		startTime,
		adv.relayRegion,
	)
	if battleServerExists {
		battleServer.AppendName(&response)
	} else {
		response = append(response, nil)
	}
	return response
}

// UnsafeFirstAdvertisement requires advertisement read lock unless only safe advertisement properties are checked
func (advs *MainAdvertisements) UnsafeFirstAdvertisement(matches func(adv *MainAdvertisement) bool) *MainAdvertisement {
	_, iter := advs.store.Values()
	for adv := range iter {
		if matches(adv) {
			return adv
		}
	}
	return nil
}

func (advs *MainAdvertisements) LockedFindAdvertisementsEncoded(gameId string, length int, offset int, preMatchesLocking bool, matches func(adv *MainAdvertisement) bool) []i.A {
	var res []i.A
	_, iter := advs.store.Values()
	for adv := range iter {
		advId := adv.GetId()
		if preMatchesLocking {
			func() {
				advs.locks.RLock(advId)
				defer advs.locks.RUnlock(advId)
				if matches(adv) {
					res = append(res, adv.UnsafeEncode(gameId, advs.battleServers))
				}
			}()
		} else {
			advs.WithReadLock(adv.GetId(), func() {
				res = append(res, adv.UnsafeEncode(gameId, advs.battleServers))
			})
		}
	}
	if offset >= len(res) {
		return []i.A{}
	}
	if length == 0 {
		length = len(res)
	}
	end := length + offset
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end]
}

func (advs *MainAdvertisements) GetUserAdvertisement(userId int32) *MainAdvertisement {
	return advs.UnsafeFirstAdvertisement(func(adv *MainAdvertisement) bool {
		_, peerIter := adv.peers.Keys()
		for usId := range peerIter {
			if usId == userId {
				return true
			}
		}
		return false
	})
}
