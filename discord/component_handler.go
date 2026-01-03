package discord

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/factorio"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

// HandleComponentInteraction verarbeitet Button- und andere Interaktionen
func HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID

	// Hole UserID - in DMs ist Member nil, User ist direkt verfügbar
	var userID string
	if i.User != nil {
		userID = i.User.ID
	} else if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else {
		return
	}

	// Verifikation prüfen
	if !IsUserVerified(userID) {
		respondWithMessage(s, i.Interaction, "❌ Du bist nicht verifiziert.")
		return
	}

	// Route basierend auf CustomID
	switch {
	case customID == "mod_select_menu":
		handleModSelectMenu(s, i)

	case strings.HasPrefix(customID, "mod_select_"):
		handleModSelection(s, i, customID)

	case strings.HasPrefix(customID, "tab_"):
		handleTabSwitch(s, i, customID)

	case strings.HasPrefix(customID, "page_"):
		handlePageNavigation(s, i, customID)

	case strings.HasPrefix(customID, "modlist_page_"):
		handleModlistPage(s, i, customID)

	case customID == "modlist_refresh":
		handleModlistRefresh(s, i)

	case customID == "modlist_help":
		handleModlistHelp(s, i)

	case customID == "back_to_mods":
		handleBackToModList(s, i)

	case customID == "back_to_tabs":
		handleBackToTabs(s, i)

	case strings.HasPrefix(customID, "back_to_mod_"):
		handleBackToMod(s, i, customID)

	case strings.HasPrefix(customID, "edit_"):
		handleEditRequest(s, i, customID)

	case customID == "open_save_dialog":
		handleOpenSaveDialog(s, i)

	case customID == "confirm_save":
		handleSaveConfirmation(s, i)

	case customID == "confirm_cancel":
		handleCancelConfirmation(s, i)

	case strings.HasPrefix(customID, "setting_select_"):
		handleSettingInteraction(s, i, customID)

	case strings.HasPrefix(customID, "setting_"):
		handleSettingInteraction(s, i, customID)
	}
}

// handleModSelection verarbeitet die Auswahl eines Mods
func handleModSelection(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	modName := strings.TrimPrefix(customID, "mod_select_")
	userID := getUserID(i)

	// Acknowledge die Interaktion
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	// Lade Mod
	mod := factorio.GlobalModManager.GetModByName(modName)
	if mod == nil {
		updateMessage(s, i.Interaction, "❌ Mod nicht gefunden: "+modName)
		return
	}

	// Erstelle/aktualisiere Session
	globalSessionManager.CreateSession(userID, modName)

	// Zeige Tab-Auswahl
	displayTabSelection(s, i.Interaction, mod, "")
}

// handleTabSwitch verarbeitet Tab-Wechsel
func handleTabSwitch(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Format: tab_<type>_<modname>
	parts := strings.SplitN(customID, "_", 3)
	if len(parts) < 3 {
		return
	}

	tabType := parts[1]
	modName := parts[2]

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mod := factorio.GlobalModManager.GetModByName(modName)
	if mod == nil {
		updateMessage(s, i.Interaction, "❌ Mod nicht gefunden: "+modName)
		return
	}

	displaySettingsTab(s, i.Interaction, mod, tabType, 0)
}

// handlePageNavigation verarbeitet Seiten-Navigation
func handlePageNavigation(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Neues Format: page_<modname>_<tabtype>_<targetpage>
	parts := strings.Split(customID, "_")
	if len(parts) < 4 {
		return
	}

	modName := parts[1]
	tabType := parts[2]
	targetPage, _ := strconv.Atoi(parts[3])

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mod := factorio.GlobalModManager.GetModByName(modName)
	if mod == nil {
		return
	}

	displaySettingsTab(s, i.Interaction, mod, tabType, targetPage)
}

// handleBackToModList kehrt zur Mod-Liste zurück
func handleBackToModList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mods := factorio.GlobalModManager.GetModsWithSettings()
	displayModListUpdate(s, i.Interaction, mods)
}

// handleBackToTabs kehrt zur Tab-Auswahl zurück
func handleBackToTabs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := getUserID(i)
	session := globalSessionManager.GetSession(userID)
	if session == nil {
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mod := factorio.GlobalModManager.GetModByName(session.ModName)
	if mod == nil {
		return
	}

	displayTabSelection(s, i.Interaction, mod, "")
}

// handleEditRequest verarbeitet Bearbeitungsanfragen
func handleEditRequest(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Format: edit_<modname>_<tabtype>_<page>
	parts := strings.Split(customID, "_")
	if len(parts) < 4 {
		return
	}

	modName := parts[1]
	tabType := parts[2]
	page, _ := strconv.Atoi(parts[3])

	mod := factorio.GlobalModManager.GetModByName(modName)
	if mod == nil {
		return
	}

	// Zeige Modal zur Bearbeitung
	showEditModal(s, i, mod, tabType, page)
}

