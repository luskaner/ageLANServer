package models

import (
	"encoding/binary"
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/spf13/viper"
	"hash/fnv"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type User struct {
	id             int32
	statId         int32
	alias          string
	platformUserId uint64
	profileId      int32
	reliclink      int32
	isXbox         bool
}

var presenceStore = make(map[*User]int8)
var userStore = make(map[string]*User)
var userIdToUserMap = make(map[int32]*User)
var userStatIdToUserMap = make(map[int32]*User)
var hasher = fnv.New64a()
var Lock = i.NewKeyMutex()
var presenceLock = i.NewKeyRWMutex()

func generate(identifier string, isXbox bool, platformUserId uint64, alias string) *User {
	_, _ = hasher.Write([]byte(identifier))
	hash := hasher.Sum(nil)
	seed := binary.BigEndian.Uint64(hash)
	hasher.Reset()
	rng := rand.New(rand.NewSource(int64(seed)))
	return &User{
		id:             rng.Int31(),
		statId:         rng.Int31(),
		profileId:      rng.Int31(),
		reliclink:      rng.Int31(),
		alias:          alias,
		platformUserId: platformUserId,
		isXbox:         isXbox,
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

func GetOrCreateUser(remoteAddr string, isXbox bool, platformUserId uint64, alias string) *User {
	if viper.GetBool("default.GeneratePlatformUserId") {
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
	Lock.Lock(identifier)
	user, ok := userStore[identifier]
	if !ok {
		user = generate(identifier, isXbox, platformUserId, alias)
		userIdToUserMap[user.id] = user
		userStatIdToUserMap[user.statId] = user
		userStore[identifier] = user
	}
	Lock.Unlock(identifier)
	return user
}

func (u *User) GetId() int32 {
	return u.id
}

func (u *User) GetStatId() int32 {
	return u.statId
}

func (u *User) GetProfileId() int32 {
	return u.profileId
}

func (u *User) GetReliclink() int32 {
	return u.reliclink
}

func (u *User) GetAlias() string {
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

func (u *User) GetPlatformPath() string {
	return getPlatformPath(u.isXbox, u.platformUserId)
}

func (u *User) GetPlatformId() int {
	var prefix int
	if u.isXbox {
		prefix = 9
	} else {
		prefix = 3
	}
	return prefix
}

func (u *User) GetPlatformUserID() uint64 {
	return u.platformUserId
}

func GetUserByStatId(id int32) (*User, bool) {
	user, ok := userStatIdToUserMap[id]
	return user, ok
}

func GetUserById(id int32) (*User, bool) {
	user, ok := userIdToUserMap[id]
	return user, ok
}

func (u *User) GetExtraProfileInfo() i.A {
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

func (u *User) GetProfileInfo(includePresence bool) i.A {
	profileInfo := i.A{
		time.Now().UTC().Unix() - i.Rng.Int63n(300-50+1) + 50,
		u.id,
		u.GetPlatformPath(),
		"",
		u.alias,
		"",
		u.statId,
		1,
		1,
		0,
		nil,
		strconv.FormatUint(u.platformUserId, 10),
		u.GetPlatformId(),
		i.A{},
	}
	if includePresence {
		profileInfo = append(profileInfo, i.A{u.GetPresence(), nil, i.A{}}...)
	}
	return profileInfo
}

func (u *User) SetPresence(value int8) {
	presenceLock.Lock(u)
	presenceStore[u] = value
	presenceLock.Unlock(u)
}

func (u *User) GetPresence() int8 {
	presenceLock.RLock(u)
	defer presenceLock.RUnlock(u)
	return presenceStore[u]
}

func getUsers() []*User {
	users := make([]*User, len(userStore))
	j := 0
	for _, u := range userStore {
		users[j] = u
		j++
	}
	return users
}

func GetProfileInfo(includePresence bool, matches func(user *User) bool) []i.A {
	users := getUsers()
	var presenceData = make([]i.A, 0)
	for _, u := range users {
		if matches(u) {
			presenceData = append(presenceData, u.GetProfileInfo(includePresence))
		}
	}
	return presenceData
}
