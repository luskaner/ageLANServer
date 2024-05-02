package user

import (
	"aoe2DELanServer/j"
	"aoe2DELanServer/keyLock"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
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
var Lock = keyLock.NewKeyMutex()
var presenceLock = keyLock.NewKeyRWMutex()

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

func GetOrCreate(isXbox bool, platformUserId uint64, alias string) *User {
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
	if isXbox {
		prefix = "xboxlive"
	} else {
		prefix = "steam"
	}
	return fmt.Sprintf("/%s/%d", prefix, platformUserId)
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

func GetByStatId(id int32) (*User, bool) {
	user, ok := userStatIdToUserMap[id]
	return user, ok
}

func GetById(id int32) (*User, bool) {
	user, ok := userIdToUserMap[id]
	return user, ok
}

func (u *User) GetExtraProfileInfo() j.A {
	return j.A{
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

func (u *User) GetProfileInfo(includePresence bool) j.A {
	/*isSteamInt := 1
	if u.isXbox {
		isSteamInt = 0
	}*/
	profileInfo := j.A{
		time.Now().UTC().Unix() - rand.Int63n(300-50+1) + 50,
		u.id,
		u.GetPlatformPath(),
		"",
		u.alias,
		"",
		u.statId,
		//isSteamInt,
		1,
		1,
		0,
		nil,
		strconv.FormatUint(u.platformUserId, 10),
		u.GetPlatformId(),
		j.A{},
	}
	if includePresence {
		profileInfo = append(profileInfo, j.A{u.GetPresence(), nil, j.A{}}...)
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
	i := 0
	for _, u := range userStore {
		users[i] = u
		i++
	}
	return users
}

func GetProfileInfo(includePresence bool, matches func(user *User) bool) []j.A {
	users := getUsers()
	var presenceData = make([]j.A, 0)
	for _, u := range users {
		if matches(u) {
			presenceData = append(presenceData, u.GetProfileInfo(includePresence))
		}
	}
	return presenceData
}
