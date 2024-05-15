package main

import (
	"common"
	"launcher/internal"
	"launcher/internal/game"
	"launcher/internal/server"
	"launcher/internal/userData"
	"log"
	"os/exec"
	"shared"
	"shared/executor"
	"time"
)

func main() {
	c := internal.ReadConfig()
	isAdmin := executor.IsAdmin()
	if isAdmin {
		log.Println("Running as administrator, this is not recommended. It will request elevated privileges if/when it needs.")
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
	var serverProcess *exec.Cmd = nil
	removeHost := false
	removeCertificate := false
	if c.Server.Start == "false" {
		if c.Server.Stop == "true" {
			log.Println("Server.Start is false. Ignoring Server.Stop being true.")
		}
		if c.Server.Host == "" {
			log.Fatal("Server.Start is false. Server.Host must not be empty as we need to know which host to connect to.")
		}
		if !server.CheckConnectionFromServer(c.Server.Host, true) {
			log.Fatal("Server.Start is false. " + c.Server.Host + " must be reachable. Review the host is correct, the server is started and the network configuration is correct.")
		}
	} else {
		serverExecutablePath := server.GetExecutablePath(c.Server)
		if !common.HasCertificatePair(serverExecutablePath) {
			if !c.CanTrustCertificate {
				log.Fatal("Server.Start is true and CanTrustCertificate is false. Certificate pair is missing. Generate your own certificates manually.")
			}
			certificateFolder := common.CertificatePairFolder(serverExecutablePath)
			if certificateFolder == "" {
				log.Fatal("Cannot find certificate folder of Server. Make sure the folder structure of the server is correct.")
			}
			if !server.GenerateCertificatePair(certificateFolder) {
				log.Fatal("Failed to generate certificate pair.")
			}
		}
		log.Println("Starting server...")
		serverProcess = server.StartServer(c.Server)
		if serverProcess == nil {
			log.Println("Failed to run server.")
			goto cleanServer
		}
	}

	ipOfHost = shared.ResolveHost(c.Server.Host)
	if !shared.Matches(ipOfHost, common.Domain) {
		if !c.CanAddHost {
			log.Println("Server.Start is false and CanAddHost is false but server does not match " + common.Domain + ".")
			goto cleanHost
		} else {
			removeHost = true
			if !internal.AddHost(isAdmin, ipOfHost) {
				log.Println("Failed to add host.")
				goto cleanHost
			}
		}
	} else if !server.CheckConnectionFromServer(common.Domain, true) {
		log.Fatal("Server.Start is false and host matches. " + common.Domain + " must be reachable. Review the host is reachable via this domain.")
	}

	if !server.CheckConnectionFromServer(common.Domain, false) {
		if c.CanTrustCertificate {
			removeCertificate = true
			if !server.TrustCertificateFromServer(common.Domain) {
				log.Println("Failed to trust certificate from " + common.Domain + ".")
				goto cleanCertificate
			} else if !server.CheckConnectionFromServer(common.Domain, false) {
				log.Println(common.Domain + " must have been trusted automatically at this point.")
				goto cleanCertificate
			} else if !server.LanServer(common.Domain, false) {
				log.Println("Something went wrong, " + common.Domain + " either points to the real server or certificate issue.")
				goto cleanCertificate
			}
		} else {
			log.Println(common.Domain + " must have been trusted manually.")
			goto cleanCertificate
		}
	} else if !server.LanServer(common.Domain, false) {
		log.Println("Something went wrong, " + common.Domain + " either points to the real server or certificate issue.")
		goto cleanCertificate
	}

	if c.IsolateMetadata {
		if !userData.Metadata.Backup() {
			log.Println("Failed to backup metadata.")
			goto cleanMetadata
		}
	}

	if c.IsolateProfiles {
		if !userData.BackupProfiles() {
			log.Println("Failed to backup profiles.")
			goto clean
		}
	}

	// Launch game
	log.Println("AoE2:DE looking for it and starting it...")
	if !game.RunGame(c.Client.Executable) {
		log.Println("AoE2:DE failed to start.")
		goto clean
	}
	if !game.WaitUntilProcessesStart(time.Second, 60) {
		log.Println("AoE2:DE did not start in time...")
		goto clean
	}
	log.Println("AoE2:DE started.")
	game.WaitUntilProcessesEnd(time.Second * 10)
	log.Println("AoE2:DE stopped.")
clean:
	log.Println("Cleaning up...")

	if c.IsolateProfiles {
		if !userData.RestoreProfiles() {
			log.Println("Failed to restore profiles.")
		}
	}

cleanMetadata:
	if c.IsolateMetadata {
		if !userData.Metadata.Restore() {
			log.Println("Failed to restore metadata.")
		}
	}
cleanCertificate:
	if removeCertificate {
		if !server.UntrustCertificateFromServer(common.Domain) {
			log.Println("Failed to untrust certificate from " + common.Domain + ".")
		}
	}
cleanHost:
	if removeHost {
		if !internal.RemoveHost(isAdmin) {
			log.Println("Failed to remove host.")
		}
	}
cleanServer:
	if c.Server.Stop == "true" && serverProcess != nil {
		log.Println("Stopping server...")
		err := serverProcess.Process.Kill()
		if err != nil {
			log.Println("Failed to stop server.")
		}
	}
}
