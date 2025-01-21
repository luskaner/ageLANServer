package models

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/advertisement/shared"
	"sync"
	"time"
)

type ModDll struct {
	file     string
	checksum int32
}

type Observers struct {
	enabled  bool
	delay    uint32
	password string
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
	host              *MainUser
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
	chat              []*MainMessage
	peers             *orderedmap.OrderedMap[*MainUser, *MainPeer]
	lock              *sync.RWMutex
	chatLock          *sync.RWMutex
	peerLock          *i.KeyRWMutex
}

type MainAdvertisements struct {
	store *i.SafeMap[int32, *MainAdvertisement]
	users *MainUsers
}

func (advs *MainAdvertisements) Initialize(users *MainUsers) {
	advs.store = i.NewSafeMap[int32, *MainAdvertisement]()
	advs.users = users
}

func (adv *MainAdvertisement) GetModDllChecksum() int32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.modDll.checksum
}

func (adv *MainAdvertisement) GetModDllFile() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.modDll.file
}

func (adv *MainAdvertisement) GetPasswordValue() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.password.value
}

func (adv *MainAdvertisement) GetStartTime() int64 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.startTime
}

func (adv *MainAdvertisement) GetState() int8 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.state
}

func (adv *MainAdvertisement) GetId() int32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.id
}

func (adv *MainAdvertisement) GetDescription() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.description
}

func (adv *MainAdvertisement) GetRelayRegion() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.relayRegion
}

func (adv *MainAdvertisement) GetJoinable() bool {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.joinable
}

func (adv *MainAdvertisement) GetVisible() bool {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.visible
}

func (adv *MainAdvertisement) GetHost() *MainUser {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.host
}

func (adv *MainAdvertisement) GetAppBinaryChecksum() int32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.appBinaryChecksum
}

func (adv *MainAdvertisement) GetDataChecksum() int32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.dataChecksum
}

func (adv *MainAdvertisement) GetMatchType() uint8 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.matchType
}

func (adv *MainAdvertisement) GetModName() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.modName
}

func (adv *MainAdvertisement) GetModVersion() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.modVersion
}

func (adv *MainAdvertisement) GetIp() string {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.ip
}

func (adv *MainAdvertisement) GetVersionFlags() uint32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.versionFlags
}

func (adv *MainAdvertisement) GetPlatformSessionId() uint64 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.platformSessionId
}

func (adv *MainAdvertisement) GetObserversDelay() uint32 {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.observers.delay
}

func (adv *MainAdvertisement) GetPeers() *orderedmap.OrderedMap[*MainUser, *MainPeer] {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	return adv.peers
}

func (adv *MainAdvertisement) GetPeer(user *MainUser) (*MainPeer, bool) {
	adv.lock.RLock()
	defer adv.lock.RUnlock()
	userId := user.GetId()
	adv.peerLock.RLock(userId)
	defer adv.peerLock.RUnlock(userId)
	u, exists := adv.peers.Get(user)
	if !exists {
		return nil, false
	}
	return u, true
}

func (advs *MainAdvertisements) Store(advFrom *shared.AdvertisementHostRequest) *MainAdvertisement {
	if advFrom.Id != -1 {
		return nil
	}
	var id int32
	for {
		func() {
			i.RngLock.Lock()
			defer i.RngLock.Unlock()
			id = i.Rng.Int32(math.MaxInt32)
		}()
		_, exists := advs.store.Load(id)
		if !exists {
			adv := &MainAdvertisement{
				lock:     &sync.RWMutex{},
				chatLock: &sync.RWMutex{},
				peerLock: i.NewKeyRWMutex(),
			}
			adv.id = id
			func() {
				i.RngLock.Lock()
				defer i.RngLock.Unlock()
				adv.ip = fmt.Sprintf("/10.0.11.%d", i.Rng.IntN(254)+1)
			}()
			adv.relayRegion = advFrom.RelayRegion
			adv.party = advFrom.Party
			adv.race = advFrom.Race
			adv.team = advFrom.Team
			adv.statGroup = advFrom.StatGroup
			adv.peers = orderedmap.NewOrderedMap[*MainUser, *MainPeer]()
			adv.chat = make([]*MainMessage, 0)
			advs.update(adv, &shared.AdvertisementUpdateRequest{
				Id:                adv.id,
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
				HostId:            advFrom.HostId,
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
				PlatformSessionId: advFrom.PlatformSessionId,
				State:             advFrom.State,
			})
			advs.store.Store(id, adv)
			return adv
		}
	}
}

func (adv *MainAdvertisement) AddMessage(broadcast bool, content string, typeId uint8, sender *MainUser, receivers []*MainUser) *MainMessage {
	message := &MainMessage{
		advertisementId: adv.GetId(),
		time:            time.Now().UTC().Unix(),
		broadcast:       broadcast,
		content:         content,
		typ:             typeId,
		sender:          sender,
		receivers:       receivers,
	}
	adv.chatLock.Lock()
	defer adv.chatLock.Unlock()
	adv.chat = append(adv.chat, message)
	return message
}

