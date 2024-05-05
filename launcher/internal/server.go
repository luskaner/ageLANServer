package internal

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartServer(config ServerConfig) *exec.Cmd {
	if config.Start == "false" {
		return nil
	}
	executablePath := getExecutablePath(config)
	if config.Stop == "true" {
		return StartCustomExecutable(executablePath)
	}
	return StartCustomExecutable("cmd", "/C", "start", executablePath)
}

func getExecutablePath(config ServerConfig) string {
	if config.Executable == "" {
		dir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return dir + `\server.exe`
	}
	return config.Executable
}

func LanServer(host string) bool {
	resp, err := http.Get("https://" + host + "/test")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func WaitForLanServerAnnounce() *net.UDPAddr {
	addr := net.UDPAddr{
		Port: 59999,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil
	}
	defer func(conn *net.UDPConn) {
		_ = conn.Close()
	}(conn)

	err = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err != nil {
		return nil
	}

	buf := make([]byte, 1)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil
		}
		if n == 1 && buf[0] == 43 {
			return addr
		}
		return nil
	}
}

func CertificatePairFolder(config ServerConfig) string {
	executablePath := getExecutablePath(config)
	parentDir := filepath.Dir(executablePath)
	if parentDir == "" {
		return ""
	}
	folder := parentDir + `\resources\certificates\`
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return ""
	}
	return folder
}

func HasCertificatePair(config ServerConfig) bool {
	parentDir := CertificatePairFolder(config)
	if parentDir == "" {
		return false
	}
	if _, err := os.Stat(parentDir + Cert); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(parentDir + Key); os.IsNotExist(err) {
		return false
	}
	return true
}
