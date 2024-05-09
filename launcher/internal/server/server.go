package server

import (
	"crypto/tls"
	"launcher/internal/config"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"shared/executor"
	"time"
)

const autoServerExecutable string = "server.exe"

var autoServerPaths = []string{`.\`, `..\`, `..\server\`}

func StartServer(config config.ServerConfig) *exec.Cmd {
	if config.Start == "false" {
		return nil
	}
	executablePath := getExecutablePath(config)
	if executablePath == "" {
		return nil
	}
	var cmd *exec.Cmd
	if config.Stop == "true" {
		cmd = executor.StartCustomExecutable(executablePath, true)
	} else {
		cmd = executor.StartCustomExecutable("cmd", true, "/C", "start", executablePath)
	}
	for {
		if LanServer(config.Host, true) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return cmd
}

func getExecutablePath(config config.ServerConfig) string {
	if config.Executable == "auto" {
		for _, path := range autoServerPaths {
			fullPath := path + autoServerExecutable
			if f, err := os.Stat(fullPath); err == nil && !f.IsDir() {
				return fullPath
			}
		}
		return ""
	}
	return config.Executable
}

func LanServer(host string, insecureSkipVerify bool) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://" + host + "/test")
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

func CertificatePairFolder(config config.ServerConfig) string {
	executablePath := getExecutablePath(config)
	if executablePath == "" {
		return ""
	}
	parentDir := filepath.Dir(executablePath)
	if parentDir == "" {
		return ""
	}
	folder := parentDir + `\resources\certificates\`
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if os.Mkdir(folder, os.ModeDir) == nil {
			return ""
		}
	}
	return folder
}

func HasCertificatePair(config config.ServerConfig) bool {
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