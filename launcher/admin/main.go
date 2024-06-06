package main

import (
	"admin/internal"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net"
	"shared"
	"shared/executor"
)

var actions = []string{"addHost", "removeHost", "addCert", "removeCert"}

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
	if !internal.ParentMatches("./launcher.exe") {
		log.Println("This program should only be executed by the launcher for admin tasks.")
	}
	args := Arguments{}
	flag.StringVar(&args.Action, "action", "", "Action to perform (addHost, removeHost, addCert, removeCert).")
	flag.StringVar(&args.Subarguments, "subArguments", "", "Subarguments for the action. ip for addHost, certData for addCert.")
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
	case "addHost":
		subject = "host"
		subArguments := parseSubarguments(args.Subarguments)
		if subArguments["ip"] == nil {
			log.Fatal("Missing ip in subarguments")
		} else if net.ParseIP(subArguments["ip"].(string)) == nil {
			log.Fatal("Invalid IP address")
		} else {
			log.Println("Adding host")
			ok = shared.AddHost(subArguments["ip"].(string))
		}
	case "removeHost":
		subject = "host"
		log.Println("Removing host")
		ok = shared.RemoveHost()
	case "addCert":
		subject = "certificate"
		subArguments := parseSubarguments(args.Subarguments)
		if subArguments["certData"] == nil {
			log.Fatal("Missing certData in subarguments")
		}
		log.Println("Adding certificate")
		cert := internal.Base64ToCertificate(subArguments["certData"].(string))
		if cert == nil {
			log.Fatal("Failed to parse certificate data")
		}
		ok = shared.TrustCertificate(cert)
	case "removeCert":
		subject = "certificate"
		log.Println("Removing certificate")
		ok = shared.UntrustCertificate()
	}

	if !ok {
		log.Println("Failed to update ", subject)
	} else {
		log.Println("Updated ", subject)
	}
}
