package crossover

const suffixBaseDir = "/Library/Application Support/CrossOver/Bottles"

var baseDirs = []string{
	// TODO: Add system wide installation
	"$HOME" + suffixBaseDir,
}

func defaultBottleName(gameId string) (name string) {
	return baseDefaultBottleName(gameId)
}
