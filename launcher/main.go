package main

import (
	"launcher/internal"
	"log"
	"os/exec"
	"time"
)

func main() {
	config := internal.ReadConfig()
	// Setup
	log.Println("Setting up...")
	if config.Server.Host == "" {
		log.Fatal("Server.mapping must not be empty.")
	}
	if config.Server.Start == "auto" {
		log.Println("Waiting for up to 15 seconds for any server announcement already running on LAN...")
		serverAdd := internal.WaitForLanServerAnnounce()
		hostServer := false
		if serverAdd != nil {
			ip := serverAdd.IP.String()
			log.Println("Server " + ip + " already running on LAN")
			if !internal.CheckConnectionFromServer(ip, true) {
				log.Println("Server " + ip + " is not reachable. Hosting own server instead")
			} else {
				config.Server.Host = ip
				hostServer = true
			}
		}
		if hostServer {
			config.Server.Start = "false"
			if config.Server.Stop == "auto" {
				config.Server.Stop = "false"
			}
		} else {
			config.Server.Start = "true"
			if config.Server.Stop == "auto" {
				config.Server.Stop = "true"
			}
		}
	}
	var server *exec.Cmd = nil
	removeHost := false
	removeCertificate := false
	if config.Server.Start == "false" {
		if config.Server.Stop == "true" {
			log.Fatal("Server.Stop cannot be true if Server.Start is false.")
		}
		if config.Server.Host == "0.0.0.0" {
			log.Fatal("Server.Start is false. ServerConfig.mapping cannot be 0.0.0.0. Set a specific host or ip address.")
		}
		if !internal.CheckConnectionFromServer(config.Server.Host, true) {
			log.Fatal("Server.Start is false. " + config.Server.Host + " must be reachable. Review the host is correct, the server is started and the network configuration is correct.")
		}
	} else {
		if !internal.HasCertificatePair(config.Server) {
			if !config.CanTrustCertificate {
				log.Fatal("Server.Start is true and CanTrustCertificate is false. Certificate pair is missing. Generate your own certificates manually.")
			}
			certificateFolder := internal.CertificatePairFolder(config.Server)
			if certificateFolder == "" {
				log.Fatal("Cannot find certificate folder of Server. Make sure the folder structure of the server is correct.")
			}
			if !internal.GenerateCertificatePair(certificateFolder) {
				log.Fatal("Failed to generate certificate pair.")
			}
		}
		log.Println("Starting server...")
		server = internal.StartServer(config.Server)
		if server == nil {
			log.Println("Failed to run server.")
			goto cleanServer
		}
		time.Sleep(time.Second * time.Duration(config.Server.WaitForProcessStart))
	}

	if !internal.MatchesDomain(config.Server.Host) {
		if !config.CanAddHost {
			log.Println("Server.Start is false and CanAddHost is false but server does not match " + internal.Domain + ".")
			goto cleanHost
		} else {
			removeHost = true
			if !internal.AddHost(config.Server.Host) {
				log.Println("Failed to add host.")
				goto cleanHost
			}
		}
	} else if !internal.CheckConnectionFromServer(internal.Domain, true) {
		log.Fatal("Server.Start is false and host matches. " + internal.Domain + " must be reachable. Review the host is reachable via this domain.")
	}

	if !internal.CheckConnectionFromServer(internal.Domain, false) {
		if config.CanTrustCertificate {
			removeCertificate = true
			if !internal.TrustCertificateFromServer(internal.Domain) {
				log.Println("Failed to trust certificate from " + internal.Domain + ".")
				goto cleanCertificate
			} else if !internal.CheckConnectionFromServer(internal.Domain, false) {
				log.Println(internal.Domain + " must have been trusted automatically at this point.")
				goto cleanCertificate
			} else if !internal.LanServer(internal.Domain) {
				log.Println("Something went wrong, " + internal.Domain + " points to the real server, not our LAN server.")
				goto cleanCertificate
			}
		} else {
			log.Println(internal.Domain + " must have been trusted manually.")
			goto cleanCertificate
		}
	} else if !internal.LanServer(internal.Domain) {
		log.Println("Something went wrong, " + internal.Domain + " points to the real server, not our LAN server.")
		goto cleanCertificate
	}

	if config.IsolateMetadata {
		if !internal.BackupMetadata() {
			log.Println("Failed to backup metadata.")
			goto clean
		}
	}

	// Launch game
	log.Println("AoE2:DE looking for it and starting it...")
	if !internal.RunGame(config.Client) {
		log.Println("AoE2:DE failed to start.")
		goto clean
	}
	if !internal.WaitUntilProcessesStart(time.Second, config.Client.WaitForProcessStart) {
		log.Println("AoE2:DE did not start in time..")
		goto clean
	}
	log.Println("AoE2:DE started.")
	internal.WaitUntilProcessesEnd(time.Second * time.Duration(config.Client.CheckProcessRunningEvery))
	log.Println("AoE2:DE stopped.")
clean:
	log.Println("Cleaning up...")

	if config.IsolateMetadata {
		if !internal.RestoreMetadata() {
			log.Println("Failed to restore metadata.")
		}
	}
cleanCertificate:
	if removeCertificate {
		if !internal.UntrustCertificateFromServer(internal.Domain) {
			log.Println("Failed to untrust certificate from " + internal.Domain + ".")
		}
	}
cleanHost:
	if removeHost {
		if !internal.RemoveHost() {
			log.Println("Failed to remove host.")
		}
	}
cleanServer:
	if config.Server.Stop == "true" && server != nil {
		log.Println("Stopping server...")
		err := server.Process.Kill()
		if err != nil {
			log.Println("Failed to stop server.")
		}
	}
}
