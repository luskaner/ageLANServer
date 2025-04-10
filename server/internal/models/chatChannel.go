package models

import (
	"github.com/luskaner/ageLANServer/server/internal"
	"iter"
	"strconv"
)

type MainChatChannel struct {
	Id    int32
	Name  string
	users *internal.SafeOrderedMap[int32, *MainUser]
}

func (channel *MainChatChannel) GetId() int32 {
	return channel.Id
}

func (channel *MainChatChannel) encode() internal.A {
	return internal.A{
		channel.Id,
		channel.Name,
		"",
		channel.users.Len(),
	}
}

func (channel *MainChatChannel) GetUsers() iter.Seq[*MainUser] {
	return func(yield func(user *MainUser) bool) {
		_, users := channel.users.Values()
		for v := range users {
			if !yield(v) {
				return
			}
		}
	}
}

func (channel *MainChatChannel) AddUser(user *MainUser) (exists bool, encodedUsers internal.A) {
	exists, _ = channel.users.IterAndStore(user.GetId(), user, nil, func(length int, users iter.Seq2[int32, *MainUser]) {
		i := 0
		encodedUsers = make(internal.A, length)
		for _, el := range users {
			encodedUsers[i] = internal.A{0, el.GetProfileInfo(false)}
			i++
		}
	})
	return
}

func (channel *MainChatChannel) RemoveUser(user *MainUser) bool {
	return channel.users.Delete(user.GetId())
}

func (channel *MainChatChannel) HasUser(user *MainUser) bool {
	_, ok := channel.users.Load(user.GetId())
	return ok
}

type MainChatChannels struct {
	index *internal.ReadOnlyOrderedMap[int32, *MainChatChannel]
}

func (channels *MainChatChannels) Initialize(chatChannels map[string]MainChatChannel) {
	keys := make([]int32, len(chatChannels))
	values := make(map[int32]*MainChatChannel, len(chatChannels))
	j := 0
	for id, channel := range chatChannels {
		idInt, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			panic(err)
		}
		channel.users = internal.NewSafeOrderedMap[int32, *MainUser]()
		channel.Id = int32(idInt)
		keys[j] = channel.Id
		values[channel.Id] = &channel
		j++
	}
	channels.index = internal.NewReadOnlyOrderedMap[int32, *MainChatChannel](keys, values)
}

func (channels *MainChatChannels) Encode() internal.A {
	c := make(internal.A, channels.index.Len())
	i := 0
	for el := range channels.index.Values() {
		c[i] = el.encode()
		i++
	}
	return c
}

func (channels *MainChatChannels) GetById(id int32) (*MainChatChannel, bool) {
	return channels.index.Load(id)
}

func (channels *MainChatChannels) Iter() iter.Seq2[int32, *MainChatChannel] {
	return channels.index.Iter()
}
