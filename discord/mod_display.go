package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

const MODS_PER_PAGE = 8

// DisplayModList zeigt alle Mods in einer modernen Tabellen-Ansicht mit Pagination
func DisplayModList(s *discordgo.Session, channelID string, mods []models.ModInfo) error {
	return displayModListPage(s, channelID, nil, mods, 0)
}

// displayModListPage zeigt eine Seite der Mod-Liste
func displayModListPage(s *discordgo.Session, channelID string, interaction *discordgo.Interaction, mods []models.ModInfo, page int) error {
	totalPages := (len(mods) + MODS_PER_PAGE - 1) / MODS_PER_PAGE
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * MODS_PER_PAGE
	end := start + MODS_PER_PAGE
	if end > len(mods) {
		end = len(mods)
	}

	// Erstelle moderne Tabellen-Darstellung
	embed := &discordgo.MessageEmbed{
		Title: "🔧 Mod-Einstellungen Manager",
		Description: fmt.Sprintf("```\n%-25s │ Status │ 🎮 Game │ 🗺️ Map\n%s\n```",
			"Mod-Name", strings.Repeat("─", 50)),
		Color: 0x5865F2, // Discord Blurple
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("📄 Seite %d/%d • %d Mods mit Einstellungen gefunden", page+1, totalPages, len(mods)),
		},
	}

	// Baue Tabellen-Inhalt
	var tableContent strings.Builder
	tableContent.WriteString("```\n")
	tableContent.WriteString(fmt.Sprintf("%-25s │ Status │ 🎮 Game │ 🗺️ Map\n", "Mod-Name"))
	tableContent.WriteString(strings.Repeat("─", 50) + "\n")

	for _, mod := range mods[start:end] {
		status := "✅ ON "
		if !mod.Enabled {
			status = "⚫ OFF"
		}

		gameCount, mapCount := mod.CountSettings()
		modName := mod.Name
		if len(modName) > 23 {
			modName = modName[:20] + "..."
		}

		tableContent.WriteString(fmt.Sprintf("%-25s │ %s │   %3d  │  %3d\n",
			modName, status, gameCount, mapCount))
	}
	tableContent.WriteString("```")

	embed.Description = tableContent.String()

	// Erstelle Select-Menü für Mod-Auswahl
	options := make([]discordgo.SelectMenuOption, 0)
	for _, mod := range mods[start:end] {
		emoji := "✅"
		if !mod.Enabled {
			emoji = "⚫"
		}
		gameCount, mapCount := mod.CountSettings()

		options = append(options, discordgo.SelectMenuOption{
			Label:       mod.Name,
			Value:       fmt.Sprintf("mod_select_%s", mod.Name),
			Description: fmt.Sprintf("🎮 %d Game | 🗺️ %d Map Settings", gameCount, mapCount),
			Emoji: discordgo.ComponentEmoji{
				Name: emoji,
			},
		})
	}

	// Komponenten
	components := []discordgo.MessageComponent{
		// Select-Menü für Mod-Auswahl
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "mod_select_menu",
					Placeholder: "🔍 Wähle einen Mod zum Bearbeiten...",
					Options:     options,
				},
			},
		},
	}

	// Pagination-Buttons nur wenn mehr als eine Seite
	if totalPages > 1 {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "◀️ Vorherige",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("modlist_page_%d", page-1),
					Disabled: page == 0,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Seite %d/%d", page+1, totalPages),
					Style:    discordgo.SecondaryButton,
					CustomID: "modlist_page_info",
					Disabled: true,
				},
				discordgo.Button{
					Label:    "Nächste ▶️",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("modlist_page_%d", page+1),
					Disabled: page >= totalPages-1,
				},
			},
		})
	}

	// Info-Button-Reihe
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "🔄 Aktualisieren",
				Style:    discordgo.PrimaryButton,
				CustomID: "modlist_refresh",
			},
			discordgo.Button{
				Label:    "❓ Hilfe",
				Style:    discordgo.SecondaryButton,
				CustomID: "modlist_help",
			},
		},
	})

	if interaction != nil {
		_, err := s.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &components,
		})
		return err
	}

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	return err
}

// displayModListUpdate aktualisiert eine bestehende Mod-Liste Nachricht
func displayModListUpdate(s *discordgo.Session, i *discordgo.Interaction, mods []models.ModInfo) error {
	return displayModListPage(s, "", i, mods, 0)
}

// DisplayModInfo zeigt detaillierte Informationen über einen Mod
func DisplayModInfo(s *discordgo.Session, channelID string, mod *models.ModInfo) error {
	gameCount, mapCount := mod.CountSettings()

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📦 %s", mod.Name),
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Version",
				Value:  mod.Version,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  enabledString(mod.Enabled),
				Inline: true,
			},
			{
				Name:   "Game Settings",
				Value:  fmt.Sprintf("%d", gameCount),
				Inline: true,
			},
			{
				Name:   "Map Settings",
				Value:  fmt.Sprintf("%d", mapCount),
				Inline: true,
			},
		},
	}

	if mod.Description != "" {
		embed.Description = mod.Description
	}

	if mod.Author != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Autor: %s", mod.Author),
		}
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

func enabledString(enabled bool) string {
	if enabled {
		return "✅ Aktiviert"
	}
	return "⚫ Deaktiviert"
}
