package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factoriomod"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// HandleSettingsCommand öffnet den interaktiven Settings-Manager
// Nutzung:
//
//	$settings - Zeigt Mod-Liste
//	$settings <modname> - Öffnet spezifische Mod
//	$settings list - Listet alle Mods
func HandleSettingsCommand(s *discordgo.Session, args string) {
	args = strings.TrimSpace(args)

	// Wenn keine Argumente: Zeige Mod-Liste
	if args == "" {
		displayModsForSettings(s, "")
		return
	}

	// Prüfe auf Subcommands
	parts := strings.Fields(args)
	subcommand := strings.ToLower(parts[0])

	switch subcommand {
	case "list":
		// Zeige komplette Mod-Liste
		displayModsForSettings(s, "")

	case "game":
		// Nur Game-Settings anzeigen (wenn Mod angegeben)
		if len(parts) > 1 {
			modName := strings.Join(parts[1:], " ")
			displayModSettingsQuick(s, modName, "game")
		} else {
			support.Send(s, "❌ Nutzung: `"+support.Config.Prefix+"settings game <modname>`")
		}

	case "map":
		// Nur Map-Settings anzeigen (wenn Mod angegeben)
		if len(parts) > 1 {
			modName := strings.Join(parts[1:], " ")
			displayModSettingsQuick(s, modName, "map")
		} else {
			support.Send(s, "❌ Nutzung: `"+support.Config.Prefix+"settings map <modname>`")
		}

	default:
		// Versuche als Modname zu interpretieren
		modName := args
		displayModSettingsQuick(s, modName, "")
	}
}

// displayModsForSettings zeigt die Liste aller Mods für Settings
func displayModsForSettings(s *discordgo.Session, filter string) {
	if factoriomod.GlobalModManager == nil {
		support.Send(s, "❌ Mod-Manager ist nicht initialisiert.")
		return
	}

	// Hole alle Mods aus dem Cache
	allMods := factoriomod.GlobalModManager.GetAllMods()

	if len(allMods) == 0 {
		// Versuch: Mods on-demand laden
		if err := factoriomod.GlobalModManager.DiscoverMods(); err != nil {
			support.Send(s, "❌ Keine Mods gefunden (Discover fehlgeschlagen): "+err.Error())
			return
		}
		allMods = factoriomod.GlobalModManager.GetAllMods()
		if len(allMods) == 0 {
			support.Send(s, "❌ Keine Mods gefunden.")
			return
		}
	}

	// Filtere Mods die Settings haben
	modsWithSettings := make([]models.ModInfo, 0, len(allMods))
	for _, mod := range allMods {
		if mod.HasSettings() {
			modsWithSettings = append(modsWithSettings, mod)
		}
	}

	if len(modsWithSettings) == 0 {
		support.Send(s, "⚠️ Keine Mods mit Einstellungen vorhanden")
		return
	}

	// Erstelle Embed mit Mod-Liste
	embed := &discordgo.MessageEmbed{
		Title:       "⚙️ Server-Einstellungen",
		Description: fmt.Sprintf("Verfügbare Mods zum Bearbeiten (%d)", len(modsWithSettings)),
		Color:       0x5865F2,
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	for i := range modsWithSettings {
		mod := modsWithSettings[i]
		if i >= 25 { // Discord hat Limit von 25 Fields
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("... und %d weitere", len(modsWithSettings)-i),
				Value:  "Nutze `" + support.Config.Prefix + "settings <modname>` für weitere",
				Inline: false,
			})
			break
		}

		gameCount, mapCount := mod.CountSettings()
		value := ""
		if gameCount > 0 {
			value += fmt.Sprintf("🎮 %d Game-Settings", gameCount)
		}
		if mapCount > 0 {
			if value != "" {
				value += " | "
			}
			value += fmt.Sprintf("🗺️ %d Map-Settings", mapCount)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", i+1, mod.Name),
			Value:  value,
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Nutze: " + support.Config.Prefix + "settings <modname> um eine Mod zu öffnen",
	}

	support.SendEmbed(s, embed)
}

// displayModSettingsQuick öffnet Settings für eine spezifische Mod schnell
func displayModSettingsQuick(s *discordgo.Session, modName string, tabType string) {
	if factoriomod.GlobalModManager == nil {
		support.Send(s, "❌ Mod-Manager ist nicht initialisiert.")
		return
	}

	// Finde die Mod
	mod := factoriomod.GlobalModManager.GetModByName(modName)
	if mod == nil {
		support.Send(s, fmt.Sprintf("❌ Mod '%s' nicht gefunden", modName))
		return
	}

	// Wenn kein Tab spezifiziert: Zeige Tab-Auswahl
	if tabType == "" {
		// Simuliere ein Interaction-Event
		// (In real implementation würde dies über einen Button laufen)
		displayModTabSelection(s, mod)
	} else {
		// Zeige direkt die Settings des Tabs
		displayModTabSettings(s, mod, tabType, 0)
	}
}

// displayModTabSelection zeigt die Tab-Auswahl
func displayModTabSelection(s *discordgo.Session, mod *models.ModInfo) {
	gameCount, mapCount := mod.CountSettings()

	// Browser-Style Tab-Visualisierung
	var tabHeader strings.Builder
	tabHeader.WriteString("```\n")
	tabHeader.WriteString("╭───────────────────┬─────────────────╮\n")
	tabHeader.WriteString("│ 🎮 Game Settings  │ 🗺️ Map Settings │\n")
	tabHeader.WriteString("╰───────────────────┴─────────────────╯\n")
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
			Text: fmt.Sprintf("Nutze: %ssettings %s game | %ssettings %s map", support.Config.Prefix, mod.Name, support.Config.Prefix, mod.Name),
		},
	}

	support.SendEmbed(s, embed)
}

// displayModTabSettings zeigt die Settings eines Tabs
func displayModTabSettings(s *discordgo.Session, mod *models.ModInfo, tabType string, page int) {
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

		support.SendEmbed(s, embed)
		return
	}

	// Sortiere Settings alphabetisch
	keys := getSettingKeys(settings)
	totalSettings := len(keys)
	const SETTINGS_PER_PAGE = 8
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

	support.SendEmbed(s, embed)
}

// SettingsDoc ist die Dokumentation für den Settings-Command
var SettingsDoc = support.CommandDoc{
	Name: "settings",
	Usage: "$settings\n" +
		"$settings list\n" +
		"$settings <modname>\n" +
		"$settings game <modname>\n" +
		"$settings map <modname>",
	Doc: "Verwalte interaktiv die Einstellungen aller Server-Mods. " +
		"Mit Game/Map-Flag können nur bestimmte Einstellungstypen angezeigt werden.",
	Subcommands: []support.CommandDoc{
		{
			Name: "list",
			Doc:  "Zeige alle Mods mit Einstellungen",
		},
		{
			Name: "game",
			Doc:  "Zeige nur Game-Settings einer Mod",
		},
		{
			Name: "map",
			Doc:  "Zeige nur Map-Settings einer Mod",
		},
	},
}

func getSettingKeys(settings map[string]interface{}) []string {
	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatSettingValue(value interface{}) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "✅ true"
		}
		return "❌ false"
	case float64:
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
		if actualValue, exists := v["value"]; exists {
			return formatSettingValue(actualValue)
		}
		return "[Object]"
	default:
		return fmt.Sprintf("%v", v)
	}
}
