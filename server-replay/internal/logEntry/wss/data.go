package wss

import (
	"encoding/json"
	"log"
	"net"

	"github.com/gorilla/websocket"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/wss"
	"github.com/luskaner/ageLANServer/server-replay/internal/logEntry"
)

type Data struct {
	Websocket[wss.Data]
	messageResponse
	response bool
}

func (d *Data) CheckResponse() {
	if d.response {
		if d.messageResponse.err != nil {
			log.Println(d.messageResponse.err)
			return
		}
		if d.messageResponse.messageType != websocket.TextMessage {
			log.Printf("Not a text message")
			return
		}
		if len(d.data.Body.Body) == 0 {
			if !logEntry.SameBody(d.data.BodyHash.BodyHash, d.messageResponse.body) {
				log.Println("Body hash does not match")
				return
			}
		} else {
			var from any
			if err := json.Unmarshal(d.body, &from); err != nil {
				log.Printf("Error unmarshalling expected response: %s", err)
				return
			}
			var to any
			if err := json.Unmarshal(d.messageResponse.body, &to); err != nil {
				log.Printf("Error unmarshalling actual response: %s", err)
				return
			}
			if !logEntry.CompareJSON(from, to) {
				return
			}
		}
	}
	log.Println("OK")
}

func (d *Data) Replay(_ net.IP) {
	if conn := d.conn(); conn == nil {
		panic("Cannot handle data: missing websocket connection")
	} else if d.sender() {
		if err := conn.WriteMessage(websocket.TextMessage, d.data.Body.Body); err != nil {
			log.Println(err)
		}
	} else {
		d.messageResponse.messageType, d.messageResponse.body, d.messageResponse.err = conn.ReadMessage()
		d.response = true
	}
}

func NewWebsocketData(base wss.Read, data *wss.Data) *Data {
	return &Data{
		Websocket: Websocket[wss.Data]{
			base,
			*data,
		},
	}
}
