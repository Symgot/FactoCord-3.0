package discord

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/commands"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factoriomod"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// HandleModSettingsMessage verarbeitet Mod-Settings-Befehle in DMs
func HandleModSettingsMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// Nur DMs verarbeiten
	channel, err := s.Channel(m.ChannelID)
	if err != nil || channel.Type != discordgo.ChannelTypeDM {
		return false
	}

	// Ignoriere Bot-Nachrichten
	if m.Author.ID == s.State.User.ID {
		return false
	}

	// Verifikation prüfen
	if !IsUserVerified(m.Author.ID) {
		return false // Lasse andere Handler übernehmen
	}

	raw := strings.TrimSpace(m.Content)
	prefix := strings.TrimSpace(support.Config.Prefix)

	// Entferne Prefix falls vorhanden (case-insensitive)
	if prefix != "" {
		rawLower := strings.ToLower(raw)
		prefixLower := strings.ToLower(prefix)
		if strings.HasPrefix(rawLower, prefixLower) && len(raw) >= len(prefix) {
			raw = strings.TrimSpace(raw[len(prefix):])
		}
	}

	// Legacy: "!" Commands in DMs
	if strings.HasPrefix(raw, "!") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "!"))
	}

	contentLower := strings.ToLower(strings.TrimSpace(raw))

	switch {
	case contentLower == "mods":
		handleModListRequest(s, m)
		return true
	case contentLower == "cancel":
		handleCancelRequest(s, m)
		return true
	case contentLower == "save":
		handleSaveRequest(s, m)
		return true
	case contentLower == "modshelp" || contentLower == "help" || contentLower == "?":
		handleModsHelpRequest(s, m)
		return true
	}

	// Erweiterte DM-Commands: "server ...", "backups ...", "mods reload"
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return false
	}

	switch strings.ToLower(parts[0]) {
	case "server":
		handleServerDMCommand(s, m, parts[1:])
		return true
	case "backup", "backups":
		handleBackupsDMCommand(s, m, parts[1:])
		return true
	case "mods":
		// z.B. "mods reload"
		if len(parts) >= 2 && strings.ToLower(parts[1]) == "reload" {
			handleModsReloadRequest(s, m)
			return true
		}
	}

	return false
}

func handleModsReloadRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	if factoriomod.GlobalModManager == nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Mod-Manager ist nicht initialisiert.")
		return
	}
	if err := factoriomod.GlobalModManager.DiscoverMods(); err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Fehler beim Neuladen der Mods: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Mods neu geladen (%d).", factoriomod.GlobalModManager.GetModCount()))
}

func handleServerDMCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if factoriomod.GlobalServerController == nil {
		s.ChannelMessageSend(m.ChannelID, "❌ ServerController ist nicht initialisiert.")
		return
	}

	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID,
			"ℹ️ Nutzung:\n"+
				"- `"+support.Config.Prefix+"server status`\n"+
				"- `"+support.Config.Prefix+"server save` (Admin)\n"+
				"- `"+support.Config.Prefix+"server restart` (Admin)\n"+
				"- `"+support.Config.Prefix+"server stop` (Admin)\n"+
				"- `"+support.Config.Prefix+"server start` (Admin)")
		return
	}

	action := strings.ToLower(args[0])
	isAdmin := commands.CheckAdmin(m.Author.ID)

	switch action {
	case "status":
		status := factoriomod.GlobalServerController.GetServerStatus()
		s.ChannelMessageSend(m.ChannelID, "📡 Server-Status: **"+status+"**")

	case "save":
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "❌ Du musst Admin sein, um den Server zu speichern.")
			return
		}
		if err := factoriomod.GlobalServerController.SaveGame(); err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Save fehlgeschlagen: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ Save wurde ausgelöst.")

	case "restart":
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "❌ Du musst Admin sein, um den Server neu zu starten.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "🔄 Restart wird ausgeführt...")
		go func() {
			if err := factoriomod.GlobalServerController.RestartServer(); err != nil {
				s.ChannelMessageSend(m.ChannelID, "❌ Restart fehlgeschlagen: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "✅ Server neu gestartet.")
		}()

	case "stop":
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "❌ Du musst Admin sein, um den Server zu stoppen.")
			return
		}
		if err := factoriomod.GlobalServerController.StopServer(); err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Stop fehlgeschlagen: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ Server gestoppt.")

	case "start":
		if !isAdmin {
			s.ChannelMessageSend(m.ChannelID, "❌ Du musst Admin sein, um den Server zu starten.")
			return
		}
		if err := factoriomod.GlobalServerController.StartServer(); err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Start fehlgeschlagen: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "✅ Server gestartet.")

	default:
		s.ChannelMessageSend(m.ChannelID, "❌ Unbekannter server-Befehl. Nutze `"+support.Config.Prefix+"server` für Hilfe.")
	}
}

func handleBackupsDMCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	writer := factoriomod.NewSettingsWriter(
		support.ResolveFactorioPath(),
		"./backups",
	)

	if len(args) == 0 || strings.ToLower(args[0]) == "list" {
		backups, err := writer.ListBackups()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Fehler beim Laden der Backups: "+err.Error())
			return
		}
		if len(backups) == 0 {
			s.ChannelMessageSend(m.ChannelID, "ℹ️ Keine Backups gefunden.")
			return
		}

		sort.Strings(backups)
		if len(backups) > 25 {
			backups = backups[len(backups)-25:]
		}

		var b strings.Builder
		b.WriteString("💾 Verfügbare Backups (neueste oben):\n")
		for i := len(backups) - 1; i >= 0; i-- {
			b.WriteString("- " + backups[i] + "\n")
		}
		b.WriteString("\nNutzung: `" + support.Config.Prefix + "backups restore <filename>` (Admin)")
		s.ChannelMessageSend(m.ChannelID, b.String())
		return
	}

	action := strings.ToLower(args[0])
	if action != "restore" {
		s.ChannelMessageSend(m.ChannelID, "❌ Unbekannter backups-Befehl. Nutze `"+support.Config.Prefix+"backups list`.")
		return
	}

	if !commands.CheckAdmin(m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, "❌ Du musst Admin sein, um ein Backup wiederherzustellen.")
		return
	}
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Nutzung: `"+support.Config.Prefix+"backups restore <filename>`")
		return
	}

	backupFile := args[1]
	if err := writer.RestoreBackup(backupFile); err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Restore fehlgeschlagen: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "✅ Backup wiederhergestellt: `"+backupFile+"`\n🔄 Server wird neu gestartet...")
	if factoriomod.GlobalServerController != nil {
		go func() {
			_ = factoriomod.GlobalServerController.RestartServer()
		}()
	}
}

// handleModListRequest zeigt die verfügbaren Mods an
func handleModListRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Lade Mods neu
	if err := factoriomod.GlobalModManager.DiscoverMods(); err != nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Fehler beim Laden der Mods: "+err.Error())
		return
	}

	mods := factoriomod.GlobalModManager.GetModsWithSettings()
	if len(mods) == 0 {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Keine Mods mit konfigurierbaren Settings gefunden.")
		return
	}

	// Zeige Mod-Liste
	if err := DisplayModList(s, m.ChannelID, mods); err != nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Fehler beim Anzeigen der Mod-Liste: "+err.Error())
	}
}

// handleCancelRequest verwirft alle Änderungen
func handleCancelRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	session := globalSessionManager.GetSession(m.Author.ID)
	if session == nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Du hast keine aktive Bearbeitungssession.")
		return
	}

	if err := globalSessionManager.CancelSession(m.Author.ID); err != nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Fehler beim Abbrechen: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "✅ Alle Änderungen verworfen.")
}

// handleSaveRequest zeigt die Änderungsvorschau und Speicheroptionen
func handleSaveRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	session := globalSessionManager.GetSession(m.Author.ID)
	if session == nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Du hast keine aktive Bearbeitungssession.")
		return
	}

	if !session.HasChanges() {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Es gibt keine Änderungen zum Speichern.")
		return
	}

	// Zeige Änderungsvorschau
	if err := DisplayChangesPreviewChannel(s, m.ChannelID, m.Author.ID); err != nil {
		s.ChannelMessageSend(m.ChannelID,
			"❌ Fehler beim Anzeigen der Vorschau: "+err.Error())
	}
}

// handleModsHelpRequest zeigt die Hilfe für Mod-Settings an
func handleModsHelpRequest(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := support.Config.Prefix

	embed := &discordgo.MessageEmbed{
		Title:       "🔧 Mod-Settings Manager - Hilfe",
		Description: "Verwalte Factorio Mod-Einstellungen über Discord DMs.",
		Color:       0x00D4FF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   prefix + "mods",
				Value:  "Zeigt alle verfügbaren Mods mit Settings an",
				Inline: false,
			},
			{
				Name:   prefix + "save",
				Value:  "Zeigt ausstehende Änderungen und speichert sie",
				Inline: false,
			},
			{
				Name:   prefix + "cancel",
				Value:  "Verwirft alle ausstehenden Änderungen",
				Inline: false,
			},
			{
				Name:   "📝 Workflow",
				Value:  "1. `" + prefix + "mods` - Mod-Liste anzeigen\n2. Mod per Button auswählen\n3. Tab wählen (Game/Map Settings)\n4. Settings bearbeiten\n5. `" + prefix + "save` - Änderungen speichern",
				Inline: false,
			},
			{
				Name:   "⚠️ Hinweis",
				Value:  "Nach dem Speichern wird der Server automatisch neu gestartet, damit die Änderungen wirksam werden.",
				Inline: false,
			},
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
