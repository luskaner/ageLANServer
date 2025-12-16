package wss

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/wss"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type connectionWrapper struct {
	writeLock *sync.Mutex
	connLock  *sync.RWMutex
	conn      *websocket.Conn
}

func (c *connectionWrapper) withConn(fn func(conn *websocket.Conn)) {
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	if c.conn != nil {
		fn(c.conn)
	}
}

func (c *connectionWrapper) LocalAddr() string {
	var localAddr string
	c.withConn(func(conn *websocket.Conn) {
		localAddr = c.conn.LocalAddr().String()
	})
	return localAddr
}

func (c *connectionWrapper) RemoteAddr() string {
	var remoteAddr string
	c.withConn(func(conn *websocket.Conn) {
		remoteAddr = c.conn.RemoteAddr().String()
	})
	return remoteAddr
}

func (c *connectionWrapper) WriteJSON(v any) error {
	var err error
	c.withConn(func(conn *websocket.Conn) {
		if err = c.conn.WriteJSON(v); err == nil {
			c.logJSON(c.LocalAddr(), c.RemoteAddr(), v)
		}
	})
	return err
}

func (c *connectionWrapper) ReadJSON(v any) error {
	var err error
	c.withConn(func(conn *websocket.Conn) {
		if err = c.conn.ReadJSON(v); err == nil {
			c.logJSON(c.RemoteAddr(), c.LocalAddr(), v)
		}
	})
	return err
}

func computeData(data []byte) *wss.Data {
	var dataHash [64]byte
	if len(data) > 0 {
		dataHash = sha512.Sum512(data)
	}
	if len(data) > 4_096 {
		data = []byte{}
	}
	return &wss.Data{
		Body:     serverCommunication.Body{Body: data},
		BodyHash: serverCommunication.BodyHash{BodyHash: dataHash},
	}
}

func (c *connectionWrapper) logJSON(sender string, receiver string, data any) {
	if logger.CommBuffer != nil {
		dataMarshalled, _ := json.Marshal(data)
		d := computeData(dataMarshalled)
		msg := wss.NewWrite(
			*d,
			serverCommunication.Uptime{
				Uptime: logger.Uptime(nil),
			},
			serverCommunication.Sender{Sender: sender},
			receiver,
		)
		logger.CommBuffer.Log(&msg)
	}
}

func (c *connectionWrapper) logControl(sender string, receiver string, messageType int, data []byte) {
	if logger.CommBuffer != nil {
		d := computeData(data)
		msg := wss.NewWrite(
			wss.Control{
				Data:        *d,
				MessageType: messageType,
			},
			serverCommunication.Uptime{
				Uptime: logger.Uptime(nil),
			},
			serverCommunication.Sender{Sender: sender},
			receiver,
		)
		logger.CommBuffer.Log(&msg)
	}
}

func (c *connectionWrapper) WriteControl(messageType int, data []byte, deadline time.Time) error {
	var err error
	var haveConn bool
	c.withConn(func(conn *websocket.Conn) {
		haveConn = true
		c.logControl(c.LocalAddr(), c.RemoteAddr(), messageType, data)
		err = c.conn.WriteControl(messageType, data, deadline)
	})
	if !haveConn {
		return fmt.Errorf("connection already closed")
	}
	return err
}

func (c *connectionWrapper) logClose(sender string, receiver string) {
	if logger.CommBuffer != nil {
		msg := wss.NewWrite(
			wss.Disconnection{},
			serverCommunication.Uptime{
				Uptime: logger.Uptime(nil),
			},
			serverCommunication.Sender{Sender: sender},
			receiver,
		)
		logger.CommBuffer.Log(&msg)
	}
}

func (c *connectionWrapper) Close() error {
	defer func() {
		c.connLock.Lock()
		defer c.connLock.Unlock()
		c.conn = nil
		c.writeLock = nil
	}()
	c.logClose(c.LocalAddr(), c.RemoteAddr())
	var err error
	c.withConn(func(conn *websocket.Conn) {
		err = c.conn.Close()
	})
	return err
}

func (c *connectionWrapper) SetReadDeadline(t time.Time) error {
	var err error
	c.withConn(func(conn *websocket.Conn) {
		err = c.conn.SetReadDeadline(t)
	})
	return err
}

func (c *connectionWrapper) SetPingHandler(h func(appData string) error) {
	c.conn.SetPingHandler(func(appData string) error {
		c.logControl(c.RemoteAddr(), c.LocalAddr(), websocket.PingMessage, []byte(appData))
		return h(appData)
	})
}

