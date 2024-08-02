package wss

import (
	"github.com/gorilla/websocket"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var timeoutTime = 60 * time.Minute

var lock = i.NewKeyRWMutex()
var connections = make(map[string]*websocket.Conn)

func closeConn(conn *websocket.Conn, closeCode int, text string) {
	message := websocket.FormatCloseMessage(closeCode, text)
	if err := conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second)); err != nil {
		log.Println(err)
	}
	if err := conn.Close(); err != nil {
		log.Println(err)
	}
}

func parseMessage(message i.H, currentSession *models.Session) (bool, uint32, *models.Session) {
	var sess *models.Session
	sess = nil
	op := uint32(message["operation"].(float64))
	if op == 0 {
		sessionToken, ok := message["sessionToken"]
		if ok {
			sess, ok = models.GetSessionById(sessionToken.(string))
			if ok {
				return true, 0, sess
			} else {
				return false, 0, nil
			}
		}
	}
	if currentSession != nil {
		sess, _ = models.GetSessionById(currentSession.GetId())
	}
	return false, op, sess
}

func Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	done := make(chan struct{})
	timeout := time.AfterFunc(time.Minute, func() {
		select {
		case <-done:
		default:
			closeConn(conn, websocket.CloseNormalClosure, "Timeout")
		}
	})

	var msg i.H
	err = conn.ReadJSON(&msg)

	if err != nil {
		defer func() {
			close(done)
			closeConn(conn, websocket.CloseNormalClosure, "Invalid login message format")
		}()
		return
	}
	reset, _, sess := parseMessage(msg, nil)
	if reset {
		timeout.Reset(timeoutTime)
	} else {
		defer func() {
			close(done)
			closeConn(conn, websocket.CloseNormalClosure, "Invalid login message data")
		}()
		return
	}

	sessionToken := sess.GetId()
	lock.Lock(sessionToken)
	connections[sessionToken] = conn
	lock.Unlock(sessionToken)

	for {
		err = conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		log.Printf("Received: %v", msg)
		reset, op, sess := parseMessage(msg, sess)
		if op == 0 {
			if sess == nil {
				break
			} else if sess.GetId() != sessionToken {
				lock.Lock(sessionToken)
				delete(connections, sessionToken)
				lock.Unlock(sessionToken)
				sessionToken = sess.GetId()
				lock.Lock(sessionToken)
				connections[sessionToken] = conn
				lock.Unlock(sessionToken)
			}
			if reset {
				timeout.Reset(timeoutTime)
			}
		} else if _, ok := models.GetSessionById(sessionToken); !ok {
			break
		}
		// TODO: Handle other operations
		log.Printf("Operation: %v", op)
	}

	lock.Lock(sessionToken)
	delete(connections, sessionToken)
	close(done)
	closeConn(conn, websocket.CloseNormalClosure, "Invalid message")
}

func SendMessage(sessionId string, message i.A) bool {
	lock.RLock(sessionId)

	conn, ok := connections[sessionId]

	if !ok {
		lock.RUnlock(sessionId)
		return false
	}

	err := conn.WriteJSON(message)
	lock.RUnlock(sessionId)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
