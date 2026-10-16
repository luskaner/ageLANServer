package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfiguration(t *testing.T) {
	config, err := parseConfiguration([]byte(`
Log = true
GeneratePlatformUserId = false
Authentication = 'cached'

[Games]
Enabled = ['age2']

[Games.age2]
Hosts = ['192.168.1.10']

[Announcement]
Enabled = true
Multicast = true
Port = 32000
MulticastGroup = '239.31.97.8'
`))
	if err != nil {
		t.Fatalf("parseConfiguration returned an error: %v", err)
	}
	if !config.Log || config.Authentication != "cached" || config.Announcement.Port != 32000 {
		t.Fatalf("unexpected parsed configuration: %+v", config)
	}
	if len(config.Games.Enabled) != 1 || config.Games.Enabled[0] != "age2" {
		t.Fatalf("unexpected enabled games: %#v", config.Games.Enabled)
	}
	if len(config.Games.Age2.Hosts) != 1 || config.Games.Age2.Hosts[0] != "192.168.1.10" {
		t.Fatalf("unexpected hosts: %#v", config.Games.Age2.Hosts)
	}
}

func TestParseConfigurationRejectsInvalidTOML(t *testing.T) {
	if _, err := parseConfiguration([]byte("[Announcement\nPort = 32000")); err == nil {
		t.Fatal("parseConfiguration accepted invalid TOML")
	}
}

func TestParseConfigurationRejectsInvalidType(t *testing.T) {
	if _, err := parseConfiguration([]byte("[Announcement]\nPort = 'not-a-port'")); err == nil {
		t.Fatal("parseConfiguration accepted an invalid port type")
	}
}

func TestContainsConfigurationField(t *testing.T) {
	if hasFields, err := containsConfigurationField([]byte("[other]\nValue = true")); err != nil || hasFields {
		t.Fatalf("unrelated TOML was identified as configuration: fields=%v, err=%v", hasFields, err)
	}
	if hasFields, err := containsConfigurationField([]byte("[Announcement]\nPort = 0")); err != nil || !hasFields {
		t.Fatalf("invalid known field was not identified: fields=%v, err=%v", hasFields, err)
	}
}

func TestReadConfigFileAcceptsInvalidKnownValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := []byte("Authentication = 'unsupported'\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("create test configuration: %v", err)
	}
	file, err := (&ConfigService{}).readConfigFile(path)
	if err != nil {
		t.Fatalf("readConfigFile rejected a known but invalid value: %v", err)
	}
	if len(file.ValidationErrors) == 0 || file.ValidationErrors[0].Field != "Authentication" {
		t.Fatalf("unexpected validation errors: %#v", file.ValidationErrors)
	}
}

func TestReadConfigFileRejectsUnrelatedTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[other]\nValue = true\n"), 0600); err != nil {
		t.Fatalf("create test configuration: %v", err)
	}
	if _, err := (&ConfigService{}).readConfigFile(path); err == nil {
		t.Fatal("readConfigFile accepted TOML without supported configuration fields")
	}
}

