package discord

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

const SETTINGS_PER_PAGE = 8

// displayTabSelection zeigt die Tab-Auswahl im Browser-Stil
func displayTabSelection(s *discordgo.Session, i *discordgo.Interaction, mod *models.ModInfo, activeTab string) {
	gameCount, mapCount := mod.CountSettings()

	// Browser-Style Tab-Visualisierung
	var tabHeader strings.Builder
	tabHeader.WriteString("```\n")

	if activeTab == "game" {
		tabHeader.WriteString("╭─────────────────╮                    \n")
		tabHeader.WriteString("│ 🎮 Game Settings │────────────────────╮\n")
		tabHeader.WriteString("╰─────────────────╯   🗺️ Map Settings   │\n")
		tabHeader.WriteString("─────────────────────────────────────────╯\n")
	} else if activeTab == "map" {
		tabHeader.WriteString("                    ╭─────────────────╮\n")
		tabHeader.WriteString("╭───────────────────│ 🗺️ Map Settings │\n")
		tabHeader.WriteString("│   🎮 Game Settings ╰─────────────────╯\n")
		tabHeader.WriteString("╰─────────────────────────────────────────\n")
	} else {
		// Übersicht - beide Tabs neutral
		tabHeader.WriteString("╭───────────────────┬─────────────────╮\n")
		tabHeader.WriteString("│ 🎮 Game Settings  │ 🗺️ Map Settings │\n")
		tabHeader.WriteString("╰───────────────────┴─────────────────╯\n")
	}
	tabHeader.WriteString("```")

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📦 %s", mod.Name),
		Description: tabHeader.String(),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎮 Game Settings",
				Value:  fmt.Sprintf("`%d` Einstellungen", gameCount),
				Inline: true,
			},
			{
				Name:   "🗺️ Map Settings",
				Value:  fmt.Sprintf("`%d` Einstellungen", mapCount),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Wähle einen Tab um die Einstellungen anzuzeigen",
		},
	}

	// Tab-Buttons im Browser-Style
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🎮 Game Settings",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("tab_game_%s", mod.Name),
					Disabled: gameCount == 0,
				},
				discordgo.Button{
					Label:    "🗺️ Map Settings",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("tab_map_%s", mod.Name),
					Disabled: mapCount == 0,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "📋 Zurück zur Mod-Liste",
					Style:    discordgo.SecondaryButton,
					CustomID: "back_to_mods",
				},
			},
		},
	}

	s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
}

