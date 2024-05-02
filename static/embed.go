package static

import "embed"

//go:embed config/*.json
var Config embed.FS

//go:embed cloud/*
var Cloud embed.FS
