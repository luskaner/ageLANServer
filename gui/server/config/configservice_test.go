package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseConfiguration(t *testing.T) {
	config, err := parseConfiguration([]byte(`
Log = true
GeneratePlatformUserId = false
Authentication = 'cached'

[Sign]
Enabled = true
PlatformUserIdExceptions = []

[Internet]
Enabled = false
IP = '203.0.113.10'

[Games]
Enabled = ['age2']

[Games.age2]
Hosts = ['192.168.1.10']

[Announcement]
Enabled = true
Multicast = true
Port = 32000
MulticastGroup = '239.31.97.8'

[CustomFiles]
Enabled = false
Folder = 'D:\saves'
MaxFileSize = '64MB'
MaxTotalSize = '1GB'
Expiry = '30m'
`))
	if err != nil {
		t.Fatalf("parseConfiguration returned an error: %v", err)
	}
	if !config.Log || config.Authentication != "cached" || config.Announcement.Port != 32000 {
		t.Fatalf("unexpected parsed configuration: %+v", config)
	}
	if !config.Sign.Enabled {
		t.Fatal("Sign.Enabled was not parsed")
	}
	if config.Internet.Enabled || config.Internet.IP != "203.0.113.10" {
		t.Fatalf("unexpected internet configuration: %#v", config.Internet)
	}
	if config.CustomFiles.Enabled || config.CustomFiles.Folder != `D:\saves` ||
		config.CustomFiles.MaxFileSize != "64MB" || config.CustomFiles.MaxTotalSize != "1GB" ||
		config.CustomFiles.Expiry != "30m" {
		t.Fatalf("unexpected custom files configuration: %#v", config.CustomFiles)
	}
	if len(config.Games.Enabled) != 1 || config.Games.Enabled[0] != "age2" {
		t.Fatalf("unexpected enabled games: %#v", config.Games.Enabled)
	}
	if len(config.Games.Age2.Hosts) != 1 || config.Games.Age2.Hosts[0] != "192.168.1.10" {
		t.Fatalf("unexpected hosts: %#v", config.Games.Age2.Hosts)
	}
}

func TestParseConfigurationAppliesServerDefaults(t *testing.T) {
	config, err := parseConfiguration([]byte("Log = true\n"))
	if err != nil {
		t.Fatalf("parseConfiguration returned an error: %v", err)
	}
	if config.Authentication != "disabled" {
		t.Fatalf("Authentication = %q, want the server default", config.Authentication)
	}
	if !config.Internet.Enabled || config.Internet.IP != "auto" {
		t.Fatalf("unexpected internet defaults: %#v", config.Internet)
	}
	if !config.Announcement.Enabled || !config.Announcement.Multicast ||
		config.Announcement.Port != 31978 || config.Announcement.MulticastGroup != "239.31.97.8" {
		t.Fatalf("unexpected announcement defaults: %#v", config.Announcement)
	}
	if !config.CustomFiles.Enabled || config.CustomFiles.Folder != "auto" ||
		config.CustomFiles.MaxFileSize != "32MB" || config.CustomFiles.MaxTotalSize != "256MB" ||
		config.CustomFiles.Expiry != "12h" {
		t.Fatalf("unexpected custom files defaults: %#v", config.CustomFiles)
	}
	if len(config.Games.Age2.Hosts) != 1 || config.Games.Age2.Hosts[0] != "0.0.0.0" {
		t.Fatalf("unexpected host defaults: %#v", config.Games.Age2.Hosts)
	}
}

