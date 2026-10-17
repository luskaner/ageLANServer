package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const configFileName = "config.toml"

type ConfigService struct{}

// OpenConfigFile finds the default file or asks the user to select one, then
// returns its path and raw contents. Parsing and editing are deliberately not
// part of this first loading step.
func (s *ConfigService) OpenConfigFile() (*ConfigFile, error) {
	path, err := defaultConfigPath()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return s.SelectConfigFile()
	}
	return s.readConfigFile(path)
}

// SelectConfigFile always opens the native picker, even when a default file
// exists. It is used when the user explicitly wants another configuration.
func (s *ConfigService) SelectConfigFile() (*ConfigFile, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Select configuration file").
		CanChooseFiles(true).
		CanChooseDirectories(false).
		AddFilter("TOML files", "*.toml").
		PromptForSingleSelection()
	if err != nil {
		return nil, fmt.Errorf("open configuration picker: %w", err)
	}
	if path == "" {
		return nil, errors.New("no configuration file selected")
	}
	return s.readConfigFile(path)
}

// OpenConfigFileAtPath loads a configuration dropped onto the application window.
func (s *ConfigService) OpenConfigFileAtPath(path string) (*ConfigFile, error) {
	if path == "" {
		return nil, errors.New("dropped configuration path is empty")
	}
	return s.readConfigFile(path)
}

// OpenTemplate loads the official server template as a new, unsaved document.
func (s *ConfigService) OpenTemplate() (*ConfigFile, error) {
	path, err := templateConfigPath()
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration template %q: %w", path, err)
	}
	config, err := parseConfiguration(content)
	if err != nil {
		return nil, fmt.Errorf("parse configuration template %q: %w", path, err)
	}
	return &ConfigFile{
		Content:          string(content),
		Config:           config,
		ValidationErrors: validateConfiguration(config),
	}, nil
}

// SelectConfigDestination opens the native dialog used for a new configuration.
func (s *ConfigService) SelectConfigDestination() (string, error) {
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Choose where to create the configuration").
		SetFilename(configFileName).
		AddFilter("TOML files", "*.toml").
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("open configuration destination picker: %w", err)
	}
	if path == "" {
		return "", errors.New("no configuration destination selected")
	}
	return path, nil
}

// SelectConfigSaveAsDestination opens a native dialog for an existing configuration's new path.
func (s *ConfigService) SelectConfigSaveAsDestination(fileName string) (string, error) {
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Choose where to save the configuration copy").
		SetFilename(filepath.Base(fileName)).
		AddFilter("TOML files", "*.toml").
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("open configuration save-as picker: %w", err)
	}
	if path == "" {
		return "", errors.New("no configuration destination selected")
	}
	return path, nil
}

// CreateConfigFile validates a new document and creates it at the requested path.
func (s *ConfigService) CreateConfigFile(path string, configuration Configuration, overwrite bool) (*ConfigFile, error) {
	template, err := s.OpenTemplate()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, errors.New("configuration destination cannot be empty")
	}
	if _, err := os.Stat(path); err == nil && !overwrite {
		return nil, fmt.Errorf("configuration file %q already exists; confirmation required", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("check configuration destination %q: %w", path, err)
	}
	if validationErrors := validateConfiguration(&configuration); len(validationErrors) > 0 {
		return nil, fmt.Errorf("configuration is invalid: %s", validationErrors[0].Message)
	}
	updated, err := configurationDocument([]byte(template.Content), configuration)
	if err != nil {
		return nil, fmt.Errorf("create configuration file %q: %w", path, err)
	}
	if overwrite {
		if err := writeConfigFile(path, updated); err != nil {
			return nil, err
		}
	} else if err := writeNewConfigFile(path, updated); err != nil {
		return nil, err
	}
	return s.readConfigFile(path)
}