// handleSettingInteraction verarbeitet Setting-spezifische Interaktionen
func handleSettingInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Neues Format: setting_select_<modname>_<tabtype> (vom Select-Menü)
	// Wert ist: edit_<modname>_<tabtype>_<settingkey>

	// Prüfe ob es ein Select-Menü ist
	if strings.HasPrefix(customID, "setting_select_") {
		values := i.MessageComponentData().Values
		if len(values) == 0 {
			return
		}

		// Parse den ausgewählten Wert
		selectedValue := values[0]
		if strings.HasPrefix(selectedValue, "edit_") {
			parts := strings.SplitN(selectedValue, "_", 4)
			if len(parts) >= 4 {
				modName := parts[1]
				tabType := parts[2]
				settingKey := parts[3]

				mod := factorio.GlobalModManager.GetModByName(modName)
				if mod == nil {
					return
				}

				// Hole aktuellen Wert
				var currentValue interface{}
				if tabType == "game" {
					currentValue = mod.GetGameSetting(settingKey)
				} else {
					currentValue = mod.GetMapSetting(settingKey)
				}

				// Zeige Edit-Modal
				displayEditSettingModal(s, i.Interaction, modName, tabType, settingKey, currentValue)
			}
		}
	}
}

// handleModSelectMenu verarbeitet Auswahlen aus dem Mod-Select-Menü
func handleModSelectMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) == 0 {
		return
	}

	selectedValue := values[0]
	if strings.HasPrefix(selectedValue, "mod_select_") {
		modName := strings.TrimPrefix(selectedValue, "mod_select_")
		userID := getUserID(i)

		// Acknowledge die Interaktion
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		mod := factorio.GlobalModManager.GetModByName(modName)
		if mod == nil {
			updateMessage(s, i.Interaction, "❌ Mod nicht gefunden: "+modName)
			return
		}

		globalSessionManager.CreateSession(userID, modName)
		displayTabSelection(s, i.Interaction, mod, "")
	}
}

// handleBackToMod kehrt zur Tab-Übersicht eines Mods zurück
func handleBackToMod(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	modName := strings.TrimPrefix(customID, "back_to_mod_")

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mod := factorio.GlobalModManager.GetModByName(modName)
	if mod == nil {
		return
	}

	displayTabSelection(s, i.Interaction, mod, "")
}

// handleOpenSaveDialog öffnet den Speicher-Dialog
func handleOpenSaveDialog(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := getUserID(i)
	session := globalSessionManager.GetSession(userID)
	if session == nil || len(session.Changes) == 0 {
		respondWithMessage(s, i.Interaction, "⚠️ Keine Änderungen zum Speichern vorhanden.")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	DisplayChangesPreview(s, i.Interaction, session)
}

// handleModlistPage verarbeitet Pagination der Mod-Liste
func handleModlistPage(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Format: modlist_page_<pagenum>
	parts := strings.Split(customID, "_")
	if len(parts) < 3 {
		return
	}

	page, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	mods := factorio.GlobalModManager.GetModsWithSettings()
	displayModListPage(s, "", i.Interaction, mods, page)
}

// handleModlistRefresh aktualisiert die Mod-Liste
func handleModlistRefresh(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	// Lade Mods neu
	if err := factorio.GlobalModManager.DiscoverMods(); err != nil {
		updateMessage(s, i.Interaction, "❌ Fehler beim Laden der Mods: "+err.Error())
		return
	}

	mods := factorio.GlobalModManager.GetModsWithSettings()
	displayModListPage(s, "", i.Interaction, mods, 0)
}

// handleModlistHelp zeigt Hilfe an
func handleModlistHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title: "❓ Hilfe - Mod-Settings Manager",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "🔧 Mod auswählen",
				Value: "Wähle einen Mod aus dem Dropdown-Menü um dessen Einstellungen zu bearbeiten.",
			},
			{
				Name:  "🎮 Game Settings",
				Value: "Einstellungen die beim Spielstart geladen werden (Startup & Runtime-Global).",
			},
			{
				Name:  "🗺️ Map Settings",
				Value: "Einstellungen die pro Spieler gespeichert werden (Runtime-Per-User).",
			},
			{
				Name:  "💾 Speichern",
				Value: "Nach dem Bearbeiten werden Änderungen in einer Vorschau angezeigt. Bestätige zum Speichern.",
			},
			{
				Name:  "⚠️ Server-Neustart",
				Value: "Für Startup-Settings ist ein Server-Neustart erforderlich!",
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// showEditModal zeigt ein Modal zur Bearbeitung von Settings
func showEditModal(s *discordgo.Session, i *discordgo.InteractionCreate, mod *models.ModInfo, tabType string, page int) {
	// Hole Settings für die aktuelle Seite
	var settings map[string]interface{}
	if tabType == "game" {
		settings = mod.GetAllGameSettings()
	} else {
		settings = mod.GetAllMapSettings()
	}

	keys := getSettingKeys(settings)
	start := page * SETTINGS_PER_PAGE
	end := start + SETTINGS_PER_PAGE
	if end > len(keys) {
		end = len(keys)
	}

	// Erstelle Modal-Komponenten (max 5 Eingabefelder pro Modal)
	components := []discordgo.MessageComponent{}
	for _, key := range keys[start:end] {
		value := settings[key]
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    fmt.Sprintf("input_%s", key),
					Label:       truncateString(key, 45),
					Style:       discordgo.TextInputShort,
					Placeholder: fmt.Sprintf("Aktuell: %v", value),
					Required:    false,
					MaxLength:   1000,
				},
			},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   fmt.Sprintf("modal_%s_%s_%d", mod.Name, tabType, page),
			Title:      fmt.Sprintf("Bearbeite %s Settings", tabType),
			Components: components,
		},
	})
}

// Helper-Funktionen

func getUserID(i *discordgo.InteractionCreate) string {
	if i.User != nil {
		return i.User.ID
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	return ""
}

func respondWithMessage(s *discordgo.Session, i *discordgo.Interaction, message string) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func updateMessage(s *discordgo.Session, i *discordgo.Interaction, message string) {
	content := message
	s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Content: &content,
	})
}
