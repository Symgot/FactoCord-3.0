package discord

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// GhostModeConfig stores individual ghost mode settings per user
type GhostModeConfig struct {
	Enabled            bool      `json:"enabled"`              // Ghost mode active
	PreLogin           bool      `json:"pre_login"`            // Activate on next login
	PreviousSpyMode    []string  `json:"previous_spy_mode"`    // Store previous spy settings
	AutoSilentCommands []string  `json:"auto_silent_commands"` // Commands to auto-silence
	FakeJoinTime       time.Time `json:"fake_join_time"`       // Simulated join time for fake logout
	RealJoinTime       time.Time `json:"real_join_time"`       // Actual join time
}

// GhostModeData stores all ghost mode data
type GhostModeData struct {
	Users map[string]*GhostModeConfig `json:"users"` // Key: Discord User ID
}

var (
	ghostModeData  GhostModeData
	ghostModeMutex sync.RWMutex
	ghostDataPath  = "./ghost_mode.json"
)

// HiddenPlayers tracks players that should be hidden from player lists
var HiddenPlayers = struct {
	sync.RWMutex
	players map[string]bool // Key: Factorio username
}{players: make(map[string]bool)}

// SuppressedLogLines tracks which log patterns to suppress completely
var SuppressedLogPatterns = struct {
	sync.RWMutex
	patterns map[string][]string // Key: Factorio username, Value: patterns to suppress
}{patterns: make(map[string][]string)}

func init() {
	ghostModeData = GhostModeData{
		Users: make(map[string]*GhostModeConfig),
	}
	loadGhostModeData()
}

// loadGhostModeData loads ghost mode data from file
func loadGhostModeData() {
	data, err := os.ReadFile(ghostDataPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error loading ghost mode data: %v\n", err)
		}
		return
	}

	ghostModeMutex.Lock()
	defer ghostModeMutex.Unlock()
	if err := json.Unmarshal(data, &ghostModeData); err != nil {
		fmt.Printf("Error parsing ghost mode data: %v\n", err)
	}
}

