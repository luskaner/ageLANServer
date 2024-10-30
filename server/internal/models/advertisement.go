package models

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
	"github.com/wk8/go-ordered-map/v2"
	"math"
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
	peers *i.SafeSet[*MainUser]
	hosts *i.SafeSet[*MainUser]
	store *i.SafeMap[int32, *MainAdvertisement]
	users *MainUsers
}

func (advs *MainAdvertisements) Initialize(users *MainUsers) {
	advs.peers = i.NewSafeSet[*MainUser]()
	advs.hosts = i.NewSafeSet[*MainUser]()
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
		i.RngLock.Lock()
		id = i.Rng.Int31n(math.MaxInt32)
		i.RngLock.Unlock()
		_, exists := advs.store.Load(id)
		if !exists {
			adv := &MainAdvertisement{
				lock:     &sync.RWMutex{},
				chatLock: &sync.RWMutex{},
				peerLock: i.NewKeyRWMutex(),
			}
			adv.id = id
			i.RngLock.Lock()
			adv.ip = fmt.Sprintf("/10.0.11.%d", i.Rng.Intn(254)+1)
			i.RngLock.Unlock()
			adv.relayRegion = advFrom.RelayRegion
			adv.party = advFrom.Party
			adv.race = advFrom.Race
			adv.team = advFrom.Team
			adv.statGroup = advFrom.StatGroup
			adv.peers = orderedmap.New[*MainUser, *MainPeer]()
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
	adv.lock.Lock()
	if adv.host != nil {
		advs.removeHost(adv.host)
		adv.host = nil
	}
	adv.host, _ = advs.users.GetUserById(advFrom.HostId)
	advs.addHost(adv.host)
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
	adv.lock.Unlock()
	adv.UpdateState(advFrom.State)
}

func (advs *MainAdvertisements) GetAdvertisement(id int32) (*MainAdvertisement, bool) {
	return advs.store.Load(id)
}

func (advs *MainAdvertisements) NewPeer(adv *MainAdvertisement, u *MainUser, race int32, team int32) *MainPeer {
	if advs.isPeer(u) {
		// Ignore already added peers (via host & join)
		return nil
	}
	peer := &MainPeer{
		advertisement: adv,
		user:          u,
		race:          race,
		team:          team,
		invites:       i.NewSafeSet[*MainUser](),
		lock:          &sync.RWMutex{},
	}
	userId := peer.user.GetId()
	adv.peerLock.Lock(userId)
	defer adv.peerLock.Unlock(userId)
	adv.peers.Set(peer.user, peer)
	advs.addPeer(u)
	return peer
}

func (advs *MainAdvertisements) RemovePeer(adv *MainAdvertisement, user *MainUser) {
	adv.peerLock.Lock(user.GetId())
	adv.peers.Delete(user)
	advs.removePeer(user)
	adv.peerLock.Unlock(user.GetId())
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
	advs.removeHost(adv.host)
	for el := adv.peers.Oldest(); el != nil; el = el.Next() {
		advs.removePeer(el.Key)
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

func (adv *MainAdvertisement) EncodePeers() i.A {
	var peers = make(i.A, adv.peers.Len())
	j := 0
	for el := adv.peers.Oldest(); el != nil; el = el.Next() {
		p := el.Value
		userId := el.Key.GetId()
		adv.peerLock.RLock(userId)
		peers[j] = p.Encode()
		adv.peerLock.RUnlock(userId)
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
		0,
		1,
		1,
		startTime,
		adv.relayRegion,
		nil,
	)
	return response
}

func (advs *MainAdvertisements) FindAdvertisements(matches func(adv *MainAdvertisement) bool) []*MainAdvertisement {
	var res []*MainAdvertisement
	for adv := range advs.store.Iter() {
		adv.Value.lock.RLock()
		if matches(adv.Value) {
			res = append(res, adv.Value)
		}
		adv.Value.lock.RUnlock()
	}
	return res
}

func (advs *MainAdvertisements) FindAdvertisementsEncoded(gameId string, matches func(adv *MainAdvertisement) bool) []i.A {
	var res []i.A
	advsOriginal := advs.FindAdvertisements(matches)
	for _, adv := range advsOriginal {
		adv.lock.RLock()
		res = append(res, adv.Encode(gameId))
		adv.lock.RUnlock()
	}
	return res
}

func (advs *MainAdvertisements) IsInAdvertisement(user *MainUser) bool {
	return advs.IsHost(user) || advs.IsPeer(user)
}

func (advs *MainAdvertisements) IsPeer(user *MainUser) bool {
	return advs.isPeer(user)
}

func (advs *MainAdvertisements) IsHost(user *MainUser) bool {
	return advs.isHost(user)
}

func (advs *MainAdvertisements) isPeer(user *MainUser) bool {
	return advs.peers.Has(user)
}

func (advs *MainAdvertisements) isHost(user *MainUser) bool {
	return advs.hosts.Has(user)
}

func (advs *MainAdvertisements) addPeer(user *MainUser) {
	if !advs.isPeer(user) {
		advs.peers.Add(user)
	}
}

func (advs *MainAdvertisements) removePeer(user *MainUser) {
	if advs.isPeer(user) {
		advs.peers.Delete(user)
	}
}

func (advs *MainAdvertisements) addHost(user *MainUser) {
	if !advs.isHost(user) {
		advs.hosts.Add(user)
	}
}

func (advs *MainAdvertisements) removeHost(user *MainUser) {
	if advs.isHost(user) {
		advs.hosts.Delete(user)
	}
}
