package main

// ConfigFile is the bridge between Go and the React editor.
type ConfigFile struct {
	Path             string            `json:"path"`
	Content          string            `json:"content"`
	Config           *Configuration    `json:"config"`
	ValidationErrors []ValidationError `json:"validationErrors"`
}

type Configuration struct {
	Log                    bool               `toml:"Log" json:"Log"`
	GeneratePlatformUserID bool               `toml:"GeneratePlatformUserId" json:"GeneratePlatformUserId"`
	Authentication         string             `toml:"Authentication" json:"Authentication"`
	Games                  GamesConfiguration `toml:"Games" json:"Games"`
	Announcement           AnnouncementConfig `toml:"Announcement" json:"Announcement"`
}

type GamesConfiguration struct {
	Enabled []string   `toml:"Enabled" json:"Enabled"`
	Age1    GameConfig `toml:"age1" json:"age1"`
	Age2    GameConfig `toml:"age2" json:"age2"`
	Age3    GameConfig `toml:"age3" json:"age3"`
	Age4    GameConfig `toml:"age4" json:"age4"`
	Athens  GameConfig `toml:"athens" json:"athens"`
}

type GameConfig struct {
	Hosts []string `toml:"Hosts" json:"Hosts"`
}

type AnnouncementConfig struct {
	Enabled        bool   `toml:"Enabled" json:"Enabled"`
	Multicast      bool   `toml:"Multicast" json:"Multicast"`
	Port           int    `toml:"Port" json:"Port"`
	MulticastGroup string `toml:"MulticastGroup" json:"MulticastGroup"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