// Hilfsfunktion zum Formatieren von Dauer
func formatDuration_ghost(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// saveGhostModeData saves ghost mode data to file
func saveGhostModeData() {
	ghostModeMutex.RLock()
	data, err := json.MarshalIndent(ghostModeData, "", "  ")
	ghostModeMutex.RUnlock()

	if err != nil {
		fmt.Printf("Error marshaling ghost mode data: %v\n", err)
		return
	}

	if err := os.WriteFile(ghostDataPath, data, 0644); err != nil {
		fmt.Printf("Error saving ghost mode data: %v\n", err)
	}
}

// GetGhostConfig returns ghost mode config for a user
func GetGhostConfig(discordUserID string) *GhostModeConfig {
	ghostModeMutex.RLock()
	defer ghostModeMutex.RUnlock()

	if config, exists := ghostModeData.Users[discordUserID]; exists {
		return config
	}
	return nil
}

// IsGhostModeActive checks if ghost mode is active for a Discord user
func IsGhostModeActive(discordUserID string) bool {
	ghostModeMutex.RLock()
	defer ghostModeMutex.RUnlock()

	if config, exists := ghostModeData.Users[discordUserID]; exists {
		return config.Enabled
	}
	return false
}

// IsGhostModePreLogin checks if ghost mode should activate on next login
func IsGhostModePreLogin(discordUserID string) bool {
	ghostModeMutex.RLock()
	defer ghostModeMutex.RUnlock()

	if config, exists := ghostModeData.Users[discordUserID]; exists {
		return config.PreLogin
	}
	return false
}

// IsPlayerHidden checks if a Factorio player should be hidden
func IsPlayerHidden(factorioUsername string) bool {
	HiddenPlayers.RLock()
	defer HiddenPlayers.RUnlock()
	return HiddenPlayers.players[factorioUsername]
}

// SetPlayerHidden sets whether a player should be hidden
func SetPlayerHidden(factorioUsername string, hidden bool) {
	HiddenPlayers.Lock()
	defer HiddenPlayers.Unlock()
	if hidden {
		HiddenPlayers.players[factorioUsername] = true
	} else {
		delete(HiddenPlayers.players, factorioUsername)
	}
}

// ActivateGhostMode - VERBESSERTE VERSION
// Aktiviert den Ghost-Mode für einen Spieler mit besserer Implementierung
func ActivateGhostMode(discordUserID string, factorioUsername string, preLogin bool) {
	ghostModeMutex.Lock()

	config, exists := ghostModeData.Users[discordUserID]
	if !exists {
		config = &GhostModeConfig{
			AutoSilentCommands: []string{},
		}
		ghostModeData.Users[discordUserID] = config
	}

	// Store previous spy mode settings
	config.PreviousSpyMode = GetSuppressedLogs(discordUserID)

	if preLogin {
		// Will activate on next login
		config.PreLogin = true
		config.Enabled = false
		ghostModeMutex.Unlock()
	} else {
		// Activate immediately
		config.Enabled = true
		config.PreLogin = false
		config.RealJoinTime = time.Now()
		config.FakeJoinTime = time.Now()

		ghostModeMutex.Unlock()

		// 1. Entferne Spieler aus sichtbarer Liste
		ActivePlayers.Lock()
		delete(ActivePlayers.players, factorioUsername)
		ActivePlayers.Unlock()

		// 2. Markiere Spieler als versteckt
		SetPlayerHidden(factorioUsername, true)

		// 3. Aktiviere full spy mode (ALLE Logs supprimiert)
		AddSuppressedLog(discordUserID, "ALL")

		// 4. Sende Fake-Logout sofort
		SendPlayerWatcherLeave(factorioUsername, "")

		// 5. Sende Lua-Befehl zu Factorio um Spieler zu verstecken
		if support.Factorio.IsRunning() {
			// Versuche den Spieler in den Zuschauer-Modus zu setzen
			luaCode := fmt.Sprintf(
				`/silent-command if game.players[%q] then game.players[%q].spectator = true; game.players[%q].character = nil end`,
				factorioUsername, factorioUsername, factorioUsername)
			support.Factorio.Send(luaCode)
		}
	}

	saveGhostModeData()
}

// DeactivateGhostMode - VERBESSERTE VERSION
// Deaktiviert Ghost-Mode mit Wiederherstellung
func DeactivateGhostMode(discordUserID string, factorioUsername string) (wasActive bool, fakePlayTime string) {
	ghostModeMutex.Lock()

	config, exists := ghostModeData.Users[discordUserID]
	if !exists || (!config.Enabled && !config.PreLogin) {
		ghostModeMutex.Unlock()
		return false, ""
	}

	wasActive = config.Enabled
	fakePlayTime = ""

	if config.Enabled {
		// Calculate fake play time (from fake join time)
		fakePlayTime = formatDuration_ghost(time.Since(config.FakeJoinTime))
	}

	// Restore previous spy mode settings
	ClearSuppressedLogs(discordUserID)
	for _, logType := range config.PreviousSpyMode {
		AddSuppressedLog(discordUserID, logType)
	}

	// Markiere Spieler als nicht mehr versteckt
	SetPlayerHidden(factorioUsername, false)

	config.Enabled = false
	config.PreLogin = false

	ghostModeMutex.Unlock()
	saveGhostModeData()

	// 1. Sende Lua-Befehl um den Spectator-Modus zu beenden
	if support.Factorio.IsRunning() {
		luaCode := fmt.Sprintf(
			`/silent-command if game.players[%q] then game.players[%q].spectator = false end`,
			factorioUsername, factorioUsername)
		support.Factorio.Send(luaCode)
	}

	// 2. Sende Fake-Login Message
	SendPlayerWatcherJoin(factorioUsername)

	// 3. Füge Spieler wieder zur sichtbaren Liste hinzu
	ActivePlayers.Lock()
	ActivePlayers.players[factorioUsername] = &PlayerInfo{JoinTime: time.Now()}
	ActivePlayers.Unlock()

	return wasActive, fakePlayTime
}

// OnPlayerLogin - VERBESSERTE VERSION
// Handles ghost mode activation when a player logs in
func OnPlayerLogin(factorioUsername string) bool {
	// Check if any user has pre-login ghost mode for this Factorio username
	discordUserID, exists := GetDiscordUserID(factorioUsername)
	if !exists || discordUserID == "" {
		return false
	}

	ghostModeMutex.Lock()
	config, exists := ghostModeData.Users[discordUserID]
	if !exists || !config.PreLogin {
		ghostModeMutex.Unlock()
		return false
	}

	// Activate ghost mode now
	config.Enabled = true
	config.PreLogin = false
	config.RealJoinTime = time.Now()
	config.FakeJoinTime = time.Now()
	ghostModeMutex.Unlock()

	// Mark as hidden
	SetPlayerHidden(factorioUsername, true)

	// Activate full spy mode
	AddSuppressedLog(discordUserID, "ALL")

	saveGhostModeData()

	// Send Lua command to set spectator mode
	if support.Factorio.IsRunning() {
		luaCode := fmt.Sprintf(
			`/silent-command if game.players[%q] then game.players[%q].spectator = true; game.players[%q].character = nil end`,
			factorioUsername, factorioUsername, factorioUsername)
		support.Factorio.Send(luaCode)
	}

	return true // Hide the login
}

// GetAutoSilentCommands returns the auto-silent commands for a user
func GetAutoSilentCommands(discordUserID string) []string {
	ghostModeMutex.RLock()
	defer ghostModeMutex.RUnlock()

	if config, exists := ghostModeData.Users[discordUserID]; exists {
		return config.AutoSilentCommands
	}
	return []string{}
}

// AddAutoSilentCommand adds a command to auto-silence list
func AddAutoSilentCommand(discordUserID string, command string) {
	ghostModeMutex.Lock()

	config, exists := ghostModeData.Users[discordUserID]
	if !exists {
		config = &GhostModeConfig{
			AutoSilentCommands: []string{},
		}
		ghostModeData.Users[discordUserID] = config
	}

	// Check if already exists
	for _, cmd := range config.AutoSilentCommands {
		if cmd == command {
			ghostModeMutex.Unlock()
			return
		}
	}

	config.AutoSilentCommands = append(config.AutoSilentCommands, command)
	ghostModeMutex.Unlock()
	saveGhostModeData()
}

// RemoveAutoSilentCommand removes a command from auto-silence list
func RemoveAutoSilentCommand(discordUserID string, command string) bool {
	ghostModeMutex.Lock()

	config, exists := ghostModeData.Users[discordUserID]
	if !exists {
		ghostModeMutex.Unlock()
		return false
	}

	newList := []string{}
	removed := false
	for _, cmd := range config.AutoSilentCommands {
		if cmd == command {
			removed = true
		} else {
			newList = append(newList, cmd)
		}
	}

	config.AutoSilentCommands = newList
	ghostModeMutex.Unlock()
	saveGhostModeData()
	return removed
}

// ClearAutoSilentCommands clears all auto-silent commands for a user
func ClearAutoSilentCommands(discordUserID string) {
	ghostModeMutex.Lock()

	if config, exists := ghostModeData.Users[discordUserID]; exists {
		config.AutoSilentCommands = []string{}
		ghostModeMutex.Unlock()
		saveGhostModeData()
		return
	}
	ghostModeMutex.Unlock()
}

// IsAutoSilentCommand checks if a command should be auto-silenced
func IsAutoSilentCommand(discordUserID string, command string) bool {
	ghostModeMutex.RLock()
	defer ghostModeMutex.RUnlock()

	config, exists := ghostModeData.Users[discordUserID]
	if !exists {
		return false
	}

	// Normalize command
	command = strings.TrimPrefix(command, "/")
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	commandBase := fields[0]

	for _, cmd := range config.AutoSilentCommands {
		cmd = strings.TrimPrefix(cmd, "/")
		if cmd == commandBase || cmd == command {
			return true
		}
	}
	return false
}

// ShouldSuppressLog checks if a log line should be suppressed for any hidden player
func ShouldSuppressLog(line string) bool {
	HiddenPlayers.RLock()
	defer HiddenPlayers.RUnlock()

	for playerName := range HiddenPlayers.players {
		// Check if this log line contains the hidden player's name
		if strings.Contains(line, playerName) {
			return true
		}
	}
	return false
}

// handleGhostCommand handles the !ghost command
func handleGhostCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)
	if !isVerified {
		sendDMToChannel(s, m.ChannelID, "❌ You must be verified to use ghost mode. Use `"+support.Config.Prefix+"verify <your_factorio_username>` first.")
		return
	}

	args = strings.TrimSpace(args)
	parts := strings.Fields(args)

	if len(parts) == 0 {
		// Show current ghost mode status
		config := GetGhostConfig(m.Author.ID)
		statusEmoji := "❌"
		statusText := "Inactive"
		preLoginText := ""

		if config != nil {
			if config.Enabled {
				statusEmoji = "👻"
				statusText = "**ACTIVE** - You are invisible"
			}
			if config.PreLogin {
				preLoginText = "\n⏳ **Pre-Login Mode:** Will activate on your next server join"
			}
		}

		autoCommands := GetAutoSilentCommands(m.Author.ID)
		autoCommandsText := "None"
		if len(autoCommands) > 0 {
			autoCommandsText = "`" + strings.Join(autoCommands, "`, `") + "`"
		}

		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("**👻 Ghost Mode**\n\n"+
			"**Status:** %s %s%s\n\n"+
			"**Auto-Silent Commands:** %s\n\n"+
			"**Usage:**\n"+
			"• `%sghost on` - Activate ghost mode (become invisible)\n"+
			"• `%sghost off` - Deactivate ghost mode (become visible)\n"+
			"• `%sghost prelogin` - Activate on next server join\n"+
			"• `%sghost commands` - Manage auto-silent commands\n\n"+
			"**What Ghost Mode does:**\n"+
			"• 🔇 Suppresses ALL your logs (chat, actions, etc.)\n"+
			"• 👤 Hides you from player lists\n"+
			"• 🚪 Hides your join/leave messages\n"+
			"• ⏱️ Shows fake play time on deactivation",
			statusEmoji, statusText, preLoginText, autoCommandsText,
			support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix))
		return
	}

	action := strings.ToLower(parts[0])

	switch action {
	case "on":
		if IsGhostModeActive(m.Author.ID) {
			sendDMToChannel(s, m.ChannelID, "👻 Ghost mode is already active.")
			return
		}

		// Check if player is online
		ActivePlayers.RLock()
		_, isOnline := ActivePlayers.players[factorioUsername]
		ActivePlayers.RUnlock()

		if !isOnline {
			sendDMToChannel(s, m.ChannelID, "❌ You must be online in the game to activate ghost mode.\n\nUse `"+support.Config.Prefix+"ghost prelogin` to activate on your next join.")
			return
		}

		// Send fake logout message before going invisible via PlayerWatcher
		ActivePlayers.RLock()
		info := ActivePlayers.players[factorioUsername]
		playTime := ""
		if info != nil {
			playTime = formatDuration_ghost(time.Since(info.JoinTime))
		}
		ActivePlayers.RUnlock()

		// Send fake logout via PlayerWatcher
		SendPlayerWatcherLeave(factorioUsername, playTime)

		// Remove from visible active players (but keep tracking internally)
		ActivePlayers.Lock()
		delete(ActivePlayers.players, factorioUsername)
		ActivePlayers.Unlock()

		ActivateGhostMode(m.Author.ID, factorioUsername, false)

		sendDMToChannel(s, m.ChannelID, "👻 **Ghost Mode ACTIVATED**\n\n"+
			"• You are now invisible to other players\n"+
			"• A fake logout message was sent\n"+
			"• All your actions are hidden from logs\n"+
			"• You won't appear in player lists\n\n"+
			"Use `"+support.Config.Prefix+"ghost off` to become visible again.")

	case "off":
		wasActive, fakePlayTime := DeactivateGhostMode(m.Author.ID, factorioUsername)
		if !wasActive {
			// Check if pre-login was set and cancel it safely
			ghostModeMutex.Lock()
			config := ghostModeData.Users[m.Author.ID]
			wasPreLogin := config != nil && config.PreLogin
			if wasPreLogin {
				config.PreLogin = false
			}
			ghostModeMutex.Unlock()
			if wasPreLogin {
				saveGhostModeData()
				sendDMToChannel(s, m.ChannelID, "✅ Pre-login ghost mode cancelled.")
				return
			}

			sendDMToChannel(s, m.ChannelID, "❌ Ghost mode is not active.")
			return
		}

		// Send fake login message via PlayerWatcher
		SendPlayerWatcherJoin(factorioUsername)

		// Add back to visible active players
		ActivePlayers.Lock()
		ActivePlayers.players[factorioUsername] = &PlayerInfo{JoinTime: time.Now()}
		ActivePlayers.Unlock()

		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ **Ghost Mode DEACTIVATED**\n\n"+
			"• You are visible again\n"+
			"• A fake login message was sent\n"+
			"• Your previous spy mode settings have been restored\n"+
			"• Ghost session duration: %s", fakePlayTime))

	case "prelogin", "pre-login", "pre":
		if IsGhostModeActive(m.Author.ID) {
			sendDMToChannel(s, m.ChannelID, "👻 Ghost mode is already active. Use `"+support.Config.Prefix+"ghost off` to deactivate first.")
			return
		}

		ActivateGhostMode(m.Author.ID, factorioUsername, true)
		sendDMToChannel(s, m.ChannelID, "⏳ **Pre-Login Ghost Mode SET**\n\n"+
			"Ghost mode will automatically activate when you next join the server.\n"+
			"• Your login will be hidden\n"+
			"• You'll be invisible from the start\n\n"+
			"Use `"+support.Config.Prefix+"ghost off` to cancel.")

	case "commands", "cmd", "cmds":
		handleGhostCommandsMenu(s, m, parts[1:])

	default:
		sendDMToChannel(s, m.ChannelID, "❌ Unknown ghost command. Use `"+support.Config.Prefix+"ghost` for help.")
	}
}