func TestValidateConfiguration(t *testing.T) {
	cases := []struct {
		name      string
		config    Configuration
		field     string
		hasErrors bool
	}{
		{
			name: "valid",
			config: Configuration{
				Authentication: "disabled",
				Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
				Announcement:   AnnouncementConfig{Port: 31978},
			},
		},
		{
			name: "unknown authentication",
			config: Configuration{
				Authentication: "unknown",
				Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
				Announcement:   AnnouncementConfig{Port: 31978},
			},
			field:     "Authentication",
			hasErrors: true,
		},
		{
			name: "empty games and invalid port",
			config: Configuration{
				Authentication: "disabled",
				Announcement:   AnnouncementConfig{Port: 65536},
			},
			field:     "Games.Enabled",
			hasErrors: true,
		},
		{
			name: "unknown game",
			config: Configuration{
				Authentication: "disabled",
				Games:          GamesConfiguration{Enabled: []string{"age5"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
				Announcement:   AnnouncementConfig{Port: 31978},
			},
			field:     "Games.Enabled.0",
			hasErrors: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			errors := validateConfiguration(&testCase.config)
			if !testCase.hasErrors && len(errors) != 0 {
				t.Fatalf("validateConfiguration returned unexpected errors: %#v", errors)
			}
			if testCase.hasErrors && len(errors) == 0 {
				t.Fatal("validateConfiguration accepted invalid configuration")
			}
			if testCase.field != "" && errors[0].Field != testCase.field {
				t.Fatalf("first error field = %q, want %q", errors[0].Field, testCase.field)
			}
		})
	}
}

func TestValidateConfigurationMulticastIsConditional(t *testing.T) {
	config := Configuration{
		Authentication: "disabled",
		Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
		Announcement: AnnouncementConfig{
			Enabled:        true,
			Multicast:      true,
			Port:           31978,
			MulticastGroup: "192.168.1.1",
		},
	}
	errors := validateConfiguration(&config)
	if len(errors) != 1 || errors[0].Field != "Announcement.MulticastGroup" {
		t.Fatalf("unexpected multicast errors: %#v", errors)
	}

	config.Announcement.Multicast = false
	if errors := validateConfiguration(&config); len(errors) != 0 {
		t.Fatalf("disabled multicast returned errors: %#v", errors)
	}
}

func TestReplaceAnnouncementPortPreservesDocument(t *testing.T) {
	original := []byte("# Keep this comment.\nLog = false\n\n[Announcement] # Keep this section comment.\nPort = 31978 # Keep this inline comment.\nUnknown = 'keep me'\n")
	updated, err := replaceDocumentValue(original, "Announcement", "Port", "32000")
	if err != nil {
		t.Fatalf("replaceDocumentValue returned an error: %v", err)
	}
	if !strings.Contains(string(updated), "# Keep this comment.") ||
		!strings.Contains(string(updated), "# Keep this inline comment.") ||
		!strings.Contains(string(updated), "Unknown = 'keep me'") {
		t.Fatalf("replacement did not preserve the original document: %s", updated)
	}
	if !strings.Contains(string(updated), "Port = 32000 # Keep this inline comment.") {
		t.Fatalf("replacement did not update the port: %s", updated)
	}
	config, err := parseConfiguration(updated)
	if err != nil {
		t.Fatalf("updated document could not be parsed: %v", err)
	}
	if config.Announcement.Port != 32000 {
		t.Fatalf("updated port = %d, want 32000", config.Announcement.Port)
	}
}

func TestWriteConfigFileReplacesExistingFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.toml")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatalf("create test configuration: %v", err)
	}
	if err := writeConfigFile(path, []byte("new")); err != nil {
		t.Fatalf("writeConfigFile returned an error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read replaced configuration: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("replaced content = %q, want %q", content, "new")
	}
}

func TestSaveConfigurationDoesNotOverwriteInvalidConfiguration(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.toml")
	original := []byte(`Log = false
GeneratePlatformUserId = false
Authentication = 'disabled'

[Games]
Enabled = ['age2']

[Games.age2]
Hosts = ['127.0.0.1']

[Announcement]
Enabled = true
Multicast = false
Port = 31978
MulticastGroup = '239.31.97.8'
`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatalf("create test configuration: %v", err)
	}

	configuration := Configuration{
		Authentication: "disabled",
		Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
		Announcement:   AnnouncementConfig{Port: 0},
	}
	if err := (&ConfigService{}).SaveConfiguration(path, configuration); err == nil {
		t.Fatal("SaveConfiguration accepted an invalid configuration")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unchanged configuration: %v", err)
	}
	if string(content) != string(original) {
		t.Fatalf("invalid save changed the configuration: %s", content)
	}
}

func TestSaveConfigurationReportsMissingPath(t *testing.T) {
	configuration := Configuration{Authentication: "disabled"}
	if err := (&ConfigService{}).SaveConfiguration(filepath.Join(t.TempDir(), "missing.toml"), configuration); err == nil {
		t.Fatal("SaveConfiguration accepted a missing path")
	}
}

