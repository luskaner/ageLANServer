package models

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
)

type User interface {
	GetId() int32
	GetXbox() bool
	GetStatId() int32
	GetProfileId() int32
	GetReliclink() int32
	GetAlias() string
	GetPlatformPath() string
	GetPlatformId() int
	GetPlatformUserID() uint64
	EncodeExtraProfileInfo(clientLibVersion uint16) i.A
	EncodeProfileInfo(clientLibVersion uint16) i.A
	GetPresence() int32
	SetPresence(presence int32)
	SetPresenceProperty(id int32, value string)
	GetAvatarMetadata() *PersistentJsonData[*string]
	GetProfileProperties() *PersistentJsonData[*map[string]string]
	GetProfileExperience() uint32
	GetProfileLevel() uint16
	GetPlatformRelated() uint8
	GetAvatarStats() *PersistentJsonData[*AvatarStats]
	GetPersistentData() *PersistentStringJsonMap
	GetItems() *PersistentJsonData[*map[int32]Item]
	GetItemLoadouts() *PersistentJsonData[ItemLoadouts]
	EncodeAvatarStats() i.A
	EncodePresence(definitions PresenceDefinitions) i.A
}

type MainUser struct {
	id                int32
	statId            int32
	alias             string
	platformUserId    uint64
	profileId         int32
	profileExperience uint32
	reliclink         int32
	isXbox            bool
	persistentData    *PersistentStringJsonMap
	// Dynamic from here
	profileProperties  *PersistentJsonData[*map[string]string]
	avatarMetadata     *PersistentJsonData[*string]
	presence           atomic.Int32
	presenceProperties *i.SafeMap[int32, string]
	avatarStats        *PersistentJsonData[*AvatarStats]
	items              *PersistentJsonData[*map[int32]Item]
	itemLoadouts       *PersistentJsonData[ItemLoadouts]
}

func (u *MainUser) EncodeAvatarStats() i.A {
	var result i.A
	_ = u.GetAvatarStats().WithReadOnly(func(data *AvatarStats) error {
		result = data.Encode(u.GetProfileId())
		return nil
	})
	return result
}

type Users interface {
	Initialize()
	GetOrCreateUser(gameId string, itemDefinitions Items, avatarStatsDefinitions AvatarStatDefinitions, remoteAddr string, remoteMacAddress string, isXbox bool, platformUserId uint64, alias string) User
	GetUserByStatId(id int32) (User, bool)
	GetUserById(id int32) (User, bool)
	GetUserIds() func(func(int32) bool)
	EncodeProfileInfo(definitions PresenceDefinitions, matches func(user User) bool, clientLibVersion uint16) []i.A
	GetUserByPlatformUserId(xbox bool, id uint64) (User, bool)
}

type MainUsers struct {
	store      *i.SafeMap[string, User]
	GenerateFn func(
		gameId string,
		persistentData *PersistentStringJsonMap,
		itemDefinitions Items,
		avatarStatsDefinitions AvatarStatDefinitions,
		identifier string,
		isXbox bool,
		platformUserId uint64,
		alias string,
	) User
}

func (users *MainUsers) Initialize() {
	users.store = i.NewSafeMap[string, User]()
	if users.GenerateFn == nil {
		users.GenerateFn = users.Generate
	}
}

