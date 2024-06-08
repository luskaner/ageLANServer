package main

import (
	"common"
	"fmt"
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
	"sort"
	"strings"
	"syscall"
	"time"
)

var removeHost = false
var removeCertificate = false
var c internal.Config
var serverProcess *exec.Cmd = nil

func cleanup() {
	log.Println("Cleaning up...")
	isAdmin := sharedExecutor.IsAdmin()
	if c.IsolateProfiles {
		if !userData.RestoreProfiles() {
			log.Println(`Failed to restore profiles. Some or all folders (numeric) might need to be switched around manually "*.bak" <-> "*"`)
		}
	}

	if c.IsolateMetadata {
		if !userData.Metadata.Restore() {
			log.Println(`Failed to restore metadata. Switch the folder names around manually "%USERPROFILE%\Games\Age of Empires 2 DE\metadata" <-> "%USERPROFILE%\Games\Age of Empires 2 DE\metadata.bak" if needed`)
		}
	}

	if removeCertificate {
		if isAdmin {
			log.Println("Removing server certificate from store.")
		} else {
			log.Println("Removing server certificate from store, accept any dialog if it appears...")
		}
		if !executor.RemoveCertificateInternal(c.CanTrustCertificate == "local" && !isAdmin) {
			var manager string
			if c.CanTrustCertificate == "local" {
				manager = "certlm.msc"
			} else {
				manager = "certmgr.msc"
			}
			log.Println(fmt.Sprintf(`Failed to untrust certificate from store. Remove manually by opening "%s" and deleting the certificate named "%s" from the "Trusted Root Certification Authorities" folder if needed.`, manager, common.Domain))
		}
	}

	if removeHost {
		if isAdmin {
			log.Println("Removing host from hosts file.")
		} else {
			log.Println("Removing host from hosts file, accept any dialog if it appears...")
		}
		if !executor.RemoveHostsInternal(!isAdmin) {
			log.Println(fmt.Sprintf(`Failed to remove host. Remove manually by opening "%%WINDIR%%\System32\drivers\etc\hosts" file in a text editor with admin rights and deleting the line with "%s" if needed`, common.Domain))
		}
	}

	if serverProcess != nil {
		log.Println("Stopping server...")
		if !server.StopServer(serverProcess) {
			log.Println(fmt.Sprintf(`Failed to stop server. Kill the process "server.exe" (PID: %d) with Task Manager if needed.`, serverProcess.Process.Pid))
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
		log.Println("Game is already running, exiting...")
		return
	}
	c = internal.ReadConfig()
	isAdmin := sharedExecutor.IsAdmin()
	if isAdmin {
		log.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when it needs.")
	}
	// Setup
	log.Println("Setting up...")
	if c.Server.Start == "auto" {
		log.Println("Waiting 15 seconds for server announcements on LAN...")
		servers := server.LanServersAnnounced()
		joinServer := false
		if servers != nil && !servers.IsEmpty() {
			serverTags := make([]string, servers.Cardinality())
			serversStr := servers.ToSlice()
			for i, ip := range serversStr {
				hosts := shared.IpToHosts(ip)
				hostsStr := ""
				if !hosts.IsEmpty() {
					hostsSlice := hosts.ToSlice()
					sort.Strings(hostsSlice)
					hostsStr = strings.Join(hostsSlice, ", ")
				}
				serverTags[i] = fmt.Sprintf("%s (%s)", ip, hostsStr)
			}
			if servers.Cardinality() == 1 {
				log.Printf("Found a single server \"%s\", will connect to it...\n", serverTags[0])
				c.Server.Host = serversStr[0]
				joinServer = true
			} else {
				var option int
				for {
					log.Println("Found the following servers:")
					for i := range serversStr {
						log.Printf("%d. %s\n", i+1, serverTags[i])
					}
					log.Printf("%d. Host server\n", len(serversStr)+1)
					log.Printf("Enter the number of the server (1-%d): ", len(serversStr)+1)
					_, err := fmt.Scan(&option)
					if err != nil || option < 1 || option > len(serversStr)+1 {
						log.Println("Invalid (or error reading) option. Please enter a number from the list.")
						continue
					}
					if option == len(serversStr)+1 {
						break
					}
					ip := serversStr[option-1]
					if !server.CheckConnectionFromServer(ip, true) {
						tag := serverTags[option-1]
						log.Println(fmt.Sprintf("Server %s is not reachable. Please enter another number from the list", tag))
					} else {
						c.Server.Host = ip
						joinServer = true
						break
					}
				}
			}
		}
		if joinServer {
			c.Server.Start = "false"
			if c.Server.Stop == "auto" {
				c.Server.Stop = "false"
			}
		} else {
			c.Server.Host = "127.0.0.1"
			c.Server.Start = "true"
			if c.Server.Stop == "auto" {
				c.Server.Stop = "true"
			}
		}
	}
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
			if c.CanTrustCertificate == "false" {
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

	if !shared.Matches(c.Server.Host, common.Domain) {
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
			if !executor.AddHostsInternal(!isAdmin, c.Server.Host) {
				log.Println("Failed to add host.")
				return
			}
		}
	} else if !server.CheckConnectionFromServer(common.Domain, true) {
		log.Println("Server.Start is false and host matches. " + common.Domain + " must be reachable. Review the host is reachable via this domain.")
		return
	}

	if !server.CheckConnectionFromServer(common.Domain, false) {
		if c.CanTrustCertificate != "false" {
			removeCertificate = true
			if isAdmin {
				log.Println("Adding server certificate to store.")
			} else {
				log.Println("Adding server certificate to store, accept any dialog if it appears...")
			}
			cert := server.ReadCertificateFromServer(common.Domain)
			if cert == nil {
				log.Println("Failed to read certificate from " + common.Domain + ".")
				return
			} else if !executor.AddCertificateInternal(c.CanTrustCertificate == "local" && !isAdmin, cert) {
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
	if !game.RunGame(c.Client.Executable, c.CanTrustCertificate == "user") {
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