// SaveConfiguration validates and saves all editable configuration fields.
func (s *ConfigService) SaveConfiguration(path string, configuration Configuration) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read configuration file %q: %w", path, err)
	}
	config, err := parseConfiguration(content)
	if err != nil {
		return fmt.Errorf("parse configuration file %q: %w", path, err)
	}
	config = &configuration
	if validationErrors := validateConfiguration(config); len(validationErrors) > 0 {
		return fmt.Errorf("configuration is invalid: %s", validationErrors[0].Message)
	}

	updated, err := configurationDocument(content, configuration)
	if err != nil {
		return fmt.Errorf("edit configuration file %q: %w", path, err)
	}
	if _, err := parseConfiguration(updated); err != nil {
		return fmt.Errorf("validate edited configuration file %q: %w", path, err)
	}
	return writeConfigFile(path, updated)
}

// PreviewConfiguration returns the edited TOML document without writing it.
func (s *ConfigService) PreviewConfiguration(content string, configuration Configuration) (string, error) {
	if _, err := parseConfiguration([]byte(content)); err != nil {
		return "", fmt.Errorf("parse configuration document: %w", err)
	}
	updated, err := configurationDocument([]byte(content), configuration)
	if err != nil {
		return "", fmt.Errorf("preview configuration document: %w", err)
	}
	if _, err := parseConfiguration(updated); err != nil {
		return "", fmt.Errorf("validate preview configuration document: %w", err)
	}
	return string(updated), nil
}

// SaveConfigurationAs validates and saves an edited configuration to a new path.
func (s *ConfigService) SaveConfigurationAs(sourcePath string, destination string, configuration Configuration, overwrite bool) (*ConfigFile, error) {
	if sourcePath == "" || destination == "" {
		return nil, errors.New("configuration source and destination are required")
	}
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read configuration file %q: %w", sourcePath, err)
	}
	if _, err := parseConfiguration(content); err != nil {
		return nil, fmt.Errorf("parse configuration file %q: %w", sourcePath, err)
	}
	if validationErrors := validateConfiguration(&configuration); len(validationErrors) > 0 {
		return nil, fmt.Errorf("configuration is invalid: %s", validationErrors[0].Message)
	}
	updated, err := configurationDocument(content, configuration)
	if err != nil {
		return nil, fmt.Errorf("edit configuration file %q: %w", sourcePath, err)
	}
	if _, err := parseConfiguration(updated); err != nil {
		return nil, fmt.Errorf("validate edited configuration file: %w", err)
	}

	if _, err := os.Stat(destination); err == nil && !overwrite {
		return nil, fmt.Errorf("configuration file %q already exists; confirmation required", destination)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("check configuration destination %q: %w", destination, err)
	}
	if overwrite {
		if err := writeConfigFile(destination, updated); err != nil {
			return nil, err
		}
	} else if err := writeNewConfigFile(destination, updated); err != nil {
		return nil, err
	}
	return s.readConfigFile(destination)
}

// configurationDocument applies editable values without serializing the whole model.
func configurationDocument(content []byte, configuration Configuration) ([]byte, error) {
	updated := content
	replacements := []struct {
		section string
		key     string
		value   string
	}{
		{"", "Log", strconv.FormatBool(configuration.Log)},
		{"", "GeneratePlatformUserId", strconv.FormatBool(configuration.GeneratePlatformUserID)},
		{"", "Authentication", "'" + configuration.Authentication + "'"},
		{"Games", "Enabled", tomlArray(configuration.Games.Enabled)},
		{"Games.age1", "Hosts", tomlArray(configuration.Games.Age1.Hosts)},
		{"Games.age2", "Hosts", tomlArray(configuration.Games.Age2.Hosts)},
		{"Games.age3", "Hosts", tomlArray(configuration.Games.Age3.Hosts)},
		{"Games.age4", "Hosts", tomlArray(configuration.Games.Age4.Hosts)},
		{"Games.athens", "Hosts", tomlArray(configuration.Games.Athens.Hosts)},
		{"Announcement", "Enabled", strconv.FormatBool(configuration.Announcement.Enabled)},
		{"Announcement", "Multicast", strconv.FormatBool(configuration.Announcement.Multicast)},
		{"Announcement", "Port", strconv.Itoa(configuration.Announcement.Port)},
		{"Announcement", "MulticastGroup", "'" + configuration.Announcement.MulticastGroup + "'"},
	}
	for _, replacement := range replacements {
		var err error
		updated, err = replaceDocumentValue(updated, replacement.section, replacement.key, replacement.value)
		if err != nil {
			return nil, fmt.Errorf("replace %s.%s: %w", replacement.section, replacement.key, err)
		}
	}
	return updated, nil
}

