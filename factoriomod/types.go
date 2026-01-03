package factoriomod

// SettingType definiert die möglichen Setting-Typen in Factorio
type SettingType string

const (
	SettingTypeBool         SettingType = "bool-setting"
	SettingTypeInt          SettingType = "int-setting"
	SettingTypeDouble       SettingType = "double-setting"
	SettingTypeString       SettingType = "string-setting"
	SettingTypeColorSetting SettingType = "color-setting"
)

// SettingScope definiert den Gültigkeitsbereich eines Settings
type SettingScope string

const (
	SettingScopeStartup       SettingScope = "startup"
	SettingScopeRuntimeGlobal SettingScope = "runtime-global"
	SettingScopeRuntimePlayer SettingScope = "runtime-per-user"
)

// SettingDefinition beschreibt ein Mod-Setting
type SettingDefinition struct {
	Name          string        `json:"name"`
	Type          SettingType   `json:"type"`
	Scope         SettingScope  `json:"setting_type"`
	DefaultValue  interface{}   `json:"default_value"`
	AllowedValues []interface{} `json:"allowed_values,omitempty"`
	MinimumValue  *float64      `json:"minimum_value,omitempty"`
	MaximumValue  *float64      `json:"maximum_value,omitempty"`
	Order         string        `json:"order,omitempty"`
	Hidden        bool          `json:"hidden,omitempty"`
	LocalisedName string        `json:"localised_name,omitempty"`
	LocalisedDesc string        `json:"localised_description,omitempty"`
}

// ChangeRecord speichert eine Änderung für Audit-Zwecke
type ChangeRecord struct {
	UserID      string                 `json:"user_id"`
	ModName     string                 `json:"mod_name"`
	Timestamp   int64                  `json:"timestamp"`
	Changes     map[string]interface{} `json:"changes"`
	OldValues   map[string]interface{} `json:"old_values"`
	WasApplied  bool                   `json:"was_applied"`
	AppliedAt   int64                  `json:"applied_at,omitempty"`
	ServerState string                 `json:"server_state,omitempty"`
}

// BackupInfo enthält Informationen über ein Backup
type BackupInfo struct {
	Filename    string `json:"filename"`
	Timestamp   int64  `json:"timestamp"`
	Size        int64  `json:"size"`
	Reason      string `json:"reason,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	IsAutomatic bool   `json:"is_automatic"`
}

// ModState repräsentiert den Zustand eines Mods
type ModState struct {
	Enabled bool   `json:"enabled"`
	Version string `json:"version,omitempty"`
}

// ModListEntry ist ein Eintrag in der mod-list.json
type ModListEntry struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// ModList repräsentiert die mod-list.json Struktur
type ModList struct {
	Mods []ModListEntry `json:"mods"`
}

// ServerStatus repräsentiert den aktuellen Server-Status
type ServerStatus struct {
	IsRunning   bool   `json:"is_running"`
	IsPaused    bool   `json:"is_paused"`
	PlayerCount int    `json:"player_count"`
	MapName     string `json:"map_name,omitempty"`
	Uptime      int64  `json:"uptime,omitempty"`
}

// SettingChange repräsentiert eine einzelne Setting-Änderung
type SettingChange struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Scope    string      `json:"scope"`
}

// ApplyResult ist das Ergebnis einer Änderungsanwendung
type ApplyResult struct {
	Success      bool            `json:"success"`
	BackupFile   string          `json:"backup_file,omitempty"`
	AppliedCount int             `json:"applied_count"`
	FailedCount  int             `json:"failed_count"`
	Errors       []string        `json:"errors,omitempty"`
	Changes      []SettingChange `json:"changes,omitempty"`
}
