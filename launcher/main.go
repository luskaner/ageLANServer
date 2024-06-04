package main

import (
	"common"
	"launcher/internal"
	"launcher/internal/executor"
	"launcher/internal/game"
	"launcher/internal/server"
	"launcher/internal/userData"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"shared"
	sharedExecutor "shared/executor"
	"syscall"
	"time"
)

var removeHost = false
var removeCertificate = false
var config internal.Config
var serverProcess *exec.Cmd = nil

func cleanup() {
	log.Println("Cleaning up...")
	isAdmin := sharedExecutor.IsAdmin()
	if config.IsolateProfiles {
		if !userData.RestoreProfiles() {
			log.Println("Failed to restore profiles.")
		}
	}

	if config.IsolateMetadata {
		if !userData.Metadata.Restore() {
			log.Println("Failed to restore metadata.")
		}
	}

	if removeCertificate {
		if isAdmin {
			log.Println("Removing server certificate from store.")
		} else {
			log.Println("Removing server certificate from store, accept any dialog if it appears...")
		}
		if !server.UntrustCertificate() {
			log.Println("Failed to untrust certificate from " + common.Domain + ".")
		}
	}

	if removeHost {
		if isAdmin {
			log.Println("Removing host from hosts file.")
		} else {
			log.Println("Removing host from hosts file, accept any dialog if it appears...")
		}
		if !executor.RemoveHost(sharedExecutor.IsAdmin()) {
			log.Println("Failed to remove host.")
		}
	}

	if serverProcess != nil {
		log.Println("Stopping server...")
		if !server.StopServer(serverProcess) {
			log.Println("Failed to stop server.")
		}
	}
}

func main() {
	errorCode := 0
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			close(sigs)
			cleanup()
			os.Exit(errorCode)
		}
	}()
	defer func() {
		os.Exit(errorCode)
	}()
	defer func() {
		cleanup()
	}()
	if game.ProcessesExists() {
		log.Println("Game is already running.")
		return
	}
	c := internal.ReadConfig()
	isAdmin := sharedExecutor.IsAdmin()
	if isAdmin {
		log.Println("Running as administrator, this is not recommended for security reasons. It will request elevated privileges if/when it needs.")
	}
	// Setup
	log.Println("Setting up...")
	if c.Server.Start == "auto" {
		log.Println("Waiting for up to 15 seconds for any server announcement already running on LAN...")
		serverAdd := server.WaitForLanServerAnnounce()
		hostServer := false
		if serverAdd != nil {
			ip := serverAdd.IP.String()
			log.Println("Server " + ip + " already running on LAN")
			if !server.CheckConnectionFromServer(ip, true) {
				log.Println("Server " + ip + " is not reachable. Hosting own server instead")
			} else {
				c.Server.Host = ip
				hostServer = true
			}
		}
		if hostServer {
			c.Server.Start = "false"
			if c.Server.Stop == "auto" {
				c.Server.Stop = "false"
			}
		} else {
			c.Server.Start = "true"
			if c.Server.Stop == "auto" {
				c.Server.Stop = "true"
			}
		}
	}
	var ipOfHost string
	if c.Server.Start == "false" {
		if c.Server.Stop == "true" {
			log.Println("Server.Start is false. Ignoring Server.Stop being true.")
			return
		}
		if c.Server.Host == "" {
			log.Println("Server.Start is false. Server.Host must not be empty as we need to know which host to connect to.")
			return
		}
		if !server.CheckConnectionFromServer(c.Server.Host, true) {
			log.Println("Server.Start is false. " + c.Server.Host + " must be reachable. Review the host is correct, the server is started and the network configuration is correct.")
			return
		}
	} else {
		serverExecutablePath := server.GetExecutablePath(c.Server)
		if !common.HasCertificatePair(serverExecutablePath) {
			if !c.CanTrustCertificate {
				log.Println("Server.Start is true and CanTrustCertificate is false. Certificate pair is missing. Generate your own certificates manually.")
				return
			}
			certificateFolder := common.CertificatePairFolder(serverExecutablePath)
			if certificateFolder == "" {
				log.Println("Cannot find certificate folder of Server. Make sure the folder structure of the server is correct.")
				return
			}
			if !server.GenerateCertificatePair(certificateFolder) {
				log.Println("Failed to generate certificate pair.")
				return
			}
		}
		log.Println("Starting server...")
		var ok bool
		ok, serverProcess = server.StartServer(c.Server)
		if !ok {
			log.Println("Failed to start server.")
			return
		}
	}

	ipOfHost = shared.ResolveHost(c.Server.Host)
	if !shared.Matches(ipOfHost, common.Domain) {
		if !c.CanAddHost {
			log.Println("Server.Start is false and CanAddHost is false but server does not match " + common.Domain + ".")
			return
		} else {
			removeHost = true
			if isAdmin {
				log.Println("Adding host to hosts file.")
			} else {
				log.Println("Adding host to hosts file, accept any dialog if it appears...")
			}
			if !executor.AddHost(isAdmin, ipOfHost) {
				log.Println("Failed to add host.")
				return
			}
		}
	} else if !server.CheckConnectionFromServer(common.Domain, true) {
		log.Println("Server.Start is false and host matches. " + common.Domain + " must be reachable. Review the host is reachable via this domain.")
		return
	}

	if !server.CheckConnectionFromServer(common.Domain, false) {
		if c.CanTrustCertificate {
			removeCertificate = true
			if isAdmin {
				log.Println("Adding server certificate to store.")
			} else {
				log.Println("Adding server certificate to store, accept any dialog if it appears...")
			}
			if !server.TrustCertificateFromServer(common.Domain) {
				log.Println("Failed to trust certificate from " + common.Domain + ".")
				return
			} else if !server.CheckConnectionFromServer(common.Domain, false) {
				log.Println(common.Domain + " must have been trusted automatically at this point.")
				return
			} else if !server.LanServer(common.Domain, false) {
				log.Println("Something went wrong, " + common.Domain + " either points to the real server or certificate issue.")
				return
			}
		} else {
			log.Println(common.Domain + " must have been trusted manually.")
			return
		}
	} else if !server.LanServer(common.Domain, false) {
		log.Println("Something went wrong, " + common.Domain + " either points to the real server or certificate issue.")
		return
	}

	if c.IsolateMetadata {
		if !userData.Metadata.Backup() {
			log.Println("Failed to backup metadata.")
			return
		}
	}

	if c.IsolateProfiles {
		if !userData.BackupProfiles() {
			log.Println("Failed to backup profiles.")
			return
		}
	}

	// Launch game
	log.Println("AoE2:DE looking for it and starting it...")
	if !game.RunGame(c.Client.Executable) {
		log.Println("AoE2:DE failed to start.")
		return
	}
	if !game.WaitUntilProcessesStart(time.Second, 60) {
		log.Println("AoE2:DE did not start in time...")
		return
	}
	log.Println("AoE2:DE started.")
	game.WaitUntilProcessesEnd(time.Second * 10)
	log.Println("AoE2:DE stopped.")
}
