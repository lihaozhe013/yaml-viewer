// Package config persists the small amount of application state that should
// survive between viewer sessions.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const (
	configDirectory = ".config/yaml-viewer"
	configFileName  = "config.yaml"
)

//go:embed default.yaml
var defaultConfigData []byte

// SearchMode identifies how search queries are matched.
type SearchMode string

const (
	SearchModeSmartFuzzy SearchMode = "smart_fuzzy"
	SearchModeKeyword    SearchMode = "keyword"
)

// ThemeMode identifies the application's color scheme.
type ThemeMode string

const (
	ThemeModeLight ThemeMode = "light"
	ThemeModeDark  ThemeMode = "dark"
)

// Config is intentionally small and extensible. Unknown fields from the
// user's config are kept in raw so newer versions can add defaults without
// deleting user-owned extensions.
type Config struct {
	LastOpenedFile string     `yaml:"last_opened_file"`
	SearchMode     SearchMode `yaml:"search_mode"`
	ThemeMode      ThemeMode  `yaml:"theme"`
	Indent         int        `yaml:"indent"`
	SortKeys       bool       `yaml:"sort_keys"`

	raw           map[string]any
	invalidSource bool
}

// NormalizeSearchMode returns a supported mode and falls back to the default
// when a config value is missing or invalid.
func NormalizeSearchMode(value SearchMode) SearchMode {
	switch value {
	case SearchModeKeyword:
		return SearchModeKeyword
	case SearchModeSmartFuzzy:
		return SearchModeSmartFuzzy
	default:
		return SearchModeSmartFuzzy
	}
}

// NormalizeThemeMode returns a supported theme and falls back to the light
// theme when a config value is missing or invalid.
func NormalizeThemeMode(value ThemeMode) ThemeMode {
	switch value {
	case ThemeModeDark:
		return ThemeModeDark
	case ThemeModeLight:
		return ThemeModeLight
	default:
		return ThemeModeLight
	}
}

// NormalizeIndent returns a supported indent width and falls back to 2 spaces
// when a config value is missing or invalid.
func NormalizeIndent(value int) int {
	switch value {
	case 2, 4:
		return value
	default:
		return 2
	}
}

// Path returns the platform user's requested configuration path:
// ~/.config/yaml-viewer/config.yaml.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, configDirectory, configFileName), nil
}

// Load reads the configuration, merges it with the embedded defaults, and
// writes any newly added or repaired fields back to disk.
func Load() (Config, error) {
	defaults, err := decodeMap(defaultConfigData)
	if err != nil {
		return Config{}, fmt.Errorf("parse embedded config template: %w", err)
	}
	defaults, _ = normalizeKnownFields(defaults, nil)

	path, err := Path()
	if err != nil {
		return configFromMap(defaults), err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		value := configFromMap(defaults)
		if err := savePath(path, value); err != nil {
			return value, err
		}
		return value, nil
	}
	if err != nil {
		return configFromMap(defaults), fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		value := configFromMap(defaults)
		if err := savePath(path, value); err != nil {
			return value, err
		}
		return value, nil
	}

	userConfig, err := decodeMap(data)
	if err != nil {
		value := configFromMap(defaults)
		value.invalidSource = true
		return value, fmt.Errorf("parse %s: %w", path, err)
	}
	merged, changed := mergeDefaults(defaults, userConfig)
	merged, repaired := normalizeKnownFields(merged, defaults)
	if repaired {
		changed = true
	}
	value := configFromMap(merged)
	if changed {
		if err := savePath(path, value); err != nil {
			return value, err
		}
	}
	return value, nil
}

// Save writes the configuration atomically with user-only file permissions.
// The temporary file lives beside the final file so rename remains atomic on
// the same filesystem.
func Save(value Config) error {
	if value.invalidSource {
		return errors.New("refusing to overwrite an unparseable config file")
	}
	path, err := Path()
	if err != nil {
		return err
	}
	return savePath(path, value)
}

func savePath(path string, value Config) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory %s: %w", directory, err)
	}
	data, err := yaml.Marshal(&value)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".config.yaml-*")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary config: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

func (value Config) MarshalYAML() (any, error) {
	fields := cloneMap(value.raw)
	if fields == nil {
		fields = make(map[string]any)
	}
	fields["last_opened_file"] = value.LastOpenedFile
	fields["search_mode"] = NormalizeSearchMode(value.SearchMode)
	fields["theme"] = NormalizeThemeMode(value.ThemeMode)
	fields["indent"] = NormalizeIndent(value.Indent)
	fields["sort_keys"] = value.SortKeys
	return fields, nil
}

