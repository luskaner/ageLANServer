package main

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal/cmd"
)

const version = "development"

func main() {
	if executor.IsAdmin() {
		fmt.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when it needs.")
	}
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
	cmd.Version = version
}
