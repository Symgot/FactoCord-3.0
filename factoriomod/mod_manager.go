package factoriomod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

// ModManager verwaltet alle installierten Factorio-Mods
type ModManager struct {
	FactorioPath string
	SavePath     string
	Mods         []models.ModInfo
	mu           sync.RWMutex
}

// GlobalModManager ist die globale Instanz des ModManagers
var GlobalModManager *ModManager

// InitModManager initialisiert den globalen ModManager
func InitModManager(factorioPath, savePath string) {
	GlobalModManager = &ModManager{
		FactorioPath: factorioPath,
		SavePath:     savePath,
		Mods:         make([]models.ModInfo, 0),
	}
}

// DiscoverMods scannt das mod-settings.dat und lädt alle Mods
func (mm *ModManager) DiscoverMods() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Versuche mod-settings.dat zu lesen
	modSettingsPath := filepath.Join(mm.FactorioPath, "mods", "mod-settings.dat")

	// Prüfe ob die Datei existiert
	if _, err := os.Stat(modSettingsPath); os.IsNotExist(err) {
		// Versuche alternativen Pfad
		modSettingsPath = filepath.Join(mm.FactorioPath, "player-data", "mod-settings.dat")
		if _, err := os.Stat(modSettingsPath); os.IsNotExist(err) {
			return fmt.Errorf("mod-settings.dat nicht gefunden")
		}
	}

	settings, err := ParseModSettings(modSettingsPath)
	if err != nil {
		return fmt.Errorf("fehler beim Parsen der mod-settings: %w", err)
	}

	// Mod-Liste laden
	modListPath := filepath.Join(mm.FactorioPath, "mods", "mod-list.json")
	modList, err := mm.loadModList(modListPath)
	if err != nil {
		// Nicht fatal, wir können auch ohne mod-list.json arbeiten
		modList = make(map[string]bool)
	}

	// Mods zurücksetzen
	mm.Mods = make([]models.ModInfo, 0)

	// Verarbeite Startup-Settings
	if startup, ok := settings["startup"].(map[string]interface{}); ok {
		for modName, modSettings := range startup {
			mod := mm.findOrCreateMod(modName)
			if settingsMap, ok := modSettings.(map[string]interface{}); ok {
				mod.StartupSettings = extractSettings(settingsMap)
			}
		}
	}

	// Verarbeite Runtime-Global-Settings
	if runtimeGlobal, ok := settings["runtime-global"].(map[string]interface{}); ok {
		for modName, modSettings := range runtimeGlobal {
			mod := mm.findOrCreateMod(modName)
			if settingsMap, ok := modSettings.(map[string]interface{}); ok {
				mod.RuntimeSettings = extractSettings(settingsMap)
			}
		}
	}

	// Verarbeite Runtime-Per-User-Settings (als Map-Settings behandeln)
	if runtimePerUser, ok := settings["runtime-per-user"].(map[string]interface{}); ok {
		for modName, modSettings := range runtimePerUser {
			mod := mm.findOrCreateMod(modName)
			if settingsMap, ok := modSettings.(map[string]interface{}); ok {
				mod.MapSettings = extractSettings(settingsMap)
			}
		}
	}

	// Aktualisiere Enabled-Status aus mod-list.json
	for i := range mm.Mods {
		if enabled, exists := modList[mm.Mods[i].Name]; exists {
			mm.Mods[i].Enabled = enabled
		} else {
			mm.Mods[i].Enabled = true // Standard: aktiviert
		}
	}

	// Sortiere Mods alphabetisch
	sort.Slice(mm.Mods, func(i, j int) bool {
		return mm.Mods[i].Name < mm.Mods[j].Name
	})

	return nil
}

// findOrCreateMod findet einen Mod oder erstellt einen neuen
func (mm *ModManager) findOrCreateMod(name string) *models.ModInfo {
	for i := range mm.Mods {
		if mm.Mods[i].Name == name {
			return &mm.Mods[i]
		}
	}

	// Neuen Mod erstellen
	newMod := models.ModInfo{
		Name:            name,
		Enabled:         true,
		StartupSettings: make(map[string]interface{}),
		RuntimeSettings: make(map[string]interface{}),
		GameSettings:    make(map[string]interface{}),
		MapSettings:     make(map[string]interface{}),
	}
	mm.Mods = append(mm.Mods, newMod)
	return &mm.Mods[len(mm.Mods)-1]
}

// GetModByName gibt einen Mod nach Namen zurück
func (mm *ModManager) GetModByName(name string) *models.ModInfo {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	for i := range mm.Mods {
		if mm.Mods[i].Name == name {
			return &mm.Mods[i]
		}
	}
	return nil
}

// GetAllMods gibt alle geladenen Mods zurück
func (mm *ModManager) GetAllMods() []models.ModInfo {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make([]models.ModInfo, len(mm.Mods))
	copy(result, mm.Mods)
	return result
}

// GetModsWithSettings gibt nur Mods mit konfigurierbaren Settings zurück
func (mm *ModManager) GetModsWithSettings() []models.ModInfo {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make([]models.ModInfo, 0)
	for _, mod := range mm.Mods {
		if mod.HasSettings() {
			result = append(result, mod)
		}
	}
	return result
}

// loadModList lädt die mod-list.json Datei
func (mm *ModManager) loadModList(path string) (map[string]bool, error) {
	result := make(map[string]bool)

	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}

	var modList struct {
		Mods []struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"mods"`
	}

	if err := json.Unmarshal(data, &modList); err != nil {
		return result, err
	}

	for _, mod := range modList.Mods {
		result[mod.Name] = mod.Enabled
	}

	return result, nil
}

// extractSettings extrahiert die Werte aus einem Settings-Map
func extractSettings(settings map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range settings {
		if valueMap, ok := value.(map[string]interface{}); ok {
			// Factorio speichert Settings als {"value": actualValue}
			if actualValue, exists := valueMap["value"]; exists {
				result[key] = actualValue
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// GetModCount gibt die Anzahl der geladenen Mods zurück
func (mm *ModManager) GetModCount() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.Mods)
}
