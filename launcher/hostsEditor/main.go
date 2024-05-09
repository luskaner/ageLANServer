package main

import (
	"flag"
	"hostsEditor/internal"
	"log"
	"net"
	"shared/executor"
)

type Arguments struct {
	Ip  string
	Add bool
}

func main() {
	log.SetFlags(0)
	if !executor.IsAdmin() {
		log.Fatal("This program must be run as administrator")
	}
	if !internal.ParentMatches("./launcher.exe") {
		log.Fatal("This program must only be executed by the launcher.")
	}
	args := Arguments{}
	flag.StringVar(&args.Ip, "ip", "", "IP address to resolve the host to")
	flag.BoolVar(&args.Add, "add", true, "Add the host to the hosts file or remove it")
	flag.Parse()

	var ok bool
	if args.Add {
		if net.ParseIP(args.Ip) != nil {
			log.Fatal("Invalid IP address")
		} else {
			log.Println("Adding host")
			ok = internal.AddHost(args.Ip)
		}
	} else {
		log.Println("Removing host (if needed) and flushed DNS cache.")
		ok = internal.RemoveHost()
	}
	if !ok {
		log.Fatal("Failed to update hosts file")
	} else {
		log.Println("Updated hosts file (if needed) and flushed DNS cache.")
	}
}
