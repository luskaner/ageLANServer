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
	"github.com/spf13/viper"
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
	GetExtraProfileInfo(clientLibVersion uint16) i.A
	GetProfileInfo(includePresence bool, clientLibVersion uint16) i.A
	GetPresence() int32
	SetPresence(presence int32)
	GetProfileMetadata() string
	GetProfileUintFlag1() uint8
	GetProfileUintFlag2() uint8
	GetAvatarStats() *PersistentJsonData[*AvatarStats]
	EncodeAvatarStats() i.A
}

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
	// Dynamic from here
	presence    atomic.Int32
	avatarStats *PersistentJsonData[*AvatarStats]
}

// EncodeAvatarStats must only be called in the same goroutine it is created in
func (u *MainUser) EncodeAvatarStats() i.A {
	return u.GetAvatarStats().Data().Encode(u.GetProfileId())
}

func newAvatarStats(values map[int32]int64) *AvatarStats {
	avatarStats := &AvatarStats{
		values: i.NewSafeMap[int32, AvatarStat](),
		locks:  i.NewKeyRWMutex[int32](),
	}
	for id, value := range values {
		avatarStats.values.Store(id, AvatarStat{
			Id:          id,
			Value:       value,
			LastUpdated: time.Now().UTC(),
		}, func(stored AvatarStat) bool {
			return true
		})
	}
	return avatarStats
}

type Users interface {
	Initialize()
	GetOrCreateUser(gameId string, avatarStatsDefinitions AvatarStatDefinitions, remoteAddr string, remoteMacAddress string, isXbox bool, platformUserId uint64, alias string) User
	GetUserByStatId(id int32) (User, bool)
	GetUserById(id int32) (User, bool)
	GetUserIds() func(func(int32) bool)
	GetProfileInfo(includePresence bool, matches func(user User) bool, clientLibVersion uint16) []i.A
	GetUserByPlatformUserId(xbox bool, id uint64) (User, bool)
}

type MainUsers struct {
	store      *i.SafeMap[string, User]
	GenerateFn func(gameId string, avatarStatsDefinitions AvatarStatDefinitions, identifier string, isXbox bool, platformUserId uint64, profileMetadata string, profileUIntFlag1 uint8, alias string) User
}

func (users *MainUsers) Initialize() {
	users.store = i.NewSafeMap[string, User]()
	if users.GenerateFn == nil {
		users.GenerateFn = users.Generate
	}
}

func (users *MainUsers) Generate(gameId string, avatarStatsDefinitions AvatarStatDefinitions, identifier string, isXbox bool, platformUserId uint64, profileMetadata string, profileUIntFlag1 uint8, alias string) User {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(identifier))
	hsh := hasher.Sum(nil)
	seed := binary.BigEndian.Uint64(hsh)
	rng := rand.New(rand.NewPCG(seed, seed))
	avatarStats, _ := NewPersistentJsonData[*AvatarStats](
		UserDataPath(gameId, !isXbox, strconv.FormatUint(platformUserId, 10), "avatarStats"),
		func() *AvatarStats {
			var values map[string]int64
			switch gameId {
			case common.GameAoE2:
				// TODO: Remove the ones not needed
				values = map[string]int64{
					"STAT_NUM_MVP_AWARDS":           0,
					"STAT_HIGHEST_SCORE_TOTAL":      0,
					"STAT_HIGHEST_SCORE_ECONOMIC":   0,
					"STAT_HIGHEST_SCORE_TECHNOLOGY": 0,
					"STAT_CAREER_UNITS_KILLED":      0,
					"STAT_CAREER_UNITS_LOST":        0,
					"STAT_CAREER_UNITS_CONVERTED":   0,
					"STAT_CAREER_BUILDINGS_RAZED":   0,
					"STAT_CAREER_BUILDINGS_LOST":    0,
					"STAT_CAREER_NUM_CASTLES":       0,
					"STAT_GAMES_PLAYED_ONLINE":      0,
					"STAT_ELO_XRM_WINS":             0,
					"STAT_POP_CAP_200_MP":           0,
					"STAT_POP_PEAK_200_MP":          0,
					"STAT_TOTAL_GAMES":              0,
				}
			case common.GameAoE3:
				// FIXME: Is this even needed?
				values = map[string]int64{
					"STAT_EVENT_EXPLORER_SKIN_CHALLENGE_14c": 16,
				}
			case common.GameAoM:
				values = map[string]int64{
					"STAT_GAUNTLET_REWARD_XP":     2_147_483_647,
					"STAT_GAUNTLET_REWARD_FAVOUR": 19_500,
				}
			default:
				values = map[string]int64{}
			}
			intValues := make(map[int32]int64, len(values))
			for k, v := range values {
				if id, ok := avatarStatsDefinitions.GetIdByName(k); ok {
					intValues[id] = v
				}
			}
			return newAvatarStats(intValues)
		},
	)
	return &MainUser{
		id:               rng.Int32(),
		statId:           rng.Int32(),
		profileId:        rng.Int32(),
		profileMetadata:  profileMetadata,
		profileUintFlag1: profileUIntFlag1,
		reliclink:        rng.Int32(),
		alias:            alias,
		platformUserId:   platformUserId,
		isXbox:           isXbox,
		avatarStats:      avatarStats,
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

func (users *MainUsers) GetOrCreateUser(gameId string, avatarStatsDefinitions AvatarStatDefinitions, remoteAddr string, remoteMacAddress string, isXbox bool, platformUserId uint64, alias string) User {
	if viper.GetBool("GeneratePlatformUserId") {
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
	var profileMetadata string
	var profileUIntFlag1 uint8
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
		profileMetadata = `{"v":1,"twr":0,"wlr":0,"ai":1,"ac":0}`
		profileUIntFlag1 = 1
	}
	newUser := users.GenerateFn(gameId, avatarStatsDefinitions, identifier, isXbox, platformUserId, profileMetadata, profileUIntFlag1, alias)
	mainUser, _ := users.store.LoadOrStore(identifier, newUser)
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

func (users *MainUsers) GetProfileInfo(includePresence bool, matches func(user User) bool, clientLibVersion uint16) []i.A {
	var presenceData = make([]i.A, 0)
	for u := range users.store.Values() {
		if matches(u) {
			presenceData = append(presenceData, u.GetProfileInfo(includePresence, clientLibVersion))
		}
	}
	return presenceData
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

func (u *MainUser) GetExtraProfileInfo(clientLibVersion uint16) i.A {
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

func (u *MainUser) GetProfileInfo(includePresence bool, clientLibVersion uint16) i.A {
	profileInfo := i.A{
		time.Date(2024, 5, 2, 3, 34, 0, 0, time.UTC).Unix(),
		u.GetId(),
		u.GetPlatformPath(),
		u.GetProfileMetadata(),
		u.GetAlias(),
	}
	if clientLibVersion >= 190 {
		profileInfo = append(profileInfo, u.GetAlias())
	}
	profileInfo = append(
		profileInfo,
		"",
		u.GetStatId(),
		u.GetProfileUintFlag1(),
		1,
		u.GetProfileUintFlag2(),
		nil,
		strconv.FormatUint(u.GetPlatformUserID(), 10),
		u.GetPlatformId(),
		i.A{},
	)
	if includePresence {
		profileInfo = append(profileInfo, u.GetPresence(), nil, i.A{})
	}
	return profileInfo
}

func (u *MainUser) GetPresence() int32 {
	return u.presence.Load()
}

func (u *MainUser) SetPresence(presence int32) {
	u.presence.Store(presence)
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
