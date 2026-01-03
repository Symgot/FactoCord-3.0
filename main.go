package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/discord"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factorio"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/utils"
)

var logger *utils.Logger

func main() {
	// Logger initialisieren
	logger = utils.NewLogger("FactoCord")
	logger.Info("FactoCord 3.0 Mod-Settings Manager startet...")

	// Automatisches Setup - erstelle benötigte Verzeichnisse und Dateien
	ensureDirectoriesExist()

	// Standard-Config laden
	support.Config.MustLoad()

	token := support.Config.DiscordToken
	if token == "" {
		logger.Fatal("Discord-Token nicht in Config gefunden")
	}

	// Discord-Session erstellen
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatal("Discord-Session-Fehler: %v", err)
	}

	// Factorio-Pfad aus Executable extrahieren oder Standard verwenden
	factorioPath := getFactorioPath()

	// Module initialisieren
	factorio.InitModManager(factorioPath, "")
	factorio.InitServerController(factorioPath, "")
	discord.InitSessionManager("./temp")

	// Verifikationsdaten laden
	if err := discord.LoadVerificationData(); err != nil {
		logger.Warn("Verifikationsdaten konnten nicht geladen werden: %v", err)
	}
	discord.StartVerificationCleanup()

	// Handler registrieren
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		handleReady(s, r)
	})

	// Message-Handler für Mod-Settings
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Versuche zuerst Mod-Settings zu verarbeiten
		if discord.HandleModSettingsMessage(s, m) {
			return
		}
		// Sonst normaler DM-Handler
		discord.HandleDMMessage(s, m)
	})

	// Component-Interaction-Handler
	dg.AddHandler(discord.HandleComponentInteraction)

	// Intents setzen
	dg.Identify.Intents = discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuilds

	// Session öffnen
	err = dg.Open()
	if err != nil {
		logger.Fatal("Fehler beim Öffnen der Session: %v", err)
	}

	logger.Info("✅ FactoCord läuft. Drücke CTRL+C zum Beenden.")

	// Warte auf Interrupt
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
	logger.Info("✅ FactoCord beendet.")
}

func getFactorioPath() string {
	// Versuche den Pfad aus der Executable zu extrahieren
	exec := support.Config.Executable
	if exec != "" {
		// Extrahiere das Basisverzeichnis aus dem Executable-Pfad
		// z.B. "../bin/x64/factorio" -> ".." oder "./factorio/bin/x64/factorio" -> "./factorio"
		// Wir gehen davon aus, dass die Struktur: [base]/bin/x64/factorio ist
		dir := filepath.Dir(exec) // bin/x64
		dir = filepath.Dir(dir)   // bin
		dir = filepath.Dir(dir)   // [base]

		// Falls der Pfad relativ ist, mache ihn absolut
		if absPath, err := filepath.Abs(dir); err == nil {
			if _, statErr := os.Stat(absPath); statErr == nil {
				logger.Info("Factorio-Pfad aus Executable ermittelt: %s", absPath)
				return absPath
			}
		}

		// Falls das nicht funktioniert, nutze ModListLocation
		if support.Config.ModListLocation != "" {
			// z.B. "./factorio/mods/mod-list.json" -> "./factorio"
			modsDir := filepath.Dir(support.Config.ModListLocation) // mods
			baseDir := filepath.Dir(modsDir)                        // factorio
			if absPath, err := filepath.Abs(baseDir); err == nil {
				if _, statErr := os.Stat(absPath); statErr == nil {
					logger.Info("Factorio-Pfad aus ModListLocation ermittelt: %s", absPath)
					return absPath
				}
			}
		}
	}

	// Standard-Pfade prüfen
	paths := []string{
		"./factorio",
		"..", // Falls FactoCord im Factorio-Verzeichnis läuft
		"/opt/factorio",
		"C:\\Program Files\\Factorio",
	}

	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		modsPath := filepath.Join(absPath, "mods")
		if _, err := os.Stat(modsPath); err == nil {
			logger.Info("Factorio-Pfad gefunden: %s", absPath)
			return absPath
		}
	}

	// Fallback: Verwende aktuelles Verzeichnis
	logger.Warn("Kein Factorio-Pfad gefunden, verwende Standard")
	return "."
}

func handleReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("✅ Bot eingeloggt als: %s#%s (ID: %s)",
		event.User.Username, event.User.Discriminator, event.User.ID)

	// Speichere Guild ID falls verfügbar
	if len(event.Guilds) > 0 {
		support.GuildID = event.Guilds[0].ID
	}
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
			logger.Warn("Konnte Verzeichnis %s nicht erstellen: %v", dir, err)
		}
	}

	// Erstelle verification.json falls nicht vorhanden
	verificationPath := support.Config.VerificationDataPath
	if verificationPath == "" {
		verificationPath = "./verification.json"
	}

	if _, err := os.Stat(verificationPath); os.IsNotExist(err) {
		// Erstelle leere Verifikationsdatei
		emptyData := map[string]interface{}{
			"discord_to_factorio": map[string]string{},
			"factorio_to_discord": map[string]string{},
		}
		data, _ := json.MarshalIndent(emptyData, "", "  ")
		if err := os.WriteFile(verificationPath, data, 0644); err != nil {
			logger.Warn("Konnte verification.json nicht erstellen: %v", err)
		} else {
			logger.Info("✅ verification.json erstellt")
		}
	}

	// Erstelle mods-Verzeichnis falls benötigt
	modsPath := filepath.Dir(support.Config.ModListLocation)
	if modsPath != "" && modsPath != "." {
		if err := os.MkdirAll(modsPath, 0755); err != nil {
			logger.Warn("Konnte Mods-Verzeichnis %s nicht erstellen: %v", modsPath, err)
		}
	}

	logger.Info("✅ Verzeichnisstruktur geprüft")
}
