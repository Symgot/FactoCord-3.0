package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

// DisplayChangesPreview zeigt eine Vorschau aller Änderungen (für Interaction-Response)
func DisplayChangesPreview(s *discordgo.Session, i *discordgo.Interaction, session *models.Session) error {
	if !session.HasChanges() {
		return updateInteractionWithError(s, i, "❌ Es gibt keine Änderungen zum Speichern.")
	}

	var changesText strings.Builder
	changesText.WriteString("```yml\n")
	for key, value := range session.Changes {
		changesText.WriteString(fmt.Sprintf("%s: %s\n", key, formatSettingValue(value)))
	}
	changesText.WriteString("```")

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 Änderungen für %s", session.ModName),
		Description: changesText.String(),
		Color:       0xFEE75C,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⚠️ Warnung",
				Value:  "Der Server wird nach dem Speichern **neu gestartet**!",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%d Änderung(en)", len(session.Changes)),
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Speichern & Neustarten",
					Style:    discordgo.SuccessButton,
					CustomID: "confirm_save",
				},
				discordgo.Button{
					Label:    "❌ Abbrechen",
					Style:    discordgo.DangerButton,
					CustomID: "confirm_cancel",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "📋 Zurück zur Bearbeitung",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("back_to_mod_%s", session.ModName),
				},
			},
		},
	}

	s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	return nil
}

// DisplayChangesPreviewChannel zeigt eine Vorschau aller Änderungen (für Channel-Nachrichten)
func DisplayChangesPreviewChannel(s *discordgo.Session, channelID, userID string) error {
	session := globalSessionManager.GetSession(userID)
	if session == nil {
		return &SessionError{Message: "keine aktive Session gefunden"}
	}

	if !session.HasChanges() {
		_, err := s.ChannelMessageSend(channelID, "❌ Es gibt keine Änderungen zum Speichern.")
		return err
	}

	var changesText strings.Builder
	changesText.WriteString("```yml\n")
	for key, value := range session.Changes {
		changesText.WriteString(fmt.Sprintf("%s: %s\n", key, formatSettingValue(value)))
	}
	changesText.WriteString("```")

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 Änderungen für %s", session.ModName),
		Description: changesText.String(),
		Color:       0xFEE75C,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "⚠️ Der Server wird nach dem Speichern neu gestartet!",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Speichern & Neustarten",
					Style:    discordgo.SuccessButton,
					CustomID: "confirm_save",
				},
				discordgo.Button{
					Label:    "❌ Abbrechen",
					Style:    discordgo.DangerButton,
					CustomID: "confirm_cancel",
				},
			},
		},
	}

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	return err
}

func updateInteractionWithError(s *discordgo.Session, i *discordgo.Interaction, msg string) error {
	content := msg
	s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Content: &content,
	})
	return nil
}

// DisplayChangesPreviewEmbed erstellt nur das Embed für die Änderungsvorschau
func DisplayChangesPreviewEmbed(session *models.Session) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 Änderungen für %s", session.ModName),
		Description: "Folgende Einstellungen werden geändert:",
		Color:       0xFFAA00,
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	for key, value := range session.Changes {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   key,
			Value:  fmt.Sprintf("`%v`", formatSettingValue(value)),
			Inline: false,
		})
	}

	return embed
}

// DisplayNoChangesMessage zeigt eine Nachricht an, wenn keine Änderungen vorhanden sind
func DisplayNoChangesMessage(s *discordgo.Session, channelID string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "ℹ️ Keine Änderungen",
		Description: "Es wurden keine Änderungen vorgenommen.\n\nVerwende `!mods` um mit der Bearbeitung zu beginnen.",
		Color:       0x0099FF,
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// BuildChangesEmbed erstellt ein Embed mit den Änderungen für einen Mod
func BuildChangesEmbed(modName string, changes map[string]interface{}) *discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, 0)

	for key, value := range changes {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   key,
			Value:  fmt.Sprintf("`%v`", formatSettingValue(value)),
			Inline: true,
		})
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔧 Änderungen: %s", modName),
		Description: fmt.Sprintf("%d Einstellung(en) geändert", len(changes)),
		Color:       0x00FF00,
		Fields:      fields,
	}
}

// DisplaySuccessMessage zeigt eine Erfolgsmeldung an
func DisplaySuccessMessage(s *discordgo.Session, channelID, modName string, changeCount int) error {
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Einstellungen gespeichert",
		Description: fmt.Sprintf("**%s**: %d Einstellung(en) wurden erfolgreich gespeichert.", modName, changeCount),
		Color:       0x00FF00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Der Server wird neu gestartet...",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// DisplayErrorMessage zeigt eine Fehlermeldung an
func DisplayErrorMessage(s *discordgo.Session, channelID, errorMsg string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Fehler",
		Description: errorMsg,
		Color:       0xFF0000,
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}
