package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/discord"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factoriomod"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

var closing = false

func main() {
	if support.FactoCordVersion == "" {
		support.FactoCordVersion = "development"
	}
	fmt.Printf("Welcome to FactoCord %s!\n", support.FactoCordVersion)

	// Automatisches Setup - erstelle benötigte Verzeichnisse und Dateien
	ensureDirectoriesExist()

	// Config laden
	support.Config.MustLoad()

	// Discord-Session starten
	discord.StartSession()

	// Konsolen-Input Handler starten
	go console()

	// Factorio-Server initialisieren mit Log-Verarbeitung
	support.Factorio.Init(discord.ProcessFactorioLogLine)

	// Factorio-Pfad ermitteln für Mod-Manager
	factorioPath := getFactorioPath()

	// Mod-Settings Manager initialisieren (optional)
	factoriomod.InitModManager(factorioPath, "")
	factoriomod.InitServerController(factorioPath, "")

	// Session-Manager für DM-basierte Bearbeitung
	discord.InitSessionManager("./temp")

	// Verifikationsdaten laden (für DM-Mod-Settings)
	if err := discord.LoadVerificationData(); err != nil {
		fmt.Printf("Verifikationsdaten konnten nicht geladen werden: %v\n", err)
	}
	discord.StartVerificationCleanup()

	// Discord Chat-Handler und Bot initialisieren
	discord.Init()

	// Auf Interrupt warten
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	<-sc

	closing = true

	discord.Close()

	for support.Factorio.IsStopping() {
		time.Sleep(100 * time.Millisecond)
	}
	if support.Factorio.IsRunning() {
		fmt.Println("Waiting for factorio server to exit...")
		err := support.Factorio.Process.Wait()
		if support.Factorio.Process.ProcessState.Exited() {
			fmt.Println("\nFactorio server was closed, exit code", support.Factorio.Process.ProcessState.ExitCode())
		} else {
			fmt.Println("\nError waiting for factorio to exit")
			support.Panik(err, "Error waiting for factorio to exit")
		}
	}
}

func console() {
	Console := bufio.NewReader(os.Stdin)
	for !closing {
		line, _, err := Console.ReadLine()
		if err != nil {
			support.Panik(err, "An error occurred when attempting to read the input to pass as input to the console")
			return
		}
		support.Factorio.Send(string(line))
	}
}

func getFactorioPath() string {
	// Versuche den Pfad aus der Executable zu extrahieren
	exec := support.Config.Executable
	if exec != "" {
		// Extrahiere das Basisverzeichnis aus dem Executable-Pfad
		// z.B. "../bin/x64/factorio" -> ".." oder "./factorio/bin/x64/factorio" -> "./factorio"
		dir := filepath.Dir(exec) // bin/x64
		dir = filepath.Dir(dir)   // bin
		dir = filepath.Dir(dir)   // [base]

		// Falls der Pfad relativ ist, mache ihn absolut
		if absPath, err := filepath.Abs(dir); err == nil {
			if _, statErr := os.Stat(absPath); statErr == nil {
				fmt.Printf("Factorio-Pfad aus Executable ermittelt: %s\n", absPath)
				return absPath
			}
		}

		// Falls das nicht funktioniert, nutze ModListLocation
		if support.Config.ModListLocation != "" {
			modsDir := filepath.Dir(support.Config.ModListLocation)
			baseDir := filepath.Dir(modsDir)
			if absPath, err := filepath.Abs(baseDir); err == nil {
				if _, statErr := os.Stat(absPath); statErr == nil {
					fmt.Printf("Factorio-Pfad aus ModListLocation ermittelt: %s\n", absPath)
					return absPath
				}
			}
		}
	}

	// Standard-Pfade prüfen
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

// ensureDirectoriesExist erstellt alle benötigten Verzeichnisse und Dateien automatisch
func ensureDirectoriesExist() {
	directories := []string{
		"./temp",
		"./temp/temp_settings",
		"./backups",
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Konnte Verzeichnis %s nicht erstellen: %v\n", dir, err)
		}
	}

	// Erstelle verification.json falls nicht vorhanden
	verificationPath := "./verification.json"

	if _, err := os.Stat(verificationPath); os.IsNotExist(err) {
		emptyData := map[string]interface{}{
			"discord_to_factorio": map[string]string{},
			"factorio_to_discord": map[string]string{},
		}
		data, _ := json.MarshalIndent(emptyData, "", "  ")
		if err := os.WriteFile(verificationPath, data, 0644); err != nil {
			fmt.Printf("Konnte verification.json nicht erstellen: %v\n", err)
		}
	}
}
