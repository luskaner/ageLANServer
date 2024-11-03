package models

import (
	i "github.com/luskaner/ageLANServer/server/internal"
)

type MainMessage struct {
	advertisementId int32
	time            int64
	broadcast       bool
	content         string
	typ             uint8
	sender          *MainUser
	receivers       []*MainUser
}

func (message *MainMessage) GetTime() int64 {
	return message.time
}

func (message *MainMessage) GetBroadcast() bool {
	return message.broadcast
}

func (message *MainMessage) GetContent() string {
	return message.content
}

func (message *MainMessage) GetType() uint8 {
	return message.typ
}

func (message *MainMessage) GetSender() *MainUser {
	return message.sender
}

func (message *MainMessage) GetReceivers() []*MainUser {
	return message.receivers
}

func (message *MainMessage) GetAdvertisementId() int32 {
	return message.advertisementId
}

func (message *MainMessage) Encode() i.A {
	return i.A{
		message.sender.GetId(),
		message.content,
		message.content,
		message.typ,
		message.advertisementId,
	}
}
