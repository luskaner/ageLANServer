package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

var waitBeforeRunning time.Duration
var exitBeforeRunning bool
var executableGame string

func rootCmd(args []string) error {
	if exitBeforeRunning {
		log.Printf("Exiting before running the game")
		return nil
	}
	log.Println("Waiting before running the game...")
	time.Sleep(waitBeforeRunning)
	log.Println("Done waiting, launching the game...")
	options := exec.Options{
		File:       executableGame,
		ShowWindow: true,
		Pid:        true,
		Args:       args,
	}
	if result := options.Exec(); !result.Success() {
		return fmt.Errorf("failed to start the game: %s", result.Err)
	} else {
		log.Printf("Started the game with PID %d\n", result.Pid)
	}
	return nil
}

	log.Printf("Arguments: %v", os.Args)
	flag.DurationVar(&waitBeforeRunning, "waitBeforeRunning", 10*time.Second, "Wait time before running the game")
	flag.Parse()
	return rootCmd(flag.Args())
}