func (users *MainUsers) Generate(gameId string, persistentData *PersistentStringJsonMap, itemDefinitions Items, avatarStatsDefinitions AvatarStatDefinitions, identifier string, isXbox bool, platformUserId uint64, alias string) User {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(identifier))
	hsh := hasher.Sum(nil)
	seed := binary.BigEndian.Uint64(hsh)
	rng := rand.New(rand.NewPCG(seed, seed))
	var avatarStats *PersistentJsonData[*AvatarStats]
	if avatarStatsDefinitions != nil {
		avatarStats, _ = NewPersistentJsonData[*AvatarStats](
			persistentData,
			"avatarStats",
			NewAvatarStatsUpgradableDefaultData(gameId, avatarStatsDefinitions),
		)
	}
	var profileProperties *PersistentJsonData[*map[string]string]
	if gameId == common.GameAoE3 || gameId == common.GameAoE4 || gameId == common.GameAoM {
		profileProperties, _ = NewPersistentJsonData[*map[string]string](
			persistentData,
			"profileProperties",
			NewProfilePropertiesUpgradableDefaultData(),
		)
	}
	var items *PersistentJsonData[*map[int32]Item]
	if itemDefinitions != nil {
		items, _ = NewPersistentJsonData[*map[int32]Item](
			persistentData,
			"items",
			NewItemsUpgradableDefaultData(gameId, itemDefinitions),
		)
	}
	var itemLoadouts *PersistentJsonData[ItemLoadouts]
	if gameId != common.GameAoE1 {
		itemLoadouts, _ = NewPersistentJsonData[ItemLoadouts](
			persistentData,
			"itemLoadouts",
			NewItemLoadoutsUpgradableDefaultData(),
		)
	}
	avatarMetadata, _ := NewPersistentJsonData[*string](
		persistentData,
		"avatarMetadata",
		NewAvatarMetadataUpgradableDefaultData(gameId),
	)
	var presenceProperties *i.SafeMap[int32, string]
	if gameId != common.GameAoE1 {
		presenceProperties = i.NewSafeMap[int32, string]()
	}
	return &MainUser{
		id:                 rng.Int32(),
		statId:             rng.Int32(),
		profileId:          rng.Int32(),
		avatarMetadata:     avatarMetadata,
		items:              items,
		itemLoadouts:       itemLoadouts,
		profileProperties:  profileProperties,
		reliclink:          rng.Int32(),
		alias:              alias,
		platformUserId:     platformUserId,
		isXbox:             isXbox,
		avatarStats:        avatarStats,
		persistentData:     persistentData,
		presenceProperties: presenceProperties,
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

func (users *MainUsers) GetOrCreateUser(gameId string, itemDefinitions Items, avatarStatsDefinitions AvatarStatDefinitions, remoteAddr string, remoteMacAddress string, isXbox bool, platformUserId uint64, alias string) User {
	if i.GeneratePlatformUserId {
		entropy := make([]byte, 16)
		macAddress, err := net.ParseMAC(remoteMacAddress)
		if err == nil {
			copy(entropy[:6], macAddress)
		}
		ipStr, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			ip := net.ParseIP(ipStr)
			if ip != nil {
				ipV4 := ip.To4()
				if ipV4 != nil {
					copy(entropy[6:10], ipV4)
				}
			}
		}
		sizeAlias := min(len(alias), 6)
		copy(entropy[10:10+sizeAlias], alias[:sizeAlias])
		seed1 := binary.BigEndian.Uint64(entropy[:8])
		seed2 := binary.BigEndian.Uint64(entropy[8:])
		rng := rand.New(rand.NewPCG(seed1, seed2))
		if isXbox {
			platformUserId = generatePlatformUserIdXbox(rng)
		} else {
			platformUserId = generatePlatformUserIdSteam(rng)
		}
	}
	identifier := getPlatformPath(isXbox, platformUserId)
	mainUser, _ := users.store.LoadOrStoreFn(
		identifier,
		func() User {
			persistentData, _ := NewPersistentStringMap(
				UserDataPath(gameId, !isXbox, strconv.FormatUint(platformUserId, 10)),
				&InitialUpgradableData[*PersistentStringJsonMapRaw]{},
			)
			return users.GenerateFn(
				gameId,
				persistentData,
				itemDefinitions,
				avatarStatsDefinitions,
				identifier,
				isXbox,
				platformUserId,
				alias,
			)
		},
	)
	return mainUser
}

func (users *MainUsers) GetUserByStatId(id int32) (User, bool) {
	return users.getFirst(func(u User) bool { return u.GetStatId() == id })
}

func (users *MainUsers) GetUserById(id int32) (User, bool) {
	return users.getFirst(func(u User) bool { return u.GetId() == id })
}

func (users *MainUsers) GetUserByPlatformUserId(xbox bool, id uint64) (User, bool) {
	return users.getFirst(func(u User) bool { return u.GetXbox() == xbox && u.GetPlatformUserID() == id })
}