func TestParseConfigurationKeepsExplicitFalseValues(t *testing.T) {
	config, err := parseConfiguration([]byte("[Internet]\nEnabled = false\n\n[CustomFiles]\nEnabled = false\n"))
	if err != nil {
		t.Fatalf("parseConfiguration returned an error: %v", err)
	}
	if config.Internet.Enabled || config.CustomFiles.Enabled {
		t.Fatalf("explicit disabled switches were replaced by defaults: %#v", config)
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
		name       string
		config     Configuration
		field      string
		hasErrors  bool
		hasWarning bool
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
			name: "empty games list only warns",
			config: Configuration{
				Authentication: "disabled",
				Announcement:   AnnouncementConfig{Port: 31978},
			},
			field:      "Games.Enabled",
			hasWarning: true,
		},
		{
			name: "invalid port",
			config: Configuration{
				Authentication: "disabled",
				Announcement:   AnnouncementConfig{Port: 65536},
			},
			field:     "Announcement.Port",
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
			issues := validateConfiguration(&testCase.config)
			errors := errorValidationIssues(issues)
			if !testCase.hasErrors && len(errors) != 0 {
				t.Fatalf("validateConfiguration returned unexpected errors: %#v", errors)
			}
			if testCase.hasErrors && len(errors) == 0 {
				t.Fatal("validateConfiguration accepted invalid configuration")
			}
			if testCase.field != "" && len(errors) > 0 && errors[0].Field != testCase.field {
				t.Fatalf("first error field = %q, want %q", errors[0].Field, testCase.field)
			}
			if testCase.hasErrors && len(errors) > 0 && errors[0].Severity != severityError {
				t.Fatalf("first error severity = %q, want %q", errors[0].Severity, severityError)
			}
			if testCase.hasWarning {
				if len(issues) == 0 {
					t.Fatal("validateConfiguration did not warn")
				}
				warning := issues[0]
				if warning.Field != testCase.field {
					t.Fatalf("warning field = %q, want %q", warning.Field, testCase.field)
				}
				if warning.Severity != severityWarning {
					t.Fatalf("warning severity = %q, want %q", warning.Severity, severityWarning)
				}
				if len(errors) != 0 {
					t.Fatalf("the warning blocked saving: %#v", errors)
				}
			}
		})
	}
}