// handleGhostCommandsMenu handles the auto-silent commands submenu
func handleGhostCommandsMenu(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		// Show current auto-silent commands
		commands := GetAutoSilentCommands(m.Author.ID)

		var commandsList string
		if len(commands) == 0 {
			commandsList = "No commands configured"
		} else {
			for i, cmd := range commands {
				commandsList += fmt.Sprintf("%d. `%s`\n", i+1, cmd)
			}
		}

		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("**⚡ Auto-Silent Commands**\n\n"+
			"These commands will automatically be executed as silent commands:\n\n"+
			"%s\n"+
			"**Usage:**\n"+
			"• `%sghost commands add <command>` - Add a command\n"+
			"• `%sghost commands remove <number>` - Remove by number\n"+
			"• `%sghost commands clear` - Clear all commands\n\n"+
			"**Example:** `%sghost commands add c` - Auto-silence the `/c` command",
			commandsList, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix))
		return
	}

	action := strings.ToLower(args[0])

	switch action {
	case "add":
		if len(args) < 2 {
			sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"ghost commands add <command>`")
			return
		}
		command := strings.Join(args[1:], " ")
		AddAutoSilentCommand(m.Author.ID, command)
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Added auto-silent command: `%s`", command))

	case "remove", "rm", "del":
		if len(args) < 2 {
			sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"ghost commands remove <number>`")
			return
		}
		commands := GetAutoSilentCommands(m.Author.ID)
		var index int
		_, err := fmt.Sscanf(args[1], "%d", &index)
		if err != nil || index < 1 || index > len(commands) {
			sendDMToChannel(s, m.ChannelID, "❌ Invalid command number.")
			return
		}
		command := commands[index-1]
		RemoveAutoSilentCommand(m.Author.ID, command)
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Removed auto-silent command: `%s`", command))

	case "clear":
		ClearAutoSilentCommands(m.Author.ID)
		sendDMToChannel(s, m.ChannelID, "✅ Cleared all auto-silent commands.")

	default:
		sendDMToChannel(s, m.ChannelID, "❌ Unknown action. Use `add`, `remove`, or `clear`.")
	}
}

// GetVisiblePlayerCount returns the count of visible players (excluding hidden ones)
func GetVisiblePlayerCount() int {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()

	count := 0
	for name := range ActivePlayers.players {
		if !IsPlayerHidden(name) {
			count++
		}
	}
	return count
}

// GetVisiblePlayers returns only visible players
func GetVisiblePlayers() map[string]time.Time {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()

	result := make(map[string]time.Time)
	for name, info := range ActivePlayers.players {
		if !IsPlayerHidden(name) {
			result[name] = info.JoinTime
		}
	}
	return result
}
