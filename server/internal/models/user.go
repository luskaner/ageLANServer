package models

import (
	"encoding/binary"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/spf13/viper"
	"hash"
	"hash/fnv"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"time"
)

type MainUser struct {
	id                int32
	statId            int32
	alias             string
	platformUserId    uint64
	profileId         int32
	profileMetadata   string
	profileUintFlag1  uint8
	reliclink         int32
	isXbox            bool
	presence          int8
	presenceLock      *sync.RWMutex
	chatChannels      map[int32]*MainChatChannel
	advertisement     *MainAdvertisement
	advertisementLock *sync.RWMutex
}

type MainUsers struct {
	store      *i.SafeMap[string, *MainUser]
	hasher     hash.Hash64
	hasherLock *sync.Mutex
}

func (users *MainUsers) Initialize() {
	users.store = i.NewSafeMap[string, *MainUser]()
	users.hasher = fnv.New64a()
	users.hasherLock = &sync.Mutex{}
}

func (users *MainUsers) generate(identifier string, isXbox bool, platformUserId uint64, profileMetadata string, profileUIntFlag1 uint8, alias string) *MainUser {
	var seed uint64
	func() {
		users.hasherLock.Lock()
		defer users.hasherLock.Unlock()
		_, _ = users.hasher.Write([]byte(identifier))
		hsh := users.hasher.Sum(nil)
		seed = binary.BigEndian.Uint64(hsh)
		users.hasher.Reset()
	}()
	rng := rand.New(rand.NewPCG(seed, seed))
	return &MainUser{
		id:                rng.Int32(),
		statId:            rng.Int32(),
		profileId:         rng.Int32(),
		profileMetadata:   profileMetadata,
		profileUintFlag1:  profileUIntFlag1,
		reliclink:         rng.Int32(),
		alias:             alias,
		platformUserId:    platformUserId,
		isXbox:            isXbox,
		presenceLock:      &sync.RWMutex{},
		chatChannels:      map[int32]*MainChatChannel{},
		advertisementLock: &sync.RWMutex{},
	}
}

func generatePlatformUserIdSteam(rng *rand.Rand) uint64 {
	Z := rng.Int64N(1 << 31)
	Y := Z % 2
	id := Z*2 + Y + 76561197960265728
	return uint64(id)
}

func generateFullPlatformUserIdXbox(platformUserId int64) string {
	rng := rand.New(rand.NewPCG(uint64(platformUserId), uint64(platformUserId)))
	const chars = "0123456789ABCDEF"
	id := make([]byte, 40)
	for j := range id {
		id[j] = chars[rng.IntN(len(chars))]
	}
	return string(id)
}

func generatePlatformUserIdXbox(rng *rand.Rand) uint64 {
	return uint64(rng.Int64N(9e15) + 1e15)
}

func (users *MainUsers) GetOrCreateUser(gameId string, remoteAddr string, isXbox bool, platformUserId uint64, alias string) *MainUser {
	if viper.GetBool("GeneratePlatformUserId") {
		ipStr, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipV4 := ip.To4()
				if ipV4 != nil {
					seed := uint64(binary.BigEndian.Uint32(ipV4))
					rng := rand.New(rand.NewPCG(seed, seed))
					if isXbox {
						platformUserId = generatePlatformUserIdXbox(rng)
					} else {
						platformUserId = generatePlatformUserIdSteam(rng)
					}
				}
			}
		}
	}
	identifier := getPlatformPath(isXbox, platformUserId)
	mainUser, ok := users.store.Load(identifier)
	if !ok {
		var profileMetadata string
		if gameId == common.GameAoE3 {
			profileMetadata = `{"v":1,"twr":0,"wlr":0,"ai":1,"ac":0}`
		}
		var profileUIntFlag1 uint8
		if gameId == common.GameAoE3 {
			profileUIntFlag1 = 1
		}
		mainUser = users.generate(identifier, isXbox, platformUserId, profileMetadata, profileUIntFlag1, alias)
		users.store.Store(identifier, mainUser)
	}
	return mainUser
}

func (u *MainUser) GetId() int32 {
	return u.id
}

func (u *MainUser) GetStatId() int32 {
	return u.statId
}

func (u *MainUser) GetProfileId() int32 {
	return u.profileId
}

func (u *MainUser) GetReliclink() int32 {
	return u.reliclink
}

func (u *MainUser) GetAlias() string {
	return u.alias
}

func getPlatformPath(isXbox bool, platformUserId uint64) string {
	var prefix string
	var fullPlatformUserId string
	if isXbox {
		fullPlatformUserId = generateFullPlatformUserIdXbox(int64(platformUserId))
		prefix = "xboxlive"
	} else {
		fullPlatformUserId = strconv.FormatUint(platformUserId, 10)
		prefix = "steam"
	}
	return fmt.Sprintf("/%s/%s", prefix, fullPlatformUserId)
}

