package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
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

	content := strings.ToLower(strings.TrimSpace(m.Content))
	prefix := strings.ToLower(strings.TrimSpace(support.Config.Prefix))

	// Entferne Prefix falls vorhanden (case-insensitive)
	if strings.HasPrefix(content, prefix) {
		content = strings.TrimSpace(strings.TrimPrefix(content, prefix))
	}

	switch {
	case content == "mods" || content == "!mods":
		handleModListRequest(s, m)
		return true
	case content == "cancel" || content == "!cancel":
		handleCancelRequest(s, m)
		return true
	case content == "save" || content == "!save":
		handleSaveRequest(s, m)
		return true
	case content == "modshelp" || content == "!modshelp":
		handleModsHelpRequest(s, m)
		return true
	}

	return false
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