func configFromMap(fields map[string]any) Config {
	value := Config{raw: cloneMap(fields)}
	if raw, ok := fields["last_opened_file"].(string); ok {
		value.LastOpenedFile = raw
	}
	if raw, ok := fields["search_mode"].(string); ok {
		value.SearchMode = NormalizeSearchMode(SearchMode(raw))
	}
	if value.SearchMode == "" {
		value.SearchMode = SearchModeSmartFuzzy
	}
	if raw, ok := fields["theme"].(string); ok {
		value.ThemeMode = NormalizeThemeMode(ThemeMode(raw))
	}
	if value.ThemeMode == "" {
		value.ThemeMode = ThemeModeLight
	}
	if raw, ok := fields["indent"]; ok {
		switch v := raw.(type) {
		case int:
			value.Indent = NormalizeIndent(v)
		case float64:
			value.Indent = NormalizeIndent(int(v))
		}
	}
	if value.Indent == 0 {
		value.Indent = 2
	}
	if raw, ok := fields["sort_keys"].(bool); ok {
		value.SortKeys = raw
	} else {
		value.SortKeys = true
	}
	return value
}

func decodeMap(data []byte) (map[string]any, error) {
	var fields map[string]any
	if err := yaml.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields == nil {
		fields = make(map[string]any)
	}
	return normalizeMapValues(fields).(map[string]any), nil
}

func normalizeKnownFields(fields, fallback map[string]any) (map[string]any, bool) {
	changed := false
	defaultPath := ""
	if raw, ok := fallback["last_opened_file"].(string); ok {
		defaultPath = raw
	}
	defaultMode := string(SearchModeSmartFuzzy)
	if raw, ok := fallback["search_mode"].(string); ok && NormalizeSearchMode(SearchMode(raw)) == SearchMode(raw) {
		defaultMode = raw
	}
	defaultTheme := string(ThemeModeLight)
	if raw, ok := fallback["theme"].(string); ok && NormalizeThemeMode(ThemeMode(raw)) == ThemeMode(raw) {
		defaultTheme = raw
	}
	defaultIndent := 2
	if raw, ok := fallback["indent"]; ok {
		switch v := raw.(type) {
		case int:
			if NormalizeIndent(v) == v {
				defaultIndent = v
			}
		case float64:
			iv := int(v)
			if NormalizeIndent(iv) == iv {
				defaultIndent = iv
			}
		}
	}
	defaultSortKeys := true
	if raw, ok := fallback["sort_keys"].(bool); ok {
		defaultSortKeys = raw
	}
	if raw, ok := fields["last_opened_file"]; !ok {
		fields["last_opened_file"] = defaultPath
		changed = true
	} else if _, ok := raw.(string); !ok {
		fields["last_opened_file"] = defaultPath
		changed = true
	}

	if raw, ok := fields["search_mode"]; !ok {
		fields["search_mode"] = defaultMode
		changed = true
	} else if mode, ok := raw.(string); !ok || NormalizeSearchMode(SearchMode(mode)) != SearchMode(mode) {
		fields["search_mode"] = defaultMode
		changed = true
	}

	if raw, ok := fields["theme"]; !ok {
		fields["theme"] = defaultTheme
		changed = true
	} else if mode, ok := raw.(string); !ok || NormalizeThemeMode(ThemeMode(mode)) != ThemeMode(mode) {
		fields["theme"] = defaultTheme
		changed = true
	}

	if raw, ok := fields["indent"]; !ok {
		fields["indent"] = defaultIndent
		changed = true
	} else {
		var indent int
		switch v := raw.(type) {
		case int:
			indent = v
		case float64:
			indent = int(v)
		default:
			fields["indent"] = defaultIndent
			changed = true
		}
		if NormalizeIndent(indent) != indent {
			fields["indent"] = defaultIndent
			changed = true
		}
	}

	if raw, ok := fields["sort_keys"]; !ok {
		fields["sort_keys"] = defaultSortKeys
		changed = true
	} else if _, ok := raw.(bool); !ok {
		fields["sort_keys"] = defaultSortKeys
		changed = true
	}

	return fields, changed
}

func mergeDefaults(defaults, user map[string]any) (map[string]any, bool) {
	merged := cloneMap(user)
	changed := false
	for key, defaultValue := range defaults {
		userValue, exists := user[key]
		if !exists {
			merged[key] = cloneValue(defaultValue)
			changed = true
			continue
		}
		defaultMap, defaultOK := asMap(defaultValue)
		userMap, userOK := asMap(userValue)
		if defaultOK && userOK {
			child, childChanged := mergeDefaults(defaultMap, userMap)
			merged[key] = child
			changed = changed || childChanged
		}
	}
	return merged, changed
}

func asMap(value any) (map[string]any, bool) {
	fields, ok := value.(map[string]any)
	return fields, ok
}

func normalizeMapValues(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			result[key] = normalizeMapValues(child)
		}
		return result
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			result[fmt.Sprint(key)] = normalizeMapValues(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = normalizeMapValues(child)
		}
		return result
	default:
		return value
	}
}

func cloneMap(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	result := make(map[string]any, len(fields))
	for key, value := range fields {
		result[key] = cloneValue(value)
	}
	return result
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = cloneValue(child)
		}
		return result
	default:
		return value
	}
}
