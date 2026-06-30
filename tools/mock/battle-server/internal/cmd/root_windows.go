package cmd

import (
	"battle-server/internal"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common/battleServer"
)

// Set in go build
var defaultRelayBroadcastPortStr = ""
var defaultBsPublicPortStr = ""
var defaultOutOfBandStr = ""

// Unused
var region string
var simulationPeriod int

var name string
var publicPort uint
var relayBroadcastPort uint
var bsPort uint
var webSocketPort uint
var outOfBandPort uint

var sslCert string
var sslKey string

// var muninLog string
// var serverLog string

func rootCmd() error {
	if bsPort == 0 {
		return fmt.Errorf("battle server port must be specified and non-zero")
	}
	var secure bool
	if sslKey != "" {
		if _, err := os.Stat(sslKey); err != nil {
			return fmt.Errorf("invalid SSL key file path: %w", err)
		}
		secure = true
	}
	if sslCert != "" {
		if _, err := os.Stat(sslCert); err != nil {
			return fmt.Errorf("invalid SSL certificate file path: %w", err)
		}
		secure = true
	}
	id := uuid.New()
	log.Printf("id: %s", id)
	internal.ListenTCP(uint16(bsPort))
	if outOfBandPort != 0 {
		if secure {
			internal.ListenAndServeWebsocket(uint16(outOfBandPort), sslCert, sslKey)
		} else {
			internal.ListenTCP(uint16(outOfBandPort))
		}
	}
	if webSocketPort != 0 {
		internal.ListenAndServeWebsocket(uint16(webSocketPort), sslCert, sslKey)
	}
	if relayBroadcastPort != 0 {
		log.Printf("Broadcasting on port: %d", relayBroadcastPort)
		if err := internal.Broadcast(
			battleServer.BroadcastMessage{
				Id:            id,
				Name:          name,
				PublicPort:    uint16(bsPort),
				WebsocketPort: uint16(webSocketPort),
				OutOfBandPort: uint16(outOfBandPort),
			},
			uint16(relayBroadcastPort),
		); err != nil {
			return err
		}
	}
	select {}
}

func Execute() error {
	log.Printf("Arguments: %s", os.Args)
	var err error
	var defaultRelayBroadcastPort uint64
	if defaultRelayBroadcastPort, err = strconv.ParseUint(defaultRelayBroadcastPortStr, 10, 16); err != nil {
		return err
	}
	var defaultBsPublicPort uint64
	if defaultBsPublicPort, err = strconv.ParseUint(defaultBsPublicPortStr, 10, 16); err != nil {
		return err
	}
	if //goland:noinspection GoBoolExpressions
	defaultOutOfBandStr != "" {
		var defaultOutOfBandPort uint64
		if defaultOutOfBandPort, err = strconv.ParseUint(defaultOutOfBandStr, 10, 16); err != nil {
			return err
		}
		flag.UintVar(&outOfBandPort, "outOfBandPort", uint(defaultOutOfBandPort), "Out of Band port of the battle server")
	}
	// Unused
	flag.StringVar(&region, "region", "", "Region of the battle server")
	flag.StringVar(&name, "name", battleServer.DefaultName, "Name of the battle server")
	flag.UintVar(&publicPort, "publicPort", uint(defaultBsPublicPort), "Public port of the battle server")
	flag.UintVar(&relayBroadcastPort, "relaybroadcastPort", uint(defaultRelayBroadcastPort), "Relay broadcast port of the battle server")
	flag.UintVar(&bsPort, "bsPort", uint(defaultBsPublicPort), "Battle server port")
	flag.UintVar(&webSocketPort, "webSocketPort", 0, "WebSocket port of the battle server")
	// Unused
	flag.IntVar(&simulationPeriod, "simulationPeriod", 0, "Simulation period of the battle server")
	flag.StringVar(&sslCert, "sslCert", "", "Path to the SSL certificate file")
	flag.StringVar(&sslKey, "sslKey", "", "Path to the SSL key file")
	// TODO: Implement?
	/*
		flag.StringVar(&muninLog, "muninLog", "", "Path to the munin log file directory")
		flag.StringVar(&serverLog, "serverLog", "", "Path to the server log file directory")
	*/
	flag.Parse()
	return rootCmd()
}