func (c *connectionWrapper) SetCloseHandler(h func(code int, text string) error) {
	c.conn.SetCloseHandler(func(code int, text string) error {
		c.logClose(c.LocalAddr(), c.RemoteAddr())
		return h(code, text)
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var connections = i.NewSafeMap[string, *connectionWrapper]()
var writeWait = 1 * time.Second

func closeConn(conn *connectionWrapper, closeCode int, text string) {
	message := websocket.FormatCloseMessage(closeCode, text)
	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
	_ = conn.Close()
}

func parseMessage(sessions models.Sessions, message i.H, currentSession models.Session) (uint32, models.Session) {
	var sess models.Session
	sess = nil
	op := uint32(message["operation"].(float64))
	if op == 0 {
		sessionToken, ok := message["sessionToken"]
		if ok {
			sess, ok = sessions.GetById(sessionToken.(string))
			if ok {
				return 0, sess
			}

			return 0, nil
		}
	}
	if currentSession != nil {
		var ok bool
		sess, ok = sessions.GetById(currentSession.Id())
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
	sessions := models.G(r).Sessions()
	connWrapper := &connectionWrapper{
		connLock:  &sync.RWMutex{},
		writeLock: &sync.Mutex{},
		conn:      conn,
	}
	if logger.CommBuffer != nil {
		msg := wss.NewWrite(
			wss.Connection{},
			serverCommunication.Uptime{
				Uptime: logger.Uptime(nil),
			},
			serverCommunication.Sender{Sender: connWrapper.RemoteAddr()},
			connWrapper.LocalAddr(),
		)
		logger.CommBuffer.Log(&msg)
	}
	connWrapper.conn.SetCloseHandler(func(code int, text string) error {
		closeConn(connWrapper, code, text)
		return nil
	})
	err = connWrapper.SetReadDeadline(time.Now().Add(time.Minute))
	if err != nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Missing initial message")
		return
	}

	var loginMsg i.H
	err = connWrapper.ReadJSON(&loginMsg)

	if err != nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid or absent login message")
		return
	}

	_, sess := parseMessage(sessions, loginMsg, nil)
	if sess == nil {
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid login message data")
		return
	}

	sessionToken := sess.Id()

	connections.Store(
		sessionToken,
		connWrapper,
		nil,
	)

	sessions.ResetExpiry(sess.Id())

	connWrapper.SetPingHandler(func(message string) error {
		var pingErr error
		pingErr = connWrapper.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWait))
		if pingErr == nil {
			pingErr = connWrapper.SetReadDeadline(time.Now().Add(time.Minute))
			if pingErr == nil {
				sessions.ResetExpiry(sess.Id())
			}
		} else if errors.Is(pingErr, websocket.ErrCloseSent) {
			return nil
		}

		var e net.Error
		if errors.As(pingErr, &e) && e.Temporary() {
			return nil
		}
		return pingErr
	})

	defer func() {
		connections.Delete(sessionToken)
		closeConn(connWrapper, websocket.CloseNormalClosure, "Invalid message")
	}()

	var op uint32
	for {
		var msg i.H
		err = connWrapper.ReadJSON(&msg)
		if err != nil {
			break
		}
		op, sess = parseMessage(sessions, msg, sess)
		if op == 0 {
			if sess == nil {
				break
			}
			if sessId := sess.Id(); sessId != sessionToken {
				connections.StoreAndDelete(sessId, connWrapper, sessionToken)
				sessionToken = sessId
			}
		} else if _, ok := sessions.GetById(sessionToken); !ok {
			break
		}
		sessions.ResetExpiry(sess.Id())
	}
}

func sendMessage(sessionId string, message i.A) bool {
	connWrapper, ok := connections.Load(sessionId)

	if !ok {
		return false
	}

	connWrapper.writeLock.Lock()
	defer connWrapper.writeLock.Unlock()
	err := connWrapper.WriteJSON(message)

	if err != nil {
		return false
	}

	return true
}

func SendOrStoreMessage(session models.Session, action string, message i.A) {
	finalMessage := i.A{0, action, session.GetUserId(), message}
	go func(session models.Session, finalMessage i.A) {
		if ok := sendMessage(session.Id(), finalMessage); !ok {
			session.AddMessage(finalMessage)
		}
	}(session, finalMessage)
}
