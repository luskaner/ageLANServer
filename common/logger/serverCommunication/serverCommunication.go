package serverCommunication

import "time"

const (
	MessageRequest = "request"
	MessageWSS     = "wss"
)

type MessageType struct {
	Type string
}

type Uptime struct {
	Uptime time.Duration
}

type Sender struct {
	Sender string
}

type Body struct {
	Body []byte
}

type BodyHash struct {
	BodyHash [64]byte
}
