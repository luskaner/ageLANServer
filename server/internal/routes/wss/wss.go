package wss

import (
	"errors"
	"github.com/gorilla/websocket"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net"
	"net/http"
	"sync"
	"time"
)

type connectionWrapper struct {
	writeLock *sync.Mutex
	conn      *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var connections = i.NewSafeMap[string, *connectionWrapper]()
var writeWait = 1 * time.Second

func closeConn(conn *websocket.Conn, closeCode int, text string) {
	message := websocket.FormatCloseMessage(closeCode, text)
	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
	_ = conn.Close()
}

func parseMessage(message i.H, currentSession *models.Session) (uint32, *models.Session) {
	var sess *models.Session
	sess = nil
	op := uint32(message["operation"].(float64))
	if op == 0 {
		sessionToken, ok := message["sessionToken"]
		if ok {
			sess, ok = models.GetSessionById(sessionToken.(string))
			if ok {
				return 0, sess
			} else {
				return 0, nil
			}
		}
	}
	if currentSession != nil {
		var ok bool
		sess, ok = models.GetSessionById(currentSession.GetId())
		if !ok {
			return 0, nil
		}
	}
	return op, sess
}

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(time.Minute))
	if err != nil {
		return
	}

	var loginMsg i.H
	err = conn.ReadJSON(&loginMsg)

	if err != nil {
		closeConn(conn, websocket.CloseNormalClosure, "Invalid or absent login message")
		return
	}

	_, sess := parseMessage(loginMsg, nil)
	if sess == nil {
		closeConn(conn, websocket.CloseNormalClosure, "Invalid login message data")
		return
	}

	sessionToken := sess.GetId()
	connWrapper := &connectionWrapper{
		writeLock: &sync.Mutex{},
		conn:      conn,
	}
	connections.Store(
		sessionToken,
		connWrapper,
		nil,
	)

	sess.ResetExpiryTimer()

	conn.SetPingHandler(func(message string) error {
		var pingErr error
		pingErr = conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWait))
		if pingErr == nil {
			pingErr = conn.SetReadDeadline(time.Now().Add(time.Minute))
			if pingErr == nil {
				sess.ResetExpiryTimer()
			}
		} else if errors.Is(pingErr, websocket.ErrCloseSent) {
			return nil
		} else {
			var e net.Error
			if errors.As(pingErr, &e) && e.Temporary() {
				return nil
			}
		}
		return pingErr
	})

	defer func() {
		connections.Delete(sessionToken)
		closeConn(conn, websocket.CloseNormalClosure, "Invalid message")
	}()

	var op uint32
	for {
		var msg i.H
		err = conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		op, sess = parseMessage(msg, sess)
		if op == 0 {
			if sess == nil {
				break
			}
			if sessId := sess.GetId(); sessId != sessionToken {
				connections.StoreAndDelete(sessId, connWrapper, sessionToken)
				sessionToken = sessId
			}
		} else if _, ok := models.GetSessionById(sessionToken); !ok {
			break
		}
		sess.ResetExpiryTimer()
	}
}

func sendMessage(sessionId string, message i.A) bool {
	connWrapper, ok := connections.Load(sessionId)

	if !ok {
		return false
	}

	connWrapper.writeLock.Lock()
	defer connWrapper.writeLock.Unlock()
	err := connWrapper.conn.WriteJSON(message)

	if err != nil {
		return false
	}

	return true
}

func SendOrStoreMessage(session *models.Session, action string, message i.A) {
	finalMessage := i.A{0, action, session.GetUserId(), message}
	go func(session *models.Session, finalMessage i.A) {
		if ok := sendMessage(session.GetId(), finalMessage); !ok {
			session.AddMessage(finalMessage)
		}
	}(session, finalMessage)
}
