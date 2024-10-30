package models

import (
	"github.com/luskaner/aoe2DELanServer/server/internal"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"strconv"
	"sync"
)

type MainChatChannelMessage struct {
	sender *MainUser
	text   string
}

type MainChatChannel struct {
	Id           int32
	Name         string
	users        *orderedmap.OrderedMap[int32, *MainUser]
	usersLock    *sync.RWMutex
	messages     []MainChatChannelMessage
	messagesLock *sync.RWMutex
}

func (channel *MainChatChannel) GetId() int32 {
	return channel.Id
}

func (channel *MainChatChannel) GetName() string {
	return channel.Name
}

func (channel *MainChatChannel) Encode() internal.A {
	channel.usersLock.RLock()
	defer channel.usersLock.RUnlock()
	return internal.A{
		channel.Id,
		channel.Name,
		"",
		channel.users.Len(),
	}
}

func (channel *MainChatChannel) EncodeUsers() internal.A {
	c := make(internal.A, 0, channel.users.Len())
	i := 0
	channel.usersLock.RLock()
	for el := channel.users.Oldest(); el != nil; el = el.Next() {
		c[i] = internal.A{0, el.Value.GetProfileInfo(false)}
		i++
	}
	channel.usersLock.RUnlock()
	return c
}

func (channel *MainChatChannel) GetUsers() []*MainUser {
	channel.usersLock.RLock()
	notifyUsers := make([]*MainUser, 0, channel.users.Len())
	for el := channel.users.Oldest(); el != nil; el = el.Next() {
		notifyUsers = append(notifyUsers, el.Value)
	}
	channel.usersLock.RUnlock()
	return notifyUsers
}

func (channel *MainChatChannel) AddUser(user *MainUser) {
	channel.usersLock.Lock()
	channel.users.Set(user.GetId(), user)
	channel.usersLock.Unlock()
}

func (channel *MainChatChannel) RemoveUser(user *MainUser) {
	channel.usersLock.Lock()
	defer channel.usersLock.Unlock()
	channel.users.Delete(user.GetId())
}

func (channel *MainChatChannel) HasUser(user *MainUser) bool {
	channel.usersLock.RLock()
	_, ok := channel.users.Load(user.GetId())
	channel.usersLock.RUnlock()
	return ok
}

func (channel *MainChatChannel) AddMessage(sender *MainUser, text string) {
	channel.messagesLock.Lock()
	defer channel.messagesLock.Unlock()
	channel.messages = append(channel.messages, MainChatChannelMessage{sender, text})
}

type MainChatChannels struct {
	index *orderedmap.OrderedMap[int32, *MainChatChannel]
}

func (channels *MainChatChannels) Initialize(chatChannels map[string]MainChatChannel) {
	channels.index = orderedmap.New[int32, *MainChatChannel]()
	for id, channel := range chatChannels {
		idInt, _ := strconv.Atoi(id)
		channel.users = orderedmap.New[int32, *MainUser]()
		channel.usersLock = &sync.RWMutex{}
		channel.messagesLock = &sync.RWMutex{}
		channel.Id = int32(idInt)
		channels.index.Set(channel.Id, &channel)
	}
}

func (channels *MainChatChannels) Encode() internal.A {
	c := make(internal.A, channels.index.Len())
	i := 0
	for el := channels.index.Oldest(); el != nil; el = el.Next() {
		c[i] = el.Value.Encode()
		i++
	}
	return c
}

func (channels *MainChatChannels) GetById(id int32) (*MainChatChannel, bool) {
	return channels.index.Get(id)
}
