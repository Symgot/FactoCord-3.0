package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factoriomod"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// handleSaveConfirmation verarbeitet die Bestätigung zum Speichern
func handleSaveConfirmation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := getUserID(i)

	// Acknowledge die Interaktion
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	// Hole Session
	session, err := globalSessionManager.FinalizeSession(userID)
	if err != nil {
		updateMessage(s, i.Interaction, "❌ Fehler: "+err.Error())
		return
	}

	if !session.HasChanges() {
		updateMessage(s, i.Interaction, "❌ Keine Änderungen zum Speichern vorhanden.")
		return
	}

	// Wende Änderungen an
	writer := factoriomod.NewSettingsWriter(
		support.ResolveFactorioPath(),
		"./backups",
	)

	if err := writer.ApplyChanges(session.ModName, session.Changes); err != nil {
		updateMessage(s, i.Interaction, "❌ Fehler beim Speichern: "+err.Error())
		log.Printf("[ModSettings] Fehler beim Speichern für User %s: %v", userID, err)
		return
	}

	// Erfolgsmeldung in DM
	successEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Einstellungen gespeichert",
		Description: fmt.Sprintf("**%s**: %d Einstellung(en) wurden erfolgreich gespeichert.", session.ModName, len(session.Changes)),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Geänderte Settings",
				Value:  formatChangesList(session.Changes),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Der Server wird neu gestartet...",
		},
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{successEmbed},
		Components: &[]discordgo.MessageComponent{},
	})

	// Log
	log.Printf("[ModSettings] User %s hat Settings für %s gespeichert: %v", userID, session.ModName, session.Changes)

	// Sende Ankündigung im allgemeinen Channel
	sendRestartAnnouncement(s, userID, session.ModName, session.Changes)

	// Starte Server neu
	if factoriomod.GlobalServerController != nil {
		go func() {
			if err := factoriomod.GlobalServerController.RestartServer(); err != nil {
				log.Printf("[ModSettings] Fehler beim Server-Neustart: %v", err)
			}
		}()
	}
}

// handleCancelConfirmation verarbeitet das Abbrechen der Änderungen
func handleCancelConfirmation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := getUserID(i)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	if err := globalSessionManager.CancelSession(userID); err != nil {
		updateMessage(s, i.Interaction, "❌ Fehler: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Änderungen verworfen",
		Description: "Alle ausstehenden Änderungen wurden verworfen.",
		Color:       0x0099FF,
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{},
	})
}

// sendRestartAnnouncement sendet eine Ankündigung im Factorio-Channel
func sendRestartAnnouncement(s *discordgo.Session, userID, modName string, changes map[string]interface{}) {
	// Hole den Factorio-Channel
	channelID := support.Config.FactorioChannelID
	if channelID == "" {
		return
	}

	// Hole Benutzername
	user, err := s.User(userID)
	username := "Unbekannt"
	if err == nil && user != nil {
		username = user.Username
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🔄 Server wird neu gestartet",
		Description: fmt.Sprintf("**%s** hat Mod-Einstellungen für **%s** angepasst.", username, modName),
		Color:       0xFFAA00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Änderungen",
				Value:  formatChangesList(changes),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Der Server wird in Kürze neu gestartet.",
		},
	}

	s.ChannelMessageSendEmbed(channelID, embed)
}

// formatChangesList formatiert die Änderungen als String-Liste
func formatChangesList(changes map[string]interface{}) string {
	if len(changes) == 0 {
		return "Keine Änderungen"
	}

	result := ""
	count := 0
	for key, value := range changes {
		if count >= 10 {
			result += fmt.Sprintf("\n... und %d weitere", len(changes)-10)
			break
		}
		result += fmt.Sprintf("• `%s`: %v\n", key, formatSettingValue(value))
		count++
	}

	return result
}

// SaveSettingsDirectly speichert Settings ohne Benutzerinteraktion
func SaveSettingsDirectly(modName string, changes map[string]interface{}) error {
	writer := factoriomod.NewSettingsWriter(
		support.ResolveFactorioPath(),
		"./backups",
	)

	return writer.ApplyChanges(modName, changes)
}
