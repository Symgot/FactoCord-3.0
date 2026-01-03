package models

// ModInfo speichert Metadaten eines Factorio-Mods
type ModInfo struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Enabled         bool                   `json:"enabled"`
	Title           string                 `json:"title,omitempty"`
	Author          string                 `json:"author,omitempty"`
	Description     string                 `json:"description,omitempty"`
	StartupSettings map[string]interface{} `json:"startup_settings,omitempty"`
	RuntimeSettings map[string]interface{} `json:"runtime_settings,omitempty"`
	GameSettings    map[string]interface{} `json:"game_settings,omitempty"`
	MapSettings     map[string]interface{} `json:"map_settings,omitempty"`
}

// SettingValue represents a typed setting value with metadata
type SettingValue struct {
	Value         interface{} `json:"value"`
	Type          string      `json:"type"`         // "bool", "int", "double", "string", "color"
	DefaultType   string      `json:"default_type"` // Original Factorio type
	Order         string      `json:"order,omitempty"`
	LocalisedName string      `json:"localised_name,omitempty"`
}

// ModSettings represents the complete mod-settings structure
type ModSettings struct {
	Startup        map[string]map[string]SettingValue `json:"startup"`
	RuntimeGlobal  map[string]map[string]SettingValue `json:"runtime-global"`
	RuntimePerUser map[string]map[string]SettingValue `json:"runtime-per-user"`
}

// GetAllGameSettings returns all game-related settings (startup + runtime-global)
func (m *ModInfo) GetAllGameSettings() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m.StartupSettings {
		result[k] = v
	}
	for k, v := range m.RuntimeSettings {
		result[k] = v
	}
	for k, v := range m.GameSettings {
		result[k] = v
	}

	return result
}

// GetAllMapSettings returns all map-related settings
func (m *ModInfo) GetAllMapSettings() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m.MapSettings {
		result[k] = v
	}

	return result
}

// HasSettings returns true if the mod has any configurable settings
func (m *ModInfo) HasSettings() bool {
	return len(m.StartupSettings) > 0 ||
		len(m.RuntimeSettings) > 0 ||
		len(m.GameSettings) > 0 ||
		len(m.MapSettings) > 0
}

// CountSettings returns the total number of settings across all categories
func (m *ModInfo) CountSettings() (gameCount, mapCount int) {
	gameCount = len(m.StartupSettings) + len(m.RuntimeSettings) + len(m.GameSettings)
	mapCount = len(m.MapSettings)
	return
}

// GetGameSetting returns a specific game setting by key
func (m *ModInfo) GetGameSetting(key string) interface{} {
	if val, ok := m.StartupSettings[key]; ok {
		return val
	}
	if val, ok := m.RuntimeSettings[key]; ok {
		return val
	}
	if val, ok := m.GameSettings[key]; ok {
		return val
	}
	return nil
}

// GetMapSetting returns a specific map setting by key
func (m *ModInfo) GetMapSetting(key string) interface{} {
	if val, ok := m.MapSettings[key]; ok {
		return val
	}
	return nil
}
