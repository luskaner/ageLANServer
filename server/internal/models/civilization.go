package models

import (
	"fmt"
	"strings"
)

var aoe2Civilizations = []string{
	"Aztecs", "Berbers", "Britons", "Bulgarians", "Burmese", "Byzantines",
	"Celts", "Chinese", "Cumans", "Ethiopians", "Franks", "Goths", "Huns",
	"Incas", "Hindustanis", "Italians", "Japanese", "Khmer", "Koreans",
	"Lithuanians", "Magyars", "Malay", "Malians", "Mayans", "Mongols",
	"Persians", "Portuguese", "Saracens", "Slavs", "Spanish", "Tatars",
	"Teutons", "Turks", "Vietnamese", "Vikings", "Burgundians", "Sicilians",
	"Poles", "Bohemians", "Bengalis", "Dravidians", "Gurjaras", "Romans",
	"Armenians", "Georgians",
}

func CivilizationName(gameID string, raceID int32) string {
	if gameID == "age2" && raceID >= 0 && int(raceID) < len(aoe2Civilizations) {
		return aoe2Civilizations[raceID]
	}
	if raceID < 0 {
		return "Unknown"
	}
	return fmt.Sprintf("Civilization %d", raceID)
}

func CivilizationIconPath(gameID string, raceID int32) string {
	name := CivilizationName(gameID, raceID)
	if strings.HasPrefix(name, "Civilization ") || name == "Unknown" {
		return "/static/images/civs/unknown.svg"
	}
	return "/static/images/civs/" + strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".png"
}