func TestValidateConfigurationDuplicateHosts(t *testing.T) {
	// Two enabled games cannot bind the same address: it is an error on both
	// sides, as neither of them could be served.
	config := Configuration{
		Authentication: "disabled",
		Games: GamesConfiguration{
			Enabled: []string{"age1", "age2"},
			Age1:    GameConfig{Hosts: []string{"192.0.2.10"}},
			Age2:    GameConfig{Hosts: []string{"192.0.2.10"}},
		},
		Announcement: AnnouncementConfig{Port: 31978},
	}
	issues := validateConfiguration(&config)
	errors := errorValidationIssues(issues)
	if len(errors) != 2 {
		t.Fatalf("the shared address was not reported on both games: %#v", issues)
	}
	if errors[0].Field != "Games.age1.Hosts.0" || errors[1].Field != "Games.age2.Hosts.0" {
		t.Fatalf("unexpected error fields: %#v", errors)
	}
	for _, issue := range issues {
		if issue.Severity == severityWarning {
			t.Fatalf("the conflict between enabled games warned: %#v", issues)
		}
	}

	// A game that is not enabled is not bound, so it does not take the address
	// away from the enabled one: only the disabled one is told about it.
	config = Configuration{
		Authentication: "disabled",
		Games: GamesConfiguration{
			Enabled: []string{"age1"},
			Age1:    GameConfig{Hosts: []string{"192.0.2.10"}},
			Age2:    GameConfig{Hosts: []string{"192.0.2.10"}},
		},
		Announcement: AnnouncementConfig{Port: 31978},
	}
	issues = validateConfiguration(&config)
	if errors := errorValidationIssues(issues); len(errors) != 0 {
		t.Fatalf("a disabled game blocked the enabled one: %#v", errors)
	}
	if len(issues) != 1 || issues[0].Field != "Games.age2.Hosts.0" || issues[0].Severity != severityWarning {
		t.Fatalf("the disabled game was not warned about: %#v", issues)
	}

	// Two disabled games do not conflict at all, they only warn about it.
	config = Configuration{
		Authentication: "disabled",
		Games: GamesConfiguration{
			Enabled: []string{"age1"},
			Age1:    GameConfig{Hosts: []string{"192.0.2.10"}},
			Age2:    GameConfig{Hosts: []string{"192.0.2.10"}},
			Age3:    GameConfig{Hosts: []string{"192.0.2.10"}},
		},
		Announcement: AnnouncementConfig{Port: 31978},
	}
	issues = validateConfiguration(&config)
	if errors := errorValidationIssues(issues); len(errors) != 0 {
		t.Fatalf("the disabled games blocked saving: %#v", errors)
	}
	if len(issues) != 2 {
		t.Fatalf("the disabled games were not warned about: %#v", issues)
	}
	for _, issue := range issues {
		if issue.Severity != severityWarning {
			t.Fatalf("a disabled game reported an error: %#v", issues)
		}
	}
}
func TestValidateConfigurationPrivilegedPortWarnsWithoutBlocking(t *testing.T) {
	config := Configuration{
		Authentication: "disabled",
		Games:          GamesConfiguration{Enabled: []string{"age2"}, Age2: GameConfig{Hosts: []string{"127.0.0.1"}}},
		Announcement:   AnnouncementConfig{Enabled: true, Port: 80},
	}
	issues := validateConfiguration(&config)
	if len(issues) != 1 || issues[0].Field != "Announcement.Port" || issues[0].Severity != severityWarning {
		t.Fatalf("unexpected privileged port issues: %#v", issues)
	}
	if blocking := errorValidationIssues(issues); len(blocking) != 0 {
		t.Fatalf("privileged port blocked saving: %#v", blocking)
	}

	// The port is out of play while the announcement is disabled.
	config.Announcement.Enabled = false
	if issues := validateConfiguration(&config); len(issues) != 0 {
		t.Fatalf("disabled announcement returned issues: %#v", issues)
	}

	// A port outside the valid range is an error, not a warning.
	config.Announcement.Enabled = true
	config.Announcement.Port = 0
	issues = validateConfiguration(&config)
	if len(issues) != 1 || issues[0].Severity != severityError {
		t.Fatalf("unexpected invalid port issues: %#v", issues)
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

func TestValidateConfigurationInternet(t *testing.T) {
	config := serverDefaults()
	config.Games.Enabled = []string{"age2"}
	config.Games.Age2.Hosts = []string{"127.0.0.1"}
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("defaults were not valid: %#v", issues)
	}

	config.Internet.IP = "not-an-address"
	issues := errorValidationIssues(validateConfiguration(&config))
	if len(issues) != 1 || issues[0].Field != "Internet.IP" || issues[0].Severity != severityError {
		t.Fatalf("unexpected internet address issues: %#v", issues)
	}

	config.Internet.IP = "203.0.113.10"
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("a manual IPv4 address was rejected: %#v", issues)
	}

	// The address is out of play while the server does not use the Internet.
	config.Internet.Enabled = false
	config.Internet.IP = "not-an-address"
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("disabled internet still validated its address: %#v", issues)
	}

	// Authentication cannot be required without Internet access: the server
	// refuses to start in that combination.
	config.Authentication = "required"
	issues = errorValidationIssues(validateConfiguration(&config))
	if len(issues) != 1 || issues[0].Field != "Authentication" {
		t.Fatalf("unexpected required authentication issues: %#v", issues)
	}
}