// tomlArray formats string values as a TOML array.
func tomlArray(values []string) string {
	quoted := make([]string, len(values))
	for index, value := range values {
		quoted[index] = strconv.Quote(value)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// readConfigFile reads, parses and validates a configuration file from disk.
func (s *ConfigService) readConfigFile(path string) (*ConfigFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration file %q: %w", path, err)
	}
	config, err := parseConfiguration(content)
	if err != nil {
		return nil, fmt.Errorf("parse configuration file %q: %w", path, err)
	}
	hasFields, err := containsConfigurationField(content)
	if err != nil {
		return nil, fmt.Errorf("inspect configuration file %q: %w", path, err)
	}
	if !hasFields {
		return nil, fmt.Errorf("configuration file %q does not contain supported configuration fields", path)
	}
	return &ConfigFile{
		Path:             path,
		Content:          string(content),
		Config:           config,
		ValidationErrors: validateConfiguration(config),
	}, nil
}

// containsConfigurationField reports whether a TOML document contains a field
// understood by the configuration editor, regardless of that field's value.
func containsConfigurationField(content []byte) (bool, error) {
	parser := &unstable.Parser{KeepComments: true}
	parser.Reset(content)
	section := ""
	for parser.NextExpression() {
		expression := parser.Expression()
		switch expression.Kind {
		case unstable.Table:
			section = nodePath(expression)
		case unstable.KeyValue:
			if isConfigurationField(section, nodeKey(expression)) {
				return true, nil
			}
		}
	}
	if err := parser.Error(); err != nil {
		return false, fmt.Errorf("invalid TOML: %w", err)
	}
	return false, nil
}

// isConfigurationField identifies keys belonging to the public GUI contract.
func isConfigurationField(section string, key string) bool {
	if section == "" {
		switch key {
		case "Log", "GeneratePlatformUserId", "Authentication":
			return true
		}
	}
	if section == "Games" && key == "Enabled" {
		return true
	}
	if section == "Announcement" {
		switch key {
		case "Enabled", "Multicast", "Port", "MulticastGroup":
			return true
		}
	}
	if strings.HasPrefix(section, "Games.") && key == "Hosts" {
		return true
	}
	return false
}

// parseConfiguration decodes TOML into the public configuration model.
func parseConfiguration(content []byte) (*Configuration, error) {
	var config Configuration
	if err := toml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid TOML or configuration type: %w", err)
	}
	if config.Games.Enabled == nil {
		config.Games.Enabled = []string{}
	}
	return &config, nil
}

// replaceDocumentValue replaces one value in a named TOML table.
func replaceDocumentValue(content []byte, targetSection string, key string, replacement string) ([]byte, error) {
	parser := &unstable.Parser{KeepComments: true}
	parser.Reset(content)
	section := ""
	for parser.NextExpression() {
		expression := parser.Expression()
		switch expression.Kind {
		case unstable.Table:
			section = nodePath(expression)
		case unstable.KeyValue:
			if section == targetSection && nodeKey(expression) == key {
				raw := expression.Raw
				rawContent := content[raw.Offset : raw.Offset+raw.Length]
				equals := bytes.IndexByte(rawContent, '=')
				if equals < 0 {
					return nil, fmt.Errorf("key %s.%s has no assignment", targetSection, key)
				}
				valueOffset := int(raw.Offset) + equals + 1
				for valueOffset < len(content) && (content[valueOffset] == ' ' || content[valueOffset] == '\t') {
					valueOffset++
				}
				valueEnd := int(raw.Offset + raw.Length)
				for valueEnd > valueOffset && (content[valueEnd-1] == ' ' || content[valueEnd-1] == '\t' || content[valueEnd-1] == '\r' || content[valueEnd-1] == '\n') {
					valueEnd--
				}
				replacementBytes := []byte(replacement)
				updated := make([]byte, 0, len(content)-(valueEnd-valueOffset)+len(replacementBytes))
				updated = append(updated, content[:valueOffset]...)
				updated = append(updated, replacementBytes...)
				updated = append(updated, content[valueEnd:]...)
				return updated, nil
			}
		}
	}
	if err := parser.Error(); err != nil {
		return nil, fmt.Errorf("invalid TOML: %w", err)
	}
	return nil, fmt.Errorf("Announcement.%s was not found", key)
}

