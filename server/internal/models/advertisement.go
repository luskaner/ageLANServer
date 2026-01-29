package models

import (
	"fmt"
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

type Tags struct {
	integer map[string]int32
	text    map[string]string
}

type Advertisement interface {
	GetXboxSessionId() string
	UnsafeGetModDllChecksum() int32
	UnsafeGetModDllFile() string
	UnsafeGetPasswordValue() string
	UnsafeGetStartTime() int64
	UnsafeGetState() int8
	UnsafeGetDescription() string
	GetRelayRegion() string
	GetParty() int32
	UnsafeGetVisible() bool
	UnsafeGetJoinable() bool
	UnsafeGetAppBinaryChecksum() int32
	UnsafeGetDataChecksum() int32
	UnsafeGetMatchType() uint8
	UnsafeGetModName() string
	UnsafeGetModVersion() string
	UnsafeGetVersionFlags() uint32
	UnsafeGetPlatformSessionId() uint64
	UnsafeGetObserversDelay() uint32
	UnsafeGetObserversEnabled() bool
	UnsafeSetHostId(hostId int32)
	UnsafeUpdateState(state int8)
	UnsafeUpdatePlatformSessionId(sessionId uint64)
	UnsafeUpdateTags(integer map[string]int32, text map[string]string)
	UnsafeMatchesTags(integer map[string]int32, text map[string]string) bool
	UnsafeEncode(gameId string, battleServers BattleServers) i.A
	UnsafeUpdate(advFrom *shared.AdvertisementUpdateRequest)
	GetId() int32
	GetIp() string
	UnsafeGetHostId() int32
	GetPeers() *i.SafeOrderedMap[int32, Peer]
	MakeMessage(broadcast bool, content string, typeId uint8, sender User, receivers []User) Message
	StartObserving(userId int32)
	StopObserving(userId int32)
	EncodePeers() []i.A
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
	lan               bool
	startTime         int64
	peers             *i.SafeOrderedMap[int32, Peer]
	xboxSessionId     string
	tags              Tags
}

func containsFilter[M ~map[K]V, K, V comparable](filter M, tags M) bool {
	for fk, fv := range filter {
		if tv, ok := tags[fk]; !ok || fv != tv {
			return false
		}
	}
	return true
}

type Advertisements interface {
	Initialize(users Users, battleServers BattleServers)
	Store(advFrom *shared.AdvertisementHostRequest, generateMetadata bool, gameId string) Advertisement
	WithReadLock(id int32, action func())
	WithWriteLock(id int32, action func())
	GetAdvertisement(id int32) (Advertisement, bool)
	UnsafeNewPeer(advertisementId int32, advertisementIp string, userId int32, userStatId int32, party int32, race int32, team int32) Peer
	UnsafeRemovePeer(advertisementId int32, userId int32) bool
	UnsafeDelete(adv Advertisement)
	UnsafeFirstAdvertisement(matches func(adv Advertisement) bool) Advertisement
	LockedFindAdvertisementsEncoded(gameId string, length int, offset int, preMatchesLocking bool, matches func(adv Advertisement) bool) i.A
	GetUserAdvertisement(userId int32) Advertisement
}

type MainAdvertisements struct {
	store         *i.SafeOrderedMap[int32, Advertisement]
	locks         *i.KeyRWMutex[int32]
	users         Users
	battleServers BattleServers
}

func (advs *MainAdvertisements) Initialize(users Users, battleServers BattleServers) {
	advs.store = i.NewSafeOrderedMap[int32, Advertisement]()
	advs.locks = i.NewKeyRWMutex[int32]()
	advs.users = users
	advs.battleServers = battleServers
}

func (adv *MainAdvertisement) GetXboxSessionId() string {
	return adv.xboxSessionId
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

func (adv *MainAdvertisement) UnsafeGetHostId() int32 {
	return adv.hostId
}

func (adv *MainAdvertisement) GetParty() int32 {
	return adv.party
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

func (adv *MainAdvertisement) GetPeers() *i.SafeOrderedMap[int32, Peer] {
	return adv.peers
}

func (advs *MainAdvertisements) Store(advFrom *shared.AdvertisementHostRequest, generateXboxSessionId bool, gameId string) Advertisement {
	adv := &MainAdvertisement{}
	i.WithRng(func(rand *i.RandReader) {
		adv.ip = fmt.Sprintf("/10.0.11.%d", rand.IntN(254)+1)
	})
	adv.relayRegion = advFrom.RelayRegion
	if generateXboxSessionId {
		// FIXME: This might be just slowing things down as the session is not valid
		var scidEnd string
		switch gameId {
		case common.GameAoM:
			scidEnd = "00006fe8b971"
		case common.GameAoE4:
			scidEnd = "00007d18f66e"
		default:
			scidEnd = "000068a451d4"
		}
		adv.xboxSessionId = fmt.Sprintf(
			`{"templateName":"GameSession","name":"%s","scid":"00000000-0000-0000-0000-%s"}`,
			uuid.New().String(),
			scidEnd,
		)
	} else {
		adv.xboxSessionId = "0"
	}
	adv.hostId = advFrom.HostId
	adv.party = advFrom.Party
	adv.race = advFrom.Race
	adv.team = advFrom.Team
	adv.statGroup = advFrom.StatGroup
	adv.lan = i.NumberToBool(advFrom.ServiceType)
	adv.tags.text = make(map[string]string)
	adv.tags.integer = make(map[string]int32)
	adv.peers = i.NewSafeOrderedMap[int32, Peer]()
	adv.UnsafeUpdate(&shared.AdvertisementUpdateRequest{
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
	var storedAdv Advertisement
	for exists {
		i.WithRng(func(rand *i.RandReader) {
			adv.id = rand.Int32()
		})
		exists, storedAdv = advs.store.Store(adv.id, adv, func(_ Advertisement) bool {
			return false
		})
	}
	return storedAdv
}

func (adv *MainAdvertisement) MakeMessage(broadcast bool, content string, typeId uint8, sender User, receivers []User) Message {
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

// UnsafeUpdate is safe only if adv has not been stored yet
func (adv *MainAdvertisement) UnsafeUpdate(advFrom *shared.AdvertisementUpdateRequest) {
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

func (advs *MainAdvertisements) GetAdvertisement(id int32) (Advertisement, bool) {
	return advs.store.Load(id)
}

// UnsafeNewPeer requires advertisement write lock
func (advs *MainAdvertisements) UnsafeNewPeer(advertisementId int32, advertisementIp string, userId int32, userStatId int32, party int32, race int32, team int32) Peer {
	adv, exists := advs.GetAdvertisement(advertisementId)
	if !exists {
		return nil
	}
	peer := NewPeer(advertisementId, advertisementIp, userId, userStatId, party, race, team)
	_, storedPeer := adv.GetPeers().Store(peer.GetUserId(), peer, func(_ Peer) bool {
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
	if !adv.GetPeers().Delete(userId) {
		return false
	}
	if adv.GetPeers().Len() == 0 {
		advs.UnsafeDelete(adv)
	}
	return true
}

// UnsafeDelete requires advertisement write lock
func (advs *MainAdvertisements) UnsafeDelete(adv Advertisement) {
	advs.store.Delete(adv.GetId())
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

// UnsafeSetHostId requires advertisement write lock
func (adv *MainAdvertisement) UnsafeSetHostId(hostId int32) {
	adv.hostId = hostId
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

// UnsafeUpdateTags requires advertisement write lock
func (adv *MainAdvertisement) UnsafeUpdateTags(integer map[string]int32, text map[string]string) {
	adv.tags.integer = integer
	adv.tags.text = text
}

// UnsafeMatchesTags requires advertisement read lock
func (adv *MainAdvertisement) UnsafeMatchesTags(integer map[string]int32, text map[string]string) bool {
	return containsFilter(integer, adv.tags.integer) && containsFilter(text, adv.tags.text)
}

// UnsafeEncode requires advertisement read lock
func (adv *MainAdvertisement) UnsafeEncode(gameId string, battleServers BattleServers) i.A {
	var startTime *int64
	if i.NumberToBool(adv.startTime) {
		startTime = &adv.startTime
	} else {
		startTime = nil
	}
	response := i.A{
		adv.id,
		adv.platformSessionId,
	}
	if gameId == common.GameAoE2 || gameId == common.GameAoM || gameId == common.GameAoE4 {
		// goodolggameslobbyid (GoG lobby ID), always 0
		if gameId == common.GameAoE4 {
			response = append(response, 0)
		} else {
			response = append(response, "0")
		}
		response = append(
			response,
			"",
			"",
		)
	}
	response = append(
		response,
		adv.GetXboxSessionId(),
		adv.hostId,
		adv.state,
		adv.description,
	)
	if gameId == common.GameAoE2 || gameId == common.GameAoM || gameId == common.GameAoE4 {
		response = append(response, adv.description)
	}
	response = append(
		response,
		i.NewBoolMappedNumberFromBool(adv.visible),
		adv.mapName,
		adv.options,
		i.NewBoolMappedNumberFromBool(adv.password.enabled),
		adv.maxPlayers,
		adv.slotInfo,
		adv.matchType,
		adv.EncodePeers(),
		adv.observers.userIds.Len(),
		0, // observermax, can it be 0 so it means without limit?
		i.NewBoolMappedNumberFromBool(adv.observers.enabled),
		adv.observers.delay,
		// TODO: Ensure if the order of password and the next is correct
		i.NewBoolMappedNumberFromBool(adv.observers.password != ""),
		i.NewBoolMappedNumberFromBool(adv.lan),
		startTime,
		adv.relayRegion,
	)
	if adv.lan {
		response = append(response, nil)
	} else {
		battleServer, _ := battleServers.Get(adv.relayRegion)
		battleServer.AppendName(&response)
	}
	return response
}

// UnsafeFirstAdvertisement requires advertisement read lock unless only safe advertisement properties are checked
func (advs *MainAdvertisements) UnsafeFirstAdvertisement(matches func(adv Advertisement) bool) Advertisement {
	_, iter := advs.store.Values()
	for adv := range iter {
		if matches(adv) {
			return adv
		}
	}
	return nil
}

func (advs *MainAdvertisements) LockedFindAdvertisementsEncoded(gameId string, length int, offset int, preMatchesLocking bool, matches func(adv Advertisement) bool) i.A {
	var res i.A
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
		return i.A{}
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

func (advs *MainAdvertisements) GetUserAdvertisement(userId int32) Advertisement {
	return advs.UnsafeFirstAdvertisement(func(adv Advertisement) bool {
		_, peerIter := adv.GetPeers().Keys()
		for usId := range peerIter {
			if usId == userId {
				return true
			}
		}
		return false
	})
}
