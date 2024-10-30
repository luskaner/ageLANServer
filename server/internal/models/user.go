package models

import (
	"encoding/binary"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/spf13/viper"
	"hash"
	"hash/fnv"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type MainUser struct {
	id               int32
	statId           int32
	alias            string
	platformUserId   uint64
	profileId        int32
	profileMetadata  string
	profileUintFlag1 uint8
	reliclink        int32
	isXbox           bool
	presence         int8
	lock             *sync.Mutex
	chatChannels     map[int32]*MainChatChannel
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
	users.hasherLock.Lock()
	_, _ = users.hasher.Write([]byte(identifier))
	hsh := users.hasher.Sum(nil)
	seed := binary.BigEndian.Uint64(hsh)
	users.hasher.Reset()
	users.hasherLock.Unlock()
	rng := rand.New(rand.NewSource(int64(seed)))
	return &MainUser{
		id:               rng.Int31(),
		statId:           rng.Int31(),
		profileId:        rng.Int31(),
		profileMetadata:  profileMetadata,
		profileUintFlag1: profileUIntFlag1,
		reliclink:        rng.Int31(),
		alias:            alias,
		platformUserId:   platformUserId,
		isXbox:           isXbox,
		lock:             &sync.Mutex{},
		chatChannels:     map[int32]*MainChatChannel{},
	}
}

func generatePlatformUserIdSteam(rng *rand.Rand) uint64 {
	Z := rng.Int63n(1 << 31)
	Y := Z % 2
	id := Z*2 + Y + 76561197960265728
	return uint64(id)
}

func generateFullPlatformUserIdXbox(platformUserId int64) string {
	rng := rand.New(rand.NewSource(platformUserId))
	const chars = "0123456789ABCDEF"
	id := make([]byte, 40)
	for j := range id {
		id[j] = chars[rng.Intn(len(chars))]
	}
	return string(id)
}

func generatePlatformUserIdXbox(rng *rand.Rand) uint64 {
	return uint64(rng.Int63n(9e15) + 1e15)
}

func (users *MainUsers) GetOrCreateUser(gameId string, remoteAddr string, isXbox bool, platformUserId uint64, alias string) *MainUser {
	if viper.GetBool("GeneratePlatformUserId") {
		ipStr, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipV4 := ip.To4()
				if ipV4 != nil {
					rng := rand.New(rand.NewSource(int64(binary.BigEndian.Uint32(ipV4))))
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
		if gameId != common.GameAoE3 {
			profileUIntFlag1 = 0
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
	for u := range users.store.Iter() {
		if u.Value.statId == id {
			return u.Value, true
		}
	}
	return nil, false
}

func (users *MainUsers) GetUserById(id int32) (*MainUser, bool) {
	for u := range users.store.Iter() {
		if u.Value.id == id {
			return u.Value, true
		}
	}
	return nil, false
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
	i.RngLock.Lock()
	randomTimeDiff := i.Rng.Int63n(300-50+1) + 50
	i.RngLock.Unlock()
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
		0,
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
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.presence
}

func (u *MainUser) SetPresence(value int8) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.presence = value
}

func (u *MainUser) GetProfileMetadata() string {
	return u.profileMetadata
}

func (u *MainUser) GetProfileUintFlag1() uint8 {
	return u.profileUintFlag1
}

func (u *MainUser) JoinChatChannel(channel *MainChatChannel) {
	u.chatChannels[channel.GetId()] = channel
	channel.AddUser(u)
}

func (u *MainUser) LeaveChatChannel(channel *MainChatChannel) {
	delete(u.chatChannels, channel.GetId())
	channel.RemoveUser(u)
}

func (u *MainUser) LeaveAllChannels() {
	for _, channel := range u.chatChannels {
		channel.RemoveUser(u)
	}
	u.chatChannels = map[int32]*MainChatChannel{}
}

func (u *MainUser) SendChatChannelMessage(channel *MainChatChannel, text string) {
	u.chatChannels[channel.GetId()].AddMessage(u, text)
}

func (users *MainUsers) getUsers() []*MainUser {
	us := make([]*MainUser, users.store.Len())
	j := 0
	for u := range users.store.Iter() {
		us[j] = u.Value
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
