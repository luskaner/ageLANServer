package main

import (
	"flag"
	"hostsEditor/internal"
	"log"
	"shared/executor"
)

type Arguments struct {
	Ip  string
	Add bool
}

func main() {
	log.SetFlags(0)
	if !executor.IsAdmin() {
		log.Fatal("You need to run this program as an administrator")
	}
	args := Arguments{}
	flag.StringVar(&args.Ip, "ip", "", "IP address to the resolve the host to")
	flag.BoolVar(&args.Add, "add", true, "Add the host to the hosts file or remove it")
	flag.Parse()

	var ok bool
	if args.Add {
		log.Println("Adding host")
		ok = internal.AddHost(args.Ip)
	} else {
		log.Println("Removing host")
		ok = internal.RemoveHost()
	}
	if !ok {
		log.Fatal("Failed to update hosts file")
	}
}
