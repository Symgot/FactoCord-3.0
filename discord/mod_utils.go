package discord

import (
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

// StringPtr gibt einen Pointer auf einen String zurück
func StringPtr(s string) *string {
	return &s
}

// BoolPtr gibt einen Pointer auf einen Bool zurück
func BoolPtr(b bool) *bool {
	return &b
}

// IntPtr gibt einen Pointer auf einen Int zurück
func IntPtr(i int) *int {
	return &i
}

// Min gibt das Minimum zweier Ints zurück
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max gibt das Maximum zweier Ints zurück
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp begrenzt einen Wert auf einen Bereich
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ChunkSlice teilt eine Slice in Chunks einer bestimmten Größe
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// FormatModName formatiert einen Mod-Namen für die Anzeige
func FormatModName(name string) string {
	if len(name) > 25 {
		return name[:22] + "..."
	}
	return name
}

// GetModInfo konvertiert models.ModInfo zu factorio.ModInfo
func ConvertModInfo(mod *models.ModInfo) *ModInfo {
	return &ModInfo{
		Name:            mod.Name,
		Version:         mod.Version,
		Enabled:         mod.Enabled,
		Title:           mod.Title,
		Author:          mod.Author,
		Description:     mod.Description,
		StartupSettings: mod.StartupSettings,
		RuntimeSettings: mod.RuntimeSettings,
		GameSettings:    mod.GameSettings,
		MapSettings:     mod.MapSettings,
	}
}

// ModInfo ist eine lokale Kopie für das Discord-Package
type ModInfo struct {
	Name            string
	Version         string
	Enabled         bool
	Title           string
	Author          string
	Description     string
	StartupSettings map[string]interface{}
	RuntimeSettings map[string]interface{}
	GameSettings    map[string]interface{}
	MapSettings     map[string]interface{}
}

// CountSettings zählt Game- und Map-Settings
func (m *ModInfo) CountSettings() (gameCount, mapCount int) {
	gameCount = len(m.StartupSettings) + len(m.RuntimeSettings) + len(m.GameSettings)
	mapCount = len(m.MapSettings)
	return
}

// GetAllGameSettings gibt alle Game-Settings zurück
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

// GetAllMapSettings gibt alle Map-Settings zurück
func (m *ModInfo) GetAllMapSettings() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m.MapSettings {
		result[k] = v
	}

	return result
}
