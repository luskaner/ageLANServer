package decoder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/luskaner/ageLANServer/common/logger/serverCommunication"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/request"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/wss"
	"github.com/luskaner/ageLANServer/server-replay/internal/logEntry"
	"github.com/luskaner/ageLANServer/server-replay/internal/logEntry/http"
	logEntryWss "github.com/luskaner/ageLANServer/server-replay/internal/logEntry/wss"
)

var wssSubtypes = map[string]reflect.Type{
	"Connection":    reflect.TypeOf(wss.Connection{}),
	"Disconnection": reflect.TypeOf(wss.Disconnection{}),
	"Data":          reflect.TypeOf(wss.Data{}),
}

const maxTokenSize = 256 * 1024

func Decode(f *os.File) error {
	var messageType serverCommunication.MessageType
	scanner := bufio.NewScanner(f)
	buf := make([]byte, maxTokenSize)
	scanner.Buffer(buf, maxTokenSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.Unmarshal(line, &messageType); err != nil {
			return err
		}
		switch messageType.Type {
		case serverCommunication.MessageRequest:
			var req request.Read
			if err := json.Unmarshal(line, &req); err != nil {
				return err
			}
			logEntry.Add(http.NewRequest(req))
		case serverCommunication.MessageWSS:
			var req wss.Read
			if err := json.Unmarshal(line, &req); err != nil {
				return err
			}
			typ, ok := wssSubtypes[req.Subtype]
			if !ok {
				continue
			}
			wssValuePtr := reflect.New(typ)
			wssInterface := wssValuePtr.Interface()
			if err := json.Unmarshal(req.Data, &wssInterface); err != nil {
				return err
			}
			switch wssInterface.(type) {
			case *wss.Connection:
				logEntry.Add(logEntryWss.NewWebsocketConnection(req, wssInterface.(*wss.Connection)))
			case *wss.Disconnection:
				logEntry.Add(logEntryWss.NewWebsocketDisconnection(req, wssInterface.(*wss.Disconnection)))
			case *wss.Data:
				logEntry.Add(logEntryWss.NewWebsocketData(req, wssInterface.(*wss.Data)))
			}
		default:
			return fmt.Errorf("unrecognized message type: %v", messageType.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	logEntry.Sort()
	return nil
}