func (u *MainUser) GetPlatformPath() string {
	return getPlatformPath(u.isXbox, u.platformUserId)
}

func (u *MainUser) GetPlatformId() int {
	var prefix int
	if u.isXbox {
		prefix = 9
	} else {
		prefix = 3
	}
	return prefix
}

func (u *MainUser) GetPlatformUserID() uint64 {
	return u.platformUserId
}

func (users *MainUsers) GetUserByStatId(id int32) (*MainUser, bool) {
	for _, u := range users.store.IterValues() {
		if u.statId == id {
			return u, true
		}
	}
	return nil, false
}

func (users *MainUsers) GetUserById(id int32) (*MainUser, bool) {
	for _, u := range users.store.IterValues() {
		if u.id == id {
			return u, true
		}
	}
	return nil, false
}

func (users *MainUsers) GetUserIds() []int32 {
	userIds := make([]int32, users.store.Len())
	j := 0
	for _, u := range users.store.IterValues() {
		userIds[j] = u.GetId()
		j++
	}
	return userIds
}

func (u *MainUser) GetExtraProfileInfo() i.A {
	return i.A{
		u.statId,
		0,
		0,
		1,
		-1,
		0,
		0,
		-1,
		-1,
		-1,
		-1,
		-1,
		1000,
		// Some time in the past
		1713372625,
		0,
		0,
		0,
	}
}

func (u *MainUser) GetProfileInfo(includePresence bool) i.A {
	var randomTimeDiff int64
	func() {
		i.RngLock.Lock()
		defer i.RngLock.Unlock()
		randomTimeDiff = i.Rng.Int64N(300-50+1) + 50
	}()
	profileInfo := i.A{
		time.Now().UTC().Unix() - randomTimeDiff,
		u.GetId(),
		u.GetPlatformPath(),
		u.GetProfileMetadata(),
		u.GetAlias(),
		"",
		u.GetStatId(),
		u.GetProfileUintFlag1(),
		1,
		u.GetProfileUintFlag2(),
		nil,
		strconv.FormatUint(u.GetPlatformUserID(), 10),
		u.GetPlatformId(),
		i.A{},
	}
	if includePresence {
		profileInfo = append(profileInfo, u.GetPresence(), nil, i.A{})
	}
	return profileInfo
}

func (u *MainUser) GetPresence() int8 {
	u.presenceLock.RLock()
	defer u.presenceLock.RUnlock()
	return u.presence
}

func (u *MainUser) SetPresence(value int8) {
	u.presenceLock.Lock()
	defer u.presenceLock.Unlock()
	u.presence = value
}

func (u *MainUser) GetProfileMetadata() string {
	return u.profileMetadata
}

func (u *MainUser) GetProfileUintFlag1() uint8 {
	return u.profileUintFlag1
}

func (u *MainUser) GetProfileUintFlag2() uint8 {
	var value uint8
	if u.isXbox {
		value = 3
	}
	return value
}

func (u *MainUser) GetAdvertisement() *MainAdvertisement {
	u.advertisementLock.RLock()
	defer u.advertisementLock.RUnlock()
	return u.advertisement
}

func (u *MainUser) JoinChatChannel(channel *MainChatChannel) i.A {
	u.chatChannels[channel.GetId()] = channel
	return channel.AddUser(u)
}

func (u *MainUser) LeaveChatChannel(channel *MainChatChannel) {
	delete(u.chatChannels, channel.GetId())
	channel.RemoveUser(u)
}

func (u *MainUser) GetChannels() []*MainChatChannel {
	channels := make([]*MainChatChannel, len(u.chatChannels))
	j := 0
	for _, channel := range u.chatChannels {
		channels[j] = channel
		j++
	}
	return channels
}

func (u *MainUser) SendChatChannelMessage(channel *MainChatChannel, text string) {
	u.chatChannels[channel.GetId()].AddMessage(u, text)
}

func (u *MainUser) SetAdvertisement(adv *MainAdvertisement) {
	u.advertisementLock.Lock()
	defer u.advertisementLock.Unlock()
	u.advertisement = adv
}

func (users *MainUsers) getUsers() []*MainUser {
	us := make([]*MainUser, users.store.Len())
	j := 0
	for _, u := range users.store.IterValues() {
		us[j] = u
		j++
	}
	return us
}

func (users *MainUsers) GetProfileInfo(includePresence bool, matches func(user *MainUser) bool) []i.A {
	us := users.getUsers()
	var presenceData = make([]i.A, 0)
	for _, u := range us {
		if matches(u) {
			presenceData = append(presenceData, u.GetProfileInfo(includePresence))
		}
	}
	return presenceData
}
