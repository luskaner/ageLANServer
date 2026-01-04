package main

import "scripts/internal/goreleaser"

func main() {
	/*project := config.Project{
		Version: 2,
	}
	marshal, err := yaml.Marshal(&project)
	if err != nil {
		return
	}
	println(string(marshal))*/
	goreleaser.Generate()
}
