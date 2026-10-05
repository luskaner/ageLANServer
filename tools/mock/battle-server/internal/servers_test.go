//go:build windows

package internal

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func freeTCPPort(t *testing.T) uint16 {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return uint16(port)
}

func TestTCPListenEchoes(t *testing.T) {
	port := freeTCPPort(t)
	ListenTCP(port)

	conn, err := net.DialTimeout("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	const payload = "ping-through-echo"
	if _, err = conn.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(payload))
	if _, err = conn.Read(buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != payload {
		t.Fatalf("echo = %q, want %q", buf, payload)
	}
}

func TestWebsocketEcho(t *testing.T) {
	port := freeTCPPort(t)
	go ListenAndServeWebsocket(port, "", "")

	url := "ws://127.0.0.1:" + strconv.Itoa(int(port))
	var conn *websocket.Conn
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			conn = c
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("websocket dial: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	const msg = "hello ws"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatal(err)
	}
	kind, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if kind != websocket.TextMessage || string(data) != msg {
		t.Fatalf("echo = %d %q", kind, data)
	}
	if strings.Contains(msg, "never") {
		t.Fatal("unreachable")
	}
}
