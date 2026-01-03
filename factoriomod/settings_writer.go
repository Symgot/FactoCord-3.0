package factoriomod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SettingsWriter schreibt Änderungen in die mod-settings Datei
type SettingsWriter struct {
	FactorioPath string
	BackupPath   string
}

// NewSettingsWriter erstellt einen neuen SettingsWriter
func NewSettingsWriter(factorioPath, backupPath string) *SettingsWriter {
	return &SettingsWriter{
		FactorioPath: factorioPath,
		BackupPath:   backupPath,
	}
}

// ApplyChanges wendet Änderungen auf die mod-settings an
func (sw *SettingsWriter) ApplyChanges(modName string, changes map[string]interface{}) error {
	modSettingsPath := sw.getModSettingsPath()

	// Backup erstellen
	if err := sw.createBackup(modSettingsPath); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Backups: %w", err)
	}

	// Aktuelle Settings laden
	settings, err := ParseModSettings(modSettingsPath)
	if err != nil {
		return fmt.Errorf("fehler beim Laden der Settings: %w", err)
	}

	// Änderungen anwenden
	if err := sw.applyChangesToSettings(settings, modName, changes); err != nil {
		return fmt.Errorf("fehler beim Anwenden der Änderungen: %w", err)
	}

	// Settings speichern
	if err := sw.writeSettings(modSettingsPath, settings); err != nil {
		return fmt.Errorf("fehler beim Speichern der Settings: %w", err)
	}

	return nil
}

// getModSettingsPath gibt den Pfad zur mod-settings Datei zurück
func (sw *SettingsWriter) getModSettingsPath() string {
	// Versuche verschiedene Pfade
	paths := []string{
		filepath.Join(sw.FactorioPath, "mods", "mod-settings.dat"),
		filepath.Join(sw.FactorioPath, "player-data", "mod-settings.dat"),
		filepath.Join(sw.FactorioPath, "mods", "mod-settings.json"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Standard-Pfad zurückgeben
	return paths[0]
}

// createBackup erstellt ein Backup der mod-settings Datei
func (sw *SettingsWriter) createBackup(sourcePath string) error {
	// Backup-Verzeichnis erstellen falls nicht vorhanden
	if sw.BackupPath == "" {
		sw.BackupPath = filepath.Join(sw.FactorioPath, "backups")
	}

	if err := os.MkdirAll(sw.BackupPath, 0755); err != nil {
		return err
	}

	// Backup-Dateiname mit Timestamp
	timestamp := time.Now().Unix()
	backupName := fmt.Sprintf("mod-settings_%d.dat.backup", timestamp)
	backupPath := filepath.Join(sw.BackupPath, backupName)

	// Datei kopieren
	return copyFile(sourcePath, backupPath)
}

// applyChangesToSettings wendet die Änderungen auf die Settings an
func (sw *SettingsWriter) applyChangesToSettings(settings map[string]interface{}, modName string, changes map[string]interface{}) error {
	// Suche nach dem Mod in den verschiedenen Setting-Kategorien
	categories := []string{"startup", "runtime-global", "runtime-per-user"}

	for _, category := range categories {
		if categoryMap, ok := settings[category].(map[string]interface{}); ok {
			if modSettings, ok := categoryMap[modName].(map[string]interface{}); ok {
				// Wende Änderungen auf diesen Mod an
				for key, value := range changes {
					if existingSetting, exists := modSettings[key]; exists {
						// Aktualisiere existierende Setting
						if settingMap, ok := existingSetting.(map[string]interface{}); ok {
							settingMap["value"] = value
						} else {
							modSettings[key] = map[string]interface{}{"value": value}
						}
					} else {
						// Neue Setting hinzufügen
						modSettings[key] = map[string]interface{}{"value": value}
					}
				}
			}
		}
	}

	return nil
}

// writeSettings schreibt die Settings zurück in die Datei
func (sw *SettingsWriter) writeSettings(path string, settings map[string]interface{}) error {
	// Wenn es eine JSON-Datei ist, schreibe JSON
	if filepath.Ext(path) == ".json" {
		return sw.writeSettingsJSON(path, settings)
	}

	// Für .dat Dateien versuchen wir JSON im Binärformat zu schreiben
	// Hinweis: Das echte Factorio Binärformat ist komplexer
	// Hier verwenden wir eine JSON-Zwischenlösung
	return sw.writeSettingsJSON(path+".json", settings)
}

// writeSettingsJSON schreibt Settings als JSON
func (sw *SettingsWriter) writeSettingsJSON(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// copyFile kopiert eine Datei
func copyFile(src, dst string) error {
	source, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, source, 0644)
}

// RestoreBackup stellt ein Backup wieder her
func (sw *SettingsWriter) RestoreBackup(backupFile string) error {
	modSettingsPath := sw.getModSettingsPath()
	backupPath := filepath.Join(sw.BackupPath, backupFile)

	return copyFile(backupPath, modSettingsPath)
}

// ListBackups gibt eine Liste aller Backups zurück
func (sw *SettingsWriter) ListBackups() ([]string, error) {
	if sw.BackupPath == "" {
		sw.BackupPath = filepath.Join(sw.FactorioPath, "backups")
	}

	entries, err := os.ReadDir(sw.BackupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	backups := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}