// displaySettingsTab zeigt die Settings eines Tabs an
func displaySettingsTab(s *discordgo.Session, i *discordgo.Interaction,
	mod *models.ModInfo, tabType string, page int) {

	var settings map[string]interface{}
	var tabTitle string
	var tabEmoji string
	var activeColor int

	if tabType == "game" {
		settings = mod.GetAllGameSettings()
		tabTitle = "Game Settings"
		tabEmoji = "🎮"
		activeColor = 0x57F287 // Grün
	} else {
		settings = mod.GetAllMapSettings()
		tabTitle = "Map Settings"
		tabEmoji = "🗺️"
		activeColor = 0x3498DB // Blau
	}

	if len(settings) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s %s - %s", tabEmoji, mod.Name, tabTitle),
			Description: "⚠️ Keine Einstellungen in diesem Tab vorhanden.",
			Color:       0xFEE75C, // Gelb für Warnung
		}

		s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
			Components: &[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "🔙 Zurück",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("back_to_mod_%s", mod.Name),
						},
					},
				},
			},
		})
		return
	}

	// Sortiere Settings alphabetisch
	keys := getSettingKeys(settings)
	totalSettings := len(keys)
	totalPages := (totalSettings + SETTINGS_PER_PAGE - 1) / SETTINGS_PER_PAGE

	// Seitenvalidierung
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	// Berechne Start und Ende für aktuelle Seite
	start := page * SETTINGS_PER_PAGE
	end := start + SETTINGS_PER_PAGE
	if end > totalSettings {
		end = totalSettings
	}

	// Browser-Tab Header
	var header strings.Builder
	header.WriteString("```\n")
	if tabType == "game" {
		header.WriteString("┏━━━━━━━━━━━━━━━━┓                   \n")
		header.WriteString("┃ 🎮 GAME ACTIVE ┃─── 🗺️ Map ────────┐\n")
		header.WriteString("┗━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━┛\n")
	} else {
		header.WriteString("                   ┏━━━━━━━━━━━━━━━━┓\n")
		header.WriteString("┌─── 🎮 Game ──────┃ 🗺️ MAP ACTIVE ┃\n")
		header.WriteString("┗━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━┛\n")
	}
	header.WriteString("```")

	// Settings-Tabelle
	var tableContent strings.Builder
	tableContent.WriteString("```yml\n")
	for _, key := range keys[start:end] {
		value := settings[key]
		formattedValue := formatSettingValue(value)
		displayKey := key
		if len(displayKey) > 28 {
			displayKey = displayKey[:25] + "..."
		}
		tableContent.WriteString(fmt.Sprintf("%-30s: %s\n", displayKey, formattedValue))
	}
	tableContent.WriteString("```")

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", tabEmoji, mod.Name),
		Description: header.String() + "\n" + tableContent.String(),
		Color:       activeColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("📄 Seite %d/%d • %d Einstellungen", page+1, totalPages, totalSettings),
		},
	}

	// Erstelle Select-Menü für Setting-Bearbeitung
	settingOptions := make([]discordgo.SelectMenuOption, 0)
	for _, key := range keys[start:end] {
		value := settings[key]
		settingOptions = append(settingOptions, discordgo.SelectMenuOption{
			Label:       truncateString(key, 25),
			Value:       fmt.Sprintf("edit_%s_%s_%s", mod.Name, tabType, key),
			Description: truncateString(fmt.Sprintf("Aktuell: %s", formatSettingValue(value)), 50),
		})
	}

	// Erstelle Komponenten
	components := []discordgo.MessageComponent{
		// Tab-Wechsel Buttons (Browser-Style)
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🎮 Game",
					Style:    buttonStyleForTab("game", tabType),
					CustomID: fmt.Sprintf("tab_game_%s", mod.Name),
				},
				discordgo.Button{
					Label:    "🗺️ Map",
					Style:    buttonStyleForTab("map", tabType),
					CustomID: fmt.Sprintf("tab_map_%s", mod.Name),
				},
			},
		},
		// Setting-Auswahl
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("setting_select_%s_%s", mod.Name, tabType),
					Placeholder: "⚙️ Einstellung zum Bearbeiten auswählen...",
					Options:     settingOptions,
				},
			},
		},
	}

	// Pagination nur wenn nötig
	if totalPages > 1 {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "◀️",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("page_%s_%s_%d", mod.Name, tabType, page-1),
					Disabled: page == 0,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("%d/%d", page+1, totalPages),
					Style:    discordgo.SecondaryButton,
					CustomID: "page_info",
					Disabled: true,
				},
				discordgo.Button{
					Label:    "▶️",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("page_%s_%s_%d", mod.Name, tabType, page+1),
					Disabled: page >= totalPages-1,
				},
			},
		})
	}

	// Navigation
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "📋 Mod-Liste",
				Style:    discordgo.SecondaryButton,
				CustomID: "back_to_mods",
			},
			discordgo.Button{
				Label:    "💾 Speichern",
				Style:    discordgo.SuccessButton,
				CustomID: "open_save_dialog",
			},
		},
	})

	s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
}

// buttonStyleForTab gibt den Button-Style basierend auf aktivem Tab zurück
func buttonStyleForTab(targetTab, activeTab string) discordgo.ButtonStyle {
	if targetTab == activeTab {
		return discordgo.PrimaryButton
	}
	return discordgo.SecondaryButton
}

