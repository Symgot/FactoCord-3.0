package support

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveFactorioPath versucht den Factorio-Basis/Datapfad zu ermitteln.
//
// Wichtig: Für Mod-Settings/Mod-List braucht FactoCord den Pfad, der ein "mods" Verzeichnis enthält
// (typischerweise der Factorio-User-Data-Ordner, der mod-list.json enthält).
//
// Reihenfolge:
//  1. Config.ModListLocation -> ..\mods\mod-list.json => Parent(mods)
//  2. Config.Executable -> versuche ..\bin\x64\factorio(.exe) => Parent(Parent(Parent(executable)))
//  3. Fallback: Standardpfade
func ResolveFactorioPath() string {
	if Config.ModListLocation != "" {
		modsDir := filepath.Dir(Config.ModListLocation)
		baseDir := filepath.Dir(modsDir)
		if absPath, ok := absIfExists(baseDir); ok {
			fmt.Printf("Factorio-Pfad aus ModListLocation ermittelt: %s\n", absPath)
			return absPath
		}
	}

	execPath := strings.TrimSpace(Config.Executable)
	if execPath != "" {
		// Heuristik: <base>/bin/x64/factorio(.exe)
		candidate := filepath.Dir(execPath) // bin/x64
		candidate = filepath.Dir(candidate) // bin
		candidate = filepath.Dir(candidate) // base

		if absPath, ok := absIfExists(candidate); ok {
			fmt.Printf("Factorio-Pfad aus Executable ermittelt: %s\n", absPath)
			return absPath
		}

		// Alternativer Fallback: Verzeichnis der Executable selbst
		if absPath, ok := absIfExists(filepath.Dir(execPath)); ok {
			fmt.Printf("Factorio-Pfad aus Executable-Dir ermittelt: %s\n", absPath)
			return absPath
		}
	}

	paths := []string{
		"./factorio",
		"..",
		"/opt/factorio",
		"C:\\Program Files\\Factorio",
	}

	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		modsPath := filepath.Join(absPath, "mods")
		if _, err := os.Stat(modsPath); err == nil {
			fmt.Printf("Factorio-Pfad gefunden: %s\n", absPath)
			return absPath
		}
	}

	return "."
}

func absIfExists(p string) (string, bool) {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", false
	}
	if _, statErr := os.Stat(absPath); statErr != nil {
		return "", false
	}
	return absPath, true
}