// nodeKey returns the first key component from an AST node.
func nodeKey(node *unstable.Node) string {
	keys := node.Key()
	if !keys.Next() {
		return ""
	}
	return string(keys.Node().Data)
}

// nodePath returns all key components from an AST table node.
func nodePath(node *unstable.Node) string {
	keys := node.Key()
	var parts []string
	for keys.Next() {
		parts = append(parts, string(keys.Node().Data))
	}
	return strings.Join(parts, ".")
}

// writeConfigFile writes a new document through a temporary file replacement.
func writeConfigFile(path string, content []byte) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat configuration file: %w", err)
	}
	temporaryPath, err := createTemporaryConfig(filepath.Dir(path), content, fileInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer os.Remove(temporaryPath)

	return replaceConfigFile(path, temporaryPath)
}

// writeNewConfigFile creates a file without replacing an existing destination.
func writeNewConfigFile(path string, content []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create configuration directory: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("create configuration file: destination %q already exists", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check configuration destination: %w", err)
	}
	temporaryPath, err := createTemporaryConfig(directory, content, 0o644)
	if err != nil {
		return err
	}
	defer os.Remove(temporaryPath)
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("create configuration file: %w", err)
	}
	return nil
}

// createTemporaryConfig writes content to a temporary file with the given permissions.
func createTemporaryConfig(directory string, content []byte, permissions os.FileMode) (string, error) {
	temporary, err := os.CreateTemp(directory, ".config.toml-*")
	if err != nil {
		return "", fmt.Errorf("create temporary configuration file: %w", err)
	}
	temporaryPath := temporary.Name()

	if err := temporary.Chmod(permissions); err != nil {
		temporary.Close()
		os.Remove(temporaryPath)
		return "", fmt.Errorf("set temporary configuration permissions: %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		os.Remove(temporaryPath)
		return "", fmt.Errorf("write temporary configuration file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		os.Remove(temporaryPath)
		return "", fmt.Errorf("close temporary configuration file: %w", err)
	}
	return temporaryPath, nil
}

// replaceConfigFile swaps the original with the temporary file and rolls back on failure.
func replaceConfigFile(path string, temporaryPath string) error {
	backup, err := os.CreateTemp(filepath.Dir(path), ".config.toml-backup-*")
	if err != nil {
		return fmt.Errorf("create configuration backup: %w", err)
	}
	backupPath := backup.Name()
	if err := backup.Close(); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("close configuration backup: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("prepare configuration backup: %w", err)
	}
	defer os.Remove(backupPath)

	if err := os.Rename(path, backupPath); err != nil {
		return fmt.Errorf("create configuration backup: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if restoreErr := os.Rename(backupPath, path); restoreErr != nil {
			return fmt.Errorf("replace configuration file: %w; restore failed: %v", err, restoreErr)
		}
		return fmt.Errorf("replace configuration file: %w", err)
	}
	return nil
}

// validateConfiguration checks the semantic rules currently supported by the GUI.
func validateConfiguration(config *Configuration) []ValidationError {
	var validationErrors []ValidationError

	supportedAuthentication := map[string]bool{
		"required": true,
		"cached":   true,
		"adaptive": true,
		"disabled": true,
	}
	if !supportedAuthentication[config.Authentication] {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Authentication",
			Message: "must be one of required, cached, adaptive or disabled",
		})
	}
	if config.GeneratePlatformUserID && config.Authentication != "disabled" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "GeneratePlatformUserId",
			Message: "requires Authentication to be disabled",
		})
	}

	supportedGames := map[string]bool{
		"age1":   true,
		"age2":   true,
		"age3":   true,
		"age4":   true,
		"athens": true,
	}
	if len(config.Games.Enabled) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Games.Enabled",
			Message: "must contain at least one game",
		})
	}
	for index, game := range config.Games.Enabled {
		if !supportedGames[game] {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fmt.Sprintf("Games.Enabled.%d", index),
				Message: fmt.Sprintf("unsupported game %q", game),
			})
		}
	}
	validationErrors = append(validationErrors, validateGameHosts(config)...)

	if config.Announcement.Port < 1 || config.Announcement.Port > 65535 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Announcement.Port",
			Message: "must be an integer between 1 and 65535",
		})
	}
	if config.Announcement.Enabled && config.Announcement.Multicast {
		address, err := netip.ParseAddr(config.Announcement.MulticastGroup)
		if err != nil || !address.Is4() || !address.IsMulticast() {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "Announcement.MulticastGroup",
				Message: "must be an IPv4 multicast address",
			})
		}
	}

	return validationErrors
}

