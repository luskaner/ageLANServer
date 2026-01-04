package models

import (
	i "github.com/luskaner/ageLANServer/server/internal"
)

type Message interface {
	GetTime() int64
	GetBroadcast() bool
	GetContent() string
	GetType() uint8
	GetSender() User
	GetReceivers() []User
	GetAdvertisementId() int32
	Encode() i.A
}

type MainMessage struct {
	advertisementId int32
	time            int64
	broadcast       bool
	content         string
	typ             uint8
	sender          User
	receivers       []User
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

func (message *MainMessage) GetSender() User {
	return message.sender
}

func (message *MainMessage) GetReceivers() []User {
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
