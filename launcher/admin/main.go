package main

import (
	"admin/internal"
	"encoding/base64"
	"encoding/json"
	"flag"
	mapset "github.com/deckarep/golang-set/v2"
	"log"
	"net"
	"shared"
	"shared/executor"
)

var actions = []string{"cleanup", "addHosts", "removeHosts", "addCert", "removeCert"}

type Arguments struct {
	Action       string
	Subarguments string
}

func parseSubarguments(subArguments string) map[string]interface{} {
	jsonBytes, err := base64.StdEncoding.DecodeString(subArguments)
	if err != nil {
		log.Fatal("Invalid subArguments")
	}

	jsonString := string(jsonBytes)

	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		log.Fatal("Invalid subArguments")
	}

	return jsonData
}

func main() {
	log.SetFlags(0)
	if !executor.IsAdmin() {
		log.Fatal("This program must be run as administrator")
	}
	args := Arguments{}
	flag.StringVar(&args.Action, "action", "cleanup", "Action to perform (cleanup, addHosts, removeHosts, addCert, removeCert).")
	flag.StringVar(&args.Subarguments, "subArguments", "{}", "Subarguments for the action. ips for addHosts, certData for addCert.")
	flag.Parse()

	foundAction := false
	for _, action := range actions {
		if args.Action == action {
			foundAction = true
			break
		}
	}

	if !foundAction {
		log.Fatal("Invalid action")
	}

	var ok bool
	var subject string
	switch args.Action {
	case "addHosts":
		subject = "hosts"
		subArguments := parseSubarguments(args.Subarguments)
		if subArguments["ips"] == nil {
			log.Fatal("Missing ips in subarguments")
		} else {
			ipsSlice := subArguments["ips"].([]interface{})
			ipsMap := mapset.NewSet[string]()
			for _, ip := range ipsSlice {
				ipStr := ip.(string)
				if net.ParseIP(ipStr) == nil {
					log.Fatal("Invalid IP address")
				}
				ipsMap.Add(ipStr)
			}
			log.Println("Adding hosts")
			ok = shared.AddHosts(ipsMap)
		}
	case "removeHosts":
		subject = "hosts"
		log.Println("Removing hosts")
		ok = shared.RemoveHosts()
	case "addCert":
		subject = "certificate"
		subArguments := parseSubarguments(args.Subarguments)
		if subArguments["certData"] == nil {
			log.Fatal("Missing certData in subarguments")
		}
		log.Println("Adding local certificate")
		cert := internal.Base64ToCertificate(subArguments["certData"].(string))
		if cert == nil {
			log.Fatal("Failed to parse certificate data")
		}
		ok = shared.TrustCertificate(false, cert)
	case "removeCert":
		subject = "certificate"
		log.Println("Removing local certificate")
		ok = shared.UntrustCertificate(false)
	case "cleanup":
		subject = "hosts and certificate"
		log.Println("Cleaning up")
		log.Println("Removing user certificate")
		ok = shared.UntrustCertificate(true)
		log.Println("Removing local certificate")
		ok = ok || shared.UntrustCertificate(false)
		log.Println("Removing hosts")
		ok = ok || shared.RemoveHosts()
	}

	if !ok {
		log.Println("Failed to update", subject)
	} else {
		log.Println("Updated", subject)
	}
}