// validateGameHosts checks that enabled games have unique resolvable IPv4 hosts.
func validateGameHosts(config *Configuration) []ValidationError {
	var validationErrors []ValidationError
	usedHosts := make(map[string]string)
	for _, gameID := range config.Games.Enabled {
		hosts, supported := gameHosts(config, gameID)
		if !supported {
			continue
		}
		if len(hosts) == 0 {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fmt.Sprintf("Games.%s.Hosts", gameID),
				Message: "must contain at least one host",
			})
		}
		for index, host := range hosts {
			addresses := resolveIPv4(host)
			if len(addresses) == 0 {
				validationErrors = append(validationErrors, ValidationError{
					Field:   fmt.Sprintf("Games.%s.Hosts.%d", gameID, index),
					Message: "must resolve to an IPv4 address",
				})
				continue
			}
			for _, address := range addresses {
				if previousGame, exists := usedHosts[address]; exists && previousGame != gameID {
					validationErrors = append(validationErrors, ValidationError{
						Field:   fmt.Sprintf("Games.%s.Hosts.%d", gameID, index),
						Message: fmt.Sprintf("host is already used by %s", previousGame),
					})
				}
				usedHosts[address] = gameID
			}
		}
	}
	return validationErrors
}

// gameHosts returns the host list belonging to a supported game identifier.
func gameHosts(config *Configuration, gameID string) ([]string, bool) {
	switch gameID {
	case "age1":
		return config.Games.Age1.Hosts, true
	case "age2":
		return config.Games.Age2.Hosts, true
	case "age3":
		return config.Games.Age3.Hosts, true
	case "age4":
		return config.Games.Age4.Hosts, true
	case "athens":
		return config.Games.Athens.Hosts, true
	default:
		return nil, false
	}
}

// resolveIPv4 returns the IPv4 addresses resolved from a host value.
func resolveIPv4(host string) []string {
	if address, err := netip.ParseAddr(host); err == nil && address.Is4() {
		return []string{address.String()}
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil
	}
	var ipv4 []string
	for _, address := range addresses {
		if parsed, err := netip.ParseAddr(address.IP.String()); err == nil && parsed.Is4() {
			ipv4 = append(ipv4, parsed.String())
		}
	}
	return ipv4
}

// defaultConfigPath finds config.toml in the server's standard candidate paths.
func defaultConfigPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("find executable: %w", err)
	}

	candidates := []string{
		filepath.Join(filepath.Dir(executable), "resources", "config", configFileName),
		filepath.Join(filepath.Dir(executable), configFileName),
		filepath.Join("resources", "config", configFileName),
		configFileName,
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("check configuration file %q: %w", candidate, err)
		}
	}
	return "", os.ErrNotExist
}

// templateConfigPath finds the official template in installed and development layouts.
func templateConfigPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("find executable: %w", err)
	}
	executableDirectory := filepath.Dir(executable)
	candidates := []string{
		filepath.Join(executableDirectory, "resources", "config", configFileName),
		filepath.Join(executableDirectory, "..", "..", "..", "..", "server", "resources", "config", configFileName),
		filepath.Join("..", "..", "..", "server", "resources", "config", configFileName),
		filepath.Join("server", "resources", "config", configFileName),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("check configuration template %q: %w", candidate, err)
		}
	}
	return "", fmt.Errorf("configuration template %q was not found", configFileName)
}