func TestSaveConfigurationAsPreservesSourceAndRequiresOverwriteConfirmation(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "source.toml")
	destinationPath := filepath.Join(directory, "destination.toml")
	original := []byte(`# Preserve this comment.
	Log = false
	GeneratePlatformUserId = false
	Authentication = 'disabled'

	[Games]
	Enabled = ['age2']

	[Games.age2]
	Hosts = ['127.0.0.1']

	[Games.age1]
	Hosts = ['127.0.0.2']

	[Games.age3]
	Hosts = ['127.0.0.3']

	[Games.age4]
	Hosts = ['127.0.0.4']

	[Games.athens]
	Hosts = ['127.0.0.5']

	[Announcement]
	Enabled = true
	Multicast = false
	Port = 31978
	MulticastGroup = '239.31.97.8'
	Unknown = 'preserve me'
	`)
	if err := os.WriteFile(sourcePath, original, 0600); err != nil {
		t.Fatalf("create source configuration: %v", err)
	}
	if err := os.WriteFile(destinationPath, []byte("existing destination"), 0600); err != nil {
		t.Fatalf("create destination configuration: %v", err)
	}
	configuration := Configuration{
		Authentication: "disabled",
		Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
		Announcement:   AnnouncementConfig{Port: 32000, MulticastGroup: "239.31.97.8"},
	}
	service := &ConfigService{}
	if _, err := service.SaveConfigurationAs(sourcePath, destinationPath, configuration, false); err == nil || !strings.Contains(err.Error(), "confirmation required") {
		t.Fatalf("SaveConfigurationAs error = %v, want overwrite confirmation", err)
	}
	file, err := service.SaveConfigurationAs(sourcePath, destinationPath, configuration, true)
	if err != nil {
		t.Fatalf("SaveConfigurationAs returned an error: %v", err)
	}
	if file.Path != destinationPath || file.Config.Announcement.Port != 32000 {
		t.Fatalf("unexpected saved configuration: %+v", file)
	}
	if !strings.Contains(file.Content, "# Preserve this comment.") || !strings.Contains(file.Content, "Unknown = 'preserve me'") {
		t.Fatalf("save-as did not preserve unedited content: %s", file.Content)
	}
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source configuration: %v", err)
	}
	if string(sourceContent) != string(original) {
		t.Fatalf("save-as changed source configuration: %s", sourceContent)
	}
}

func TestWriteNewConfigFileDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")
	if err := writeNewConfigFile(path, []byte("new")); err != nil {
		t.Fatalf("writeNewConfigFile returned an error: %v", err)
	}
	if err := writeNewConfigFile(path, []byte("replacement")); err == nil {
		t.Fatal("writeNewConfigFile overwrote an existing file")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created configuration: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("created content = %q, want %q", content, "new")
	}
}

func TestCreateConfigFileUsesOfficialTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	configuration := Configuration{
		Authentication: "disabled",
		Games: GamesConfiguration{
			Enabled: []string{"age2"},
			Age2:    GameConfig{Hosts: []string{"127.0.0.1"}},
		},
		Announcement: AnnouncementConfig{
			Enabled:        true,
			Multicast:      false,
			Port:           32000,
			MulticastGroup: "239.31.97.8",
		},
	}
	templatePath, err := templateConfigPath()
	if err != nil {
		t.Fatalf("templateConfigPath returned an error: %v", err)
	}
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	updated, err := configurationDocument(templateContent, configuration)
	if err != nil {
		t.Fatalf("configurationDocument returned an error: %v", err)
	}
	if _, err := parseConfiguration(updated); err != nil {
		t.Fatalf("configurationDocument produced invalid TOML:\n%s\n%v", updated, err)
	}
	file, err := (&ConfigService{}).CreateConfigFile(path, configuration, false)
	if err != nil {
		t.Fatalf("CreateConfigFile returned an error: %v", err)
	}
	if file.Path != path || file.Config.Announcement.Port != 32000 {
		t.Fatalf("unexpected created file: %+v", file)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created configuration: %v", err)
	}
	if !strings.Contains(string(content), "# Whether to log all the info") {
		t.Fatal("created configuration did not preserve the official template comments")
	}
}

func TestCreateConfigFileRequiresOverwriteConfirmation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("existing"), 0600); err != nil {
		t.Fatalf("create existing configuration: %v", err)
	}
	configuration := Configuration{
		Authentication: "disabled",
		Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
		Announcement:   AnnouncementConfig{Port: 31978},
	}
	if _, err := (&ConfigService{}).CreateConfigFile(path, configuration, false); err == nil {
		t.Fatal("CreateConfigFile did not require overwrite confirmation")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing configuration: %v", err)
	}
	if string(content) != "existing" {
		t.Fatalf("existing content changed after rejected creation: %q", content)
	}
}
