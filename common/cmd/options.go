package cmd

import (
	"github.com/luskaner/ageLANServer/common"
)

const Name = `gameTitle`
const Shorthand = `e`
const DescriptionStart = `Game title`
const DescriptionEnd = `are supported.`

func SupportedGameSliceString() []string {
	slice := make([]string, len(common.SupportedGameTitleSlice))
	for i, v := range common.SupportedGameTitleSlice {
		slice[i] = string(v)
	}
	return slice
}
