package web

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	toml "github.com/knadh/koanf/parsers/toml/v2"
)

// ConfigField describes one setting the admin panel may edit. Only fields in
// this list are writable: everything else in config.toml stays hand-edited, so
// a bug in the panel can never mangle an unknown key.
type ConfigField struct {
	Path    string   `json:"path"`
	Label   string   `json:"label"`
	Group   string   `json:"group"`
	Type    string   `json:"type"` // bool | int | string | strings | enum
	Options []string `json:"options,omitempty"`
	Help    string   `json:"help,omitempty"`
}

var httpLogLevels = []string{"none", "error", "warning", "notice", "info", "debug"}

// ConfigFields is the whitelist of editable settings.
func ConfigFields() []ConfigField {
	fields := []ConfigField{
		{Path: "Log", Label: "Write a log file", Group: "General", Type: "bool",
			Help: "Logs every request to a file in addition to the terminal. Enable it when reporting errors."},
		{Path: "Authentication", Label: "Authentication", Group: "General", Type: "enum",
			Options: []string{"disabled", "cached", "adaptive", "required"},
			Help:    "disabled keeps everything offline. Anything else may contact the official servers."},
		{Path: "GeneratePlatformUserId", Label: "Generate platform user ids", Group: "General", Type: "bool",
			Help: "Last resort for duplicate platform ids. Incompatible with Authentication other than disabled."},
		{Path: "Sign.Enabled", Label: "Re-sign catalogues", Group: "General", Type: "bool",
			Help: "Only for clients that trust this server's signing key."},
		{Path: "Internet.Enabled", Label: "Allow Internet", Group: "Internet", Type: "bool"},
		{Path: "Internet.IP", Label: "Public IPv4", Group: "Internet", Type: "string",
			Help: "Use 'auto' to detect it."},
		{Path: "Announcement.Enabled", Label: "LAN announcement", Group: "Announcement", Type: "bool"},
		{Path: "Announcement.Multicast", Label: "Multicast announcement", Group: "Announcement", Type: "bool"},
		{Path: "Announcement.MulticastGroup", Label: "Multicast group", Group: "Announcement", Type: "string"},
		{Path: "Announcement.Port", Label: "Announcement port", Group: "Announcement", Type: "int"},
		{Path: "Games.Enabled", Label: "Enabled games", Group: "Games", Type: "strings",
			Help: "age1, age2, age3, age4, athens."},
		{Path: "CustomFiles.Enabled", Label: "Custom files", Group: "Custom files", Type: "bool"},
		{Path: "CustomFiles.Folder", Label: "Folder", Group: "Custom files", Type: "string"},
		{Path: "CustomFiles.MaxFileSize", Label: "Max file size", Group: "Custom files", Type: "string"},
		{Path: "CustomFiles.MaxTotalSize", Label: "Max total size", Group: "Custom files", Type: "string"},
		{Path: "CustomFiles.Expiry", Label: "Expiry", Group: "Custom files", Type: "string"},
		{Path: "BattleServerReplays.Enabled", Label: "Record replays", Group: "Replays", Type: "bool"},
		{Path: "BattleServerReplays.Folder", Label: "Folder", Group: "Replays", Type: "string"},
		{Path: "BattleServerReplays.MaxFileSize", Label: "Max file size", Group: "Replays", Type: "string"},
		{Path: "BattleServerReplays.MaxTotalSize", Label: "Max total size", Group: "Replays", Type: "string"},
		{Path: "BattleServerReplays.Expiry", Label: "Expiry", Group: "Replays", Type: "string"},
		{Path: "Web.Enabled", Label: "Enable public dashboard", Group: "Dashboard", Type: "bool",
			Help: "The dashboard shares HTTPS port 443. Changing this setting requires a server restart."},
		{Path: "Web.PublicTitle", Label: "Site title", Group: "Dashboard", Type: "string",
			Help: "Changing this setting requires a server restart."},
		{Path: "Web.SnapshotSeconds", Label: "Live refresh interval (s)", Group: "Dashboard", Type: "int"},
	}
	for _, game := range []string{"age1", "age2", "age3", "age4", "athens"} {
		fields = append(fields,
			ConfigField{Path: "Games." + game + ".Hosts", Label: game + " hosts", Group: "Hosts", Type: "strings"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.StartIfMissing",
				Label: game + " embedded battle server", Group: "Embedded battle servers", Type: "enum",
				Options: []string{"true", "false", "required"}},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.Region", Label: game + " region",
				Group: "Embedded battle servers", Type: "string"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.Name", Label: game + " name",
				Group: "Embedded battle servers", Type: "string"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.IPv4", Label: game + " IPv4",
				Group: "Embedded battle servers", Type: "string"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.BsPort", Label: game + " BS port",
				Group: "Embedded battle servers", Type: "int"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.WebSocketPort", Label: game + " WS port",
				Group: "Embedded battle servers", Type: "int"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.OutOfBandPort", Label: game + " OOB port",
				Group: "Embedded battle servers", Type: "int"},
			ConfigField{Path: "Games." + game + ".EmbeddedBattleServer.LogLevel", Label: game + " log level",
				Group: "Embedded battle servers", Type: "enum", Options: httpLogLevels},
		)
	}
	return fields
}