func (advs *MainAdvertisements) Update(adv *MainAdvertisement, advFrom *shared.AdvertisementUpdateRequest) {
	advs.update(adv, advFrom)
}

func (advs *MainAdvertisements) update(adv *MainAdvertisement, advFrom *shared.AdvertisementUpdateRequest) {
	func() {
		adv.lock.Lock()
		defer adv.lock.Unlock()
		if adv.host != nil {
			adv.host.SetAdvertisement(nil)
			adv.host = nil
		}
		adv.host, _ = advs.users.GetUserById(advFrom.HostId)
		adv.host.SetAdvertisement(adv)
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
		adv.platformSessionId = advFrom.PlatformSessionId
	}()
	adv.UpdateState(advFrom.State)
}

func (advs *MainAdvertisements) GetAdvertisement(id int32) (*MainAdvertisement, bool) {
	return advs.store.Load(id)
}

func (advs *MainAdvertisements) NewPeer(adv *MainAdvertisement, u *MainUser, race int32, team int32) *MainPeer {
	if peer, ok := adv.GetPeer(u); ok {
		// Ignore already added peers (via host & join)
		return peer
	}
	peer := &MainPeer{
		advertisement: adv,
		user:          u,
		race:          race,
		team:          team,
		invites:       mapset.NewSet[*MainUser](),
		lock:          &sync.RWMutex{},
	}
	userId := peer.user.GetId()
	adv.peerLock.Lock(userId)
	defer adv.peerLock.Unlock(userId)
	adv.peers.Set(peer.user, peer)
	u.SetAdvertisement(adv)
	return peer
}

func (advs *MainAdvertisements) RemovePeer(adv *MainAdvertisement, user *MainUser) {
	func() {
		adv.peerLock.Lock(user.GetId())
		defer adv.peerLock.Unlock(user.GetId())
		adv.peers.Delete(user)
		user.SetAdvertisement(nil)
	}()
	if adv.host == user {
		advs.Delete(adv)
	}
}

func (adv *MainAdvertisement) UpdatePeer(user *MainUser, race int32, team int32) {
	userId := user.GetId()
	adv.peerLock.Lock(userId)
	defer adv.peerLock.Unlock(userId)
	peer, _ := adv.peers.Get(user)
	peer.Update(race, team)
}

func (advs *MainAdvertisements) Delete(adv *MainAdvertisement) {
	adv.lock.Lock()
	defer adv.lock.Unlock()
	advs.store.Delete(adv.id)
	adv.host.SetAdvertisement(nil)
	for el := range adv.peers.Values() {
		el.GetUser().SetAdvertisement(nil)
	}
}

func (adv *MainAdvertisement) UpdateState(state int8) {
	adv.lock.Lock()
	defer adv.lock.Unlock()
	previousState := adv.state
	adv.state = state
	if adv.state == 1 && previousState != 1 {
		adv.startTime = time.Now().UTC().Unix()
		adv.visible = false
		adv.joinable = false
	}
}

func (adv *MainAdvertisement) UpdatePlatformSessionId(sessionId uint64) {
	adv.lock.Lock()
	defer adv.lock.Unlock()
	adv.platformSessionId = sessionId
}

func (adv *MainAdvertisement) EncodePeers() i.A {
	var peers = make(i.A, adv.peers.Len())
	j := 0
	for key, p := range adv.peers.AllFromFront() {
		userId := key.GetId()
		func() {
			adv.peerLock.RLock(userId)
			defer adv.peerLock.RUnlock(userId)
			peers[j] = p.Encode()
		}()
		j++
	}
	return peers
}

func (adv *MainAdvertisement) Encode(gameId string) i.A {
	var visible uint8
	adv.lock.RLock()
	defer adv.lock.RUnlock()
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
		"0",
	}
	if gameId == common.GameAoE2 {
		response = append(
			response,
			"",
			"",
			"0",
		)
	}
	response = append(
		response,
		adv.host.GetId(),
		started,
		adv.description,
	)
	if gameId == common.GameAoE2 {
		response = append(response, adv.description)
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
		0,
		0,
		0,
		adv.GetObserversDelay(),
		1,
		1,
		startTime,
		adv.relayRegion,
	)
	if gameId != common.GameAoE1 {
		response = append(response, nil)
	}
	return response
}

func (advs *MainAdvertisements) FindAdvertisements(matches func(adv *MainAdvertisement) bool) []*MainAdvertisement {
	var res []*MainAdvertisement
	for _, adv := range advs.store.IterValues() {
		func() {
			adv.lock.RLock()
			defer adv.lock.RUnlock()
			if matches(adv) {
				res = append(res, adv)
			}
		}()
	}
	return res
}

func (advs *MainAdvertisements) FindAdvertisementsEncoded(gameId string, matches func(adv *MainAdvertisement) bool) []i.A {
	var res []i.A
	advsOriginal := advs.FindAdvertisements(matches)
	for _, adv := range advsOriginal {
		func() {
			adv.lock.RLock()
			defer adv.lock.RUnlock()
			res = append(res, adv.Encode(gameId))
		}()
	}
	return res
}