// truncateString kürzt einen String auf maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// buildSettingsComponents erstellt die Buttons für die Settings-Ansicht
func buildSettingsComponents(modName, tabType string, page, totalSettings int) []discordgo.MessageComponent {
	totalPages := (totalSettings + SETTINGS_PER_PAGE - 1) / SETTINGS_PER_PAGE

	var components []discordgo.MessageComponent

	// Navigation-Reihe
	navButtons := []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "◀️ Zurück",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("page_prev_%s_%s_%d", modName, tabType, page),
			Disabled: page == 0,
		},
		discordgo.Button{
			Label:    "✏️ Bearbeiten",
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("edit_%s_%s_%d", modName, tabType, page),
		},
		discordgo.Button{
			Label:    "Weiter ▶️",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("page_next_%s_%s_%d", modName, tabType, page),
			Disabled: page >= totalPages-1,
		},
	}

	components = append(components, discordgo.ActionsRow{
		Components: navButtons,
	})

	// Back-Button-Reihe
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "🔙 Zurück zu Tabs",
				Style:    discordgo.SecondaryButton,
				CustomID: "back_to_tabs",
			},
			discordgo.Button{
				Label:    "📋 Zurück zu Mods",
				Style:    discordgo.SecondaryButton,
				CustomID: "back_to_mods",
			},
		},
	})

	return components
}

// getSettingKeys gibt die sortierten Keys eines Settings-Maps zurück
func getSettingKeys(settings map[string]interface{}) []string {
	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// formatSettingValue formatiert einen Setting-Wert für die Anzeige
func formatSettingValue(value interface{}) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "✅ true"
		}
		return "❌ false"
	case float64:
		// Zeige ohne Dezimalstellen wenn möglich
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)
	case string:
		if len(v) > 40 {
			return "\"" + v[:37] + "...\""
		}
		return "\"" + v + "\""
	case map[string]interface{}:
		// Factorio speichert oft Werte als {"value": actualValue}
		if actualValue, exists := v["value"]; exists {
			return formatSettingValue(actualValue)
		}
		return "[Object]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// DisplaySettingDetails zeigt Details eines einzelnen Settings
func DisplaySettingDetails(s *discordgo.Session, channelID string, modName, settingKey string, value interface{}) error {
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("⚙️ %s", settingKey),
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Mod",
				Value:  modName,
				Inline: true,
			},
			{
				Name:   "Aktueller Wert",
				Value:  fmt.Sprintf("`%v`", formatSettingValue(value)),
				Inline: true,
			},
			{
				Name:   "Typ",
				Value:  getValueType(value),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Antworte mit dem neuen Wert um diese Einstellung zu ändern",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// displayEditSettingModal zeigt einen Modal-Dialog zum Bearbeiten einer Einstellung
func displayEditSettingModal(s *discordgo.Session, i *discordgo.Interaction, modName, tabType, settingKey string, currentValue interface{}) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("setting_modal_%s_%s_%s", modName, tabType, settingKey),
			Title:    fmt.Sprintf("⚙️ %s bearbeiten", truncateString(settingKey, 30)),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "setting_value",
							Label:       "Neuer Wert",
							Style:       discordgo.TextInputShort,
							Placeholder: fmt.Sprintf("Aktuell: %s", formatSettingValue(currentValue)),
							Required:    true,
							MinLength:   1,
							MaxLength:   200,
						},
					},
				},
			},
		},
	})
}

// getValueType gibt den Typ eines Werts als String zurück
func getValueType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "🔘 Boolean"
	case float64, float32, int, int64, int32:
		return "🔢 Zahl"
	case string:
		return "📝 Text"
	case []interface{}:
		return "📋 Liste"
	case map[string]interface{}:
		return "📦 Objekt"
	default:
		return "❓ Unbekannt"
	}
}