func TestValidateConfigurationCustomFiles(t *testing.T) {
	config := serverDefaults()
	config.Games.Enabled = []string{"age2"}
	config.Games.Age2.Hosts = []string{"127.0.0.1"}

	config.CustomFiles.MaxFileSize = "32"
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("a plain byte count was rejected: %#v", issues)
	}

	config.CustomFiles.MaxFileSize = "32 MB"
	config.CustomFiles.MaxTotalSize = "1GB"
	config.CustomFiles.Expiry = "1h30m"
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("valid sizes and durations were rejected: %#v", issues)
	}

	config.CustomFiles.MaxFileSize = "big"
	issues := errorValidationIssues(validateConfiguration(&config))
	if len(issues) != 1 || issues[0].Field != "CustomFiles.MaxFileSize" {
		t.Fatalf("unexpected file size issues: %#v", issues)
	}

	config.CustomFiles.MaxFileSize = "2GB"
	config.CustomFiles.MaxTotalSize = "1GB"
	issues = errorValidationIssues(validateConfiguration(&config))
	if len(issues) != 1 || issues[0].Field != "CustomFiles.MaxFileSize" {
		t.Fatalf("unexpected size relation issues: %#v", issues)
	}

	config.CustomFiles.MaxFileSize = "32MB"
	config.CustomFiles.MaxTotalSize = "256MB"
	config.CustomFiles.Folder = "  "
	config.CustomFiles.Expiry = "12"
	issues = errorValidationIssues(validateConfiguration(&config))
	if len(issues) != 2 ||
		issues[0].Field != "CustomFiles.Folder" || issues[1].Field != "CustomFiles.Expiry" {
		t.Fatalf("unexpected folder or expiry issues: %#v", issues)
	}

	// Disabled endpoints never use these values, so they do not block saving.
	config.CustomFiles.Enabled = false
	config.CustomFiles.MaxFileSize = "big"
	config.CustomFiles.MaxTotalSize = "bigger"
	config.CustomFiles.Expiry = "12"
	config.CustomFiles.Folder = ""
	if issues := errorValidationIssues(validateConfiguration(&config)); len(issues) != 0 {
		t.Fatalf("disabled custom files returned issues: %#v", issues)
	}
}

func TestConfigurationDocumentAddsMissingSettings(t *testing.T) {
	original := []byte("Log = false\n\n[Announcement]\nEnabled = true\nMulticast = false\nPort = 31978\nMulticastGroup = '239.31.97.8'\n")
	configuration := serverDefaults()
	configuration.CustomFiles.MaxFileSize = "64MB"

	updated, err := configurationDocument(original, configuration)
	if err != nil {
		t.Fatalf("configurationDocument returned an error: %v", err)
	}
	if strings.Count(string(updated), "[CustomFiles]") != 1 || strings.Count(string(updated), "[Internet]") != 1 {
		t.Fatalf("tables were duplicated or not added:\n%s", updated)
	}
	if strings.Index(string(updated), "Sign = false") > strings.Index(string(updated), "[Announcement]") {
		t.Fatalf("a root key was written inside a table:\n%s", updated)
	}
	parsed, err := parseConfiguration(updated)
	if err != nil {
		t.Fatalf("updated document could not be parsed: %v\n%s", err, updated)
	}
	if parsed.Sign.Enabled || !parsed.Internet.Enabled || parsed.Internet.IP != "auto" {
		t.Fatalf("unexpected internet configuration: %#v", parsed.Internet)
	}
	if parsed.CustomFiles.MaxFileSize != "64MB" || parsed.CustomFiles.Expiry != "12h" {
		t.Fatalf("unexpected custom files configuration: %#v", parsed.CustomFiles)
	}
	if len(parsed.Games.Age2.Hosts) != 1 || parsed.Games.Age2.Hosts[0] != "0.0.0.0" {
		t.Fatalf("unexpected default hosts: %#v", parsed.Games.Age2.Hosts)
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

func TestConfigurationDocumentRoundTripsTheOfficialTemplate(t *testing.T) {
	templatePath, err := templateConfigPath()
	if err != nil {
		t.Fatalf("templateConfigPath returned an error: %v", err)
	}
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	original, err := parseConfiguration(templateContent)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	updated, err := configurationDocument(templateContent, *original)
	if err != nil {
		t.Fatalf("configurationDocument returned an error: %v", err)
	}
	roundTripped, err := parseConfiguration(updated)
	if err != nil {
		t.Fatalf("round-tripped template is invalid: %v\n%s", err, updated)
	}
	if !reflect.DeepEqual(*roundTripped, *original) {
		t.Fatalf("round trip changed the configuration:\noriginal: %#v\nupdated:  %#v", *original, *roundTripped)
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
