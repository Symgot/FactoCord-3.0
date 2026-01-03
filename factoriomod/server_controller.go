package factoriomod

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// ServerController verwaltet den Factorio-Server
type ServerController struct {
	FactorioPath string
	SavePath     string
}

// NewServerController erstellt einen neuen ServerController
func NewServerController(factorioPath, savePath string) *ServerController {
	return &ServerController{
		FactorioPath: factorioPath,
		SavePath:     savePath,
	}
}

// GlobalServerController ist die globale Instanz
var GlobalServerController *ServerController

// InitServerController initialisiert den globalen ServerController
func InitServerController(factorioPath, savePath string) {
	GlobalServerController = NewServerController(factorioPath, savePath)
}

// RestartServer startet den Server neu
func (sc *ServerController) RestartServer() error {
	// Stoppe den Server
	if err := sc.StopServer(); err != nil {
		return fmt.Errorf("fehler beim Stoppen des Servers: %w", err)
	}

	// Warte kurz
	time.Sleep(2 * time.Second)

	// Starte den Server
	if err := sc.StartServer(); err != nil {
		return fmt.Errorf("fehler beim Starten des Servers: %w", err)
	}

	return nil
}

// StopServer stoppt den Factorio-Server
func (sc *ServerController) StopServer() error {
	// Verwende das bestehende Factorio-Support-Modul
	if support.Factorio.IsRunning() {
		// Sende Quit-Befehl
		support.Factorio.Send("/quit")

		// Warte auf Beendigung
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return fmt.Errorf("timeout beim Warten auf Server-Stop")
			case <-ticker.C:
				if !support.Factorio.IsRunning() {
					return nil
				}
			}
		}
	}

	return nil
}

// StartServer startet den Factorio-Server
func (sc *ServerController) StartServer() error {
	// Wenn das Support-Modul einen laufenden Server hat, verwende dieses
	if support.Factorio.IsRunning() {
		return fmt.Errorf("server läuft bereits")
	}

	// Starte über das Support-Modul
	support.Factorio.Start(nil)

	return nil
}

// IsRunning prüft ob der Server läuft
func (sc *ServerController) IsRunning() bool {
	return support.Factorio.IsRunning()
}

// SendCommand sendet einen Befehl an den Server
func (sc *ServerController) SendCommand(command string) bool {
	return support.Factorio.Send(command)
}

// SaveGame speichert das aktuelle Spiel
func (sc *ServerController) SaveGame() error {
	if !sc.IsRunning() {
		return fmt.Errorf("server läuft nicht")
	}

	support.Factorio.Send("/save")
	return nil
}

// GetServerStatus gibt den Server-Status zurück
func (sc *ServerController) GetServerStatus() string {
	if !support.Factorio.IsRunning() {
		return "offline"
	}
	if support.Factorio.IsPaused() {
		return "paused"
	}
	return "running"
}

// GetExecutablePath gibt den Pfad zur Factorio-Executable zurück
func (sc *ServerController) GetExecutablePath() string {
	paths := []string{
		filepath.Join(sc.FactorioPath, "bin", "x64", "factorio"),
		filepath.Join(sc.FactorioPath, "bin", "x64", "factorio.exe"),
		filepath.Join(sc.FactorioPath, "factorio"),
		filepath.Join(sc.FactorioPath, "factorio.exe"),
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	// Fallback auf Config
	return support.Config.Executable
}