// ParseConfig flattens a TOML document into dotted keys.
func ParseConfig(content []byte) (map[string]any, error) {
	parsed, err := toml.Parser().Unmarshal(content)
	if err != nil {
		return nil, err
	}
	flat := make(map[string]any)
	flatten("", parsed, flat)
	return flat, nil
}

func flatten(prefix string, value any, out map[string]any) {
	if prefix == "" {
		return
	}
	if nested, ok := value.(map[string]any); ok {
		for key, child := range nested {
			flatten(prefix+"."+key, child, out)
		}
		return
	}
	out[prefix] = value
}

// FormatTomlValue renders a JSON-decoded value as a TOML literal.
func FormatTomlValue(field ConfigField, value any) (string, error) {
	switch field.Type {
	case "bool":
		boolean, ok := value.(bool)
		if !ok {
			return "", fmt.Errorf("%s must be a boolean", field.Path)
		}
		if boolean {
			return "true", nil
		}
		return "false", nil
	case "int":
		var number int64
		switch typed := value.(type) {
		case float64:
			if typed != float64(int64(typed)) {
				return "", fmt.Errorf("%s must be an integer", field.Path)
			}
			number = int64(typed)
		case int:
			number = int64(typed)
		case int64:
			number = typed
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
			if err != nil {
				return "", fmt.Errorf("%s must be an integer", field.Path)
			}
			number = parsed
		default:
			return "", fmt.Errorf("%s must be an integer", field.Path)
		}
		return strconv.FormatInt(number, 10), nil
	case "string":
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%s must be a string", field.Path)
		}
		return quoteTomlString(text), nil
	case "enum":
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%s must be a string", field.Path)
		}
		for _, option := range field.Options {
			if text == option {
				return quoteTomlString(text), nil
			}
		}
		return "", fmt.Errorf("%s must be one of %s", field.Path, strings.Join(field.Options, ", "))
	case "strings":
		var items []string
		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				text, ok := item.(string)
				if !ok {
					return "", fmt.Errorf("%s must be a list of strings", field.Path)
				}
				items = append(items, text)
			}
		case []string:
			items = append(items, typed...)
		case string:
			for _, item := range strings.Split(typed, ",") {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					items = append(items, trimmed)
				}
			}
		default:
			return "", fmt.Errorf("%s must be a list of strings", field.Path)
		}
		quoted := make([]string, 0, len(items))
		for _, item := range items {
			quoted = append(quoted, quoteTomlString(item))
		}
		return "[" + strings.Join(quoted, ", ") + "]", nil
	default:
		return "", fmt.Errorf("unknown field type %q", field.Type)
	}
}

func quoteTomlString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "'", "\\'")
	return "'" + value + "'"
}

