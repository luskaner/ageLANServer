package main

import (
	i "scripts/internal"

	e "github.com/luskaner/ageLANServer/common/executables"
)

func main() {
	i.CopyGameConfigs(e.BattleServerManager)
}