func (users *MainUsers) getFirst(fn func(u User) bool) (User, bool) {
	for u := range users.store.Values() {
		if fn(u) {
			return u, true
		}
	}
	return nil, false
}

func (users *MainUsers) GetUserIds() func(func(int32) bool) {
	return func(yield func(int32) bool) {
		for u := range users.store.Values() {
			if !yield(u.GetId()) {
				return
			}
		}
	}
}

func (users *MainUsers) EncodeProfileInfo(presenceDefinitions PresenceDefinitions, matches func(user User) bool, clientLibVersion uint16) []i.A {
	var profileInfo = make([]i.A, 0)
	for u := range users.store.Values() {
		if matches(u) {
			currentProfileInfo := u.EncodeProfileInfo(clientLibVersion)
			if presenceDefinitions != nil {
				currentProfileInfo = append(currentProfileInfo, u.EncodePresence(presenceDefinitions)...)
			}
			profileInfo = append(profileInfo, currentProfileInfo)
		}
	}
	return profileInfo
}

func (u *MainUser) GetPersistentData() *PersistentStringJsonMap {
	return u.persistentData
}

func (u *MainUser) GetAvatarStats() *PersistentJsonData[*AvatarStats] {
	return u.avatarStats
}

func (u *MainUser) GetId() int32 {
	return u.id
}

func (u *MainUser) GetXbox() bool {
	return u.isXbox
}

func (u *MainUser) GetStatId() int32 {
	return u.statId
}

func (u *MainUser) GetProfileId() int32 {
	return u.profileId
}

func (u *MainUser) GetProfileProperties() *PersistentJsonData[*map[string]string] {
	return u.profileProperties
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

func (u *MainUser) EncodeExtraProfileInfo(clientLibVersion uint16) i.A {
	info := i.A{
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
	if clientLibVersion >= 190 {
		info = append(info, 0, 0)
	}
	return info
}

func (u *MainUser) EncodeProfileInfo(clientLibVersion uint16) i.A {
	profileInfo := i.A{
		time.Date(2024, 5, 2, 3, 34, 0, 0, time.UTC).Unix(),
		u.GetId(),
		u.GetPlatformPath(),
		u.GetAvatarMetadata(),
		u.GetAlias(),
	}
	if clientLibVersion >= 190 {
		profileInfo = append(profileInfo, u.GetAlias())
	}
	profileInfo = append(
		profileInfo,
		"",
		u.GetStatId(),
		u.GetProfileExperience(),
		u.GetProfileLevel(),
		u.GetPlatformRelated(),
		nil,
		strconv.FormatUint(u.GetPlatformUserID(), 10),
		u.GetPlatformId(),
		i.A{},
	)
	return profileInfo
}

func (u *MainUser) EncodePresence(definitions PresenceDefinitions) i.A {
	var presenceProperties i.A
	for id, value := range u.presenceProperties.Iter() {
		presenceProperties = append(presenceProperties, i.A{id, value})
	}
	presenceId := u.GetPresence()
	return i.A{
		presenceId,
		(*definitions.Get(presenceId)).GetLabel(),
		presenceProperties,
	}
}

func (u *MainUser) GetPresence() int32 {
	return u.presence.Load()
}

func (u *MainUser) SetPresence(presence int32) {
	u.presence.Store(presence)
}

func (u *MainUser) SetPresenceProperty(id int32, value string) {
	if value == "" {
		u.presenceProperties.Delete(id)
	} else {
		u.presenceProperties.Store(id, value, func(_ string) bool {
			return true
		})
	}
}

func (u *MainUser) GetAvatarMetadata() *PersistentJsonData[*string] {
	return u.avatarMetadata
}

func (u *MainUser) GetPlatformRelated() uint8 {
	var value uint8
	if u.isXbox {
		value = 3
	}
	return value
}

func (u *MainUser) GetProfileLevel() uint16 {
	return 9_999
}

func (u *MainUser) GetProfileExperience() uint32 {
	return 0
}

func (u *MainUser) GetItems() *PersistentJsonData[*map[int32]Item] {
	return u.items
}

func (u *MainUser) GetItemLoadouts() *PersistentJsonData[ItemLoadouts] {
	return u.itemLoadouts
}