// ApplyConfigEdits rewrites the whitelisted keys of a TOML document in place,
// preserving every other line (comments included). Keys that do not exist yet
// are appended to their section.
func ApplyConfigEdits(content []byte, fields []ConfigField, literals map[string]string) ([]byte, error) {
	byPath := make(map[string]ConfigField, len(fields))
	for _, field := range fields {
		byPath[field.Path] = field
	}
	for path := range literals {
		if _, ok := byPath[path]; !ok {
			return nil, fmt.Errorf("%q is not editable from the dashboard", path)
		}
	}

	lines := strings.Split(string(content), "\n")
	section := ""
	applied := make(map[string]bool, len(literals))
	sectionEnd := make(map[string]int)
	sectionStart := make(map[string]int)

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(strings.TrimSuffix(trimmed, " "), "]") {
			name := strings.Trim(trimmed, "[]")
			section = strings.TrimSpace(name)
			sectionStart[section] = index
			sectionEnd[section] = len(lines) - 1
			continue
		}
		key, _, ok := splitTomlKeyValue(trimmed)
		if !ok {
			continue
		}
		fullPath := key
		if section != "" {
			fullPath = section + "." + key
		}
		literal, editable := literals[fullPath]
		if !editable {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		comment := trailingComment(line)
		newLine := indent + key + " = " + literal
		if comment != "" {
			newLine += " " + comment
		}
		lines[index] = newLine
		applied[fullPath] = true
	}

	// Discover where each section actually ends so new keys land inside their
	// own section and not at the bottom of the file.
	current := ""
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(strings.TrimSuffix(trimmed, " "), "]") {
			if current != "" {
				sectionEnd[current] = index
			}
			current = strings.TrimSpace(strings.Trim(trimmed, "[]"))
			sectionStart[current] = index
			sectionEnd[current] = len(lines) - 1
		}
	}
	if current != "" {
		sectionEnd[current] = len(lines) - 1
	}

	pending := make([]string, 0, len(literals))
	for path := range literals {
		if !applied[path] {
			pending = append(pending, path)
		}
	}
	sort.Strings(pending)

	// Insert from the bottom of the file up so earlier offsets stay valid.
	type insertion struct {
		index int
		text  string
	}
	var insertions []insertion
	for _, path := range pending {
		parent := parentSection(path)
		key := lastSegment(path)
		line := key + " = " + literals[path]
		if parent == "" {
			insertions = append(insertions, insertion{index: rootInsertIndex(lines), text: line})
			continue
		}
		if end, ok := sectionEnd[parent]; ok {
			insertions = append(insertions, insertion{index: end, text: line})
			continue
		}
		// The section does not exist: create it at the end.
		insertions = append(insertions, insertion{
			index: len(lines),
			text:  "[" + parent + "]\n" + line,
		})
	}
	sort.Slice(insertions, func(a, b int) bool { return insertions[a].index > insertions[b].index })
	for _, item := range insertions {
		if item.index >= len(lines) {
			lines = append(lines, strings.Split(item.text, "\n")...)
			continue
		}
		head := append([]string{}, lines[:item.index]...)
		tail := append([]string{}, lines[item.index:]...)
		lines = append(append(head, strings.Split(item.text, "\n")...), tail...)
	}

	result := []byte(strings.Join(lines, "\n"))
	if _, err := ParseConfig(result); err != nil {
		return nil, fmt.Errorf("the resulting file is not valid TOML: %w", err)
	}
	return result, nil
}

// rootInsertIndex returns where a root-level key should go: before the first
// section header, after the last root-level key.
func rootInsertIndex(lines []string) int {
	index := 0
	for position, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			break
		}
		if _, _, ok := splitTomlKeyValue(trimmed); ok {
			index = position + 1
		}
	}
	return index
}

func splitTomlKeyValue(line string) (key, value string, ok bool) {
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
		return "", "", false
	}
	position := strings.Index(line, "=")
	if position <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:position])
	if key == "" || strings.ContainsAny(key, " \t\"'#") {
		return "", "", false
	}
	return key, strings.TrimSpace(line[position+1:]), true
}

// trailingComment extracts a trailing "# comment" that is not inside a string.
func trailingComment(line string) string {
	inSingle, inDouble, escaped := false, false, false
	for index, char := range line {
		switch {
		case escaped:
			escaped = false
		case char == '\\' && inDouble:
			escaped = true
		case char == '\'' && !inDouble:
			inSingle = !inSingle
		case char == '"' && !inSingle:
			inDouble = !inDouble
		case char == '#' && !inSingle && !inDouble:
			return strings.TrimSpace(line[index:])
		}
	}
	return ""
}

func parentSection(path string) string {
	position := strings.LastIndex(path, ".")
	if position < 0 {
		return ""
	}
	return path[:position]
}

func lastSegment(path string) string {
	position := strings.LastIndex(path, ".")
	if position < 0 {
		return path
	}
	return path[position+1:]
}

// SaveConfigFile writes the new config, keeping a timestamped backup of the
// previous content. The backup is what makes the raw editor safe.
func SaveConfigFile(path string, content []byte) (string, error) {
	if path == "" {
		return "", errors.New("no config file is in use")
	}
	previous, err := os.ReadFile(path)
	if err == nil {
		backup := fmt.Sprintf("%s.bak-%s", path, time.Now().Format("20060102-150405"))
		if writeErr := os.WriteFile(backup, previous, 0o644); writeErr != nil {
			return "", fmt.Errorf("write backup: %w", writeErr)
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err = os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
