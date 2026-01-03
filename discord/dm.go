package discord

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/commands"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// DMSession holds information about a DM session with an admin
type DMSession struct {
	UserID   string
	Username string
}

// SpyModeData tracks which log types are suppressed for each player
type SpyModeData struct {
	SuppressedLogs map[string][]string `json:"suppressed_logs"`
}

var (
	spyModeData  SpyModeData
	spyModeMutex sync.RWMutex
)

func init() {
	spyModeData = SpyModeData{
		SuppressedLogs: make(map[string][]string),
	}
}

// Available log types that can be suppressed
// These must match the checks in control.lua
var availableLogTypes = []string{
	"ALL",          // Suppress all logs
	"JOIN",         // Player joined server
	"LEAVE",        // Player left server
	"CHAT",         // Chat messages
	"DIED",         // Player death
	"KICKED",       // Player was kicked
	"MUTED",        // Player mute/unmute
	"BUILT_ENTITY", // Entity building (if logging enabled in control.lua)
	"MINED_ENTITY", // Entity mining (if logging enabled in control.lua)
	"RESEARCH",     // Research completion (if logging enabled in control.lua)
	"COMMAND",      // Console commands (if logging enabled in control.lua)
	"ENTITY_DESTROYED",
	"CHUNK_GENERATED",
	"SECTOR_SCANNED",
	"ALL",
}

// HandleDMMessage processes messages sent in DMs
func HandleDMMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// Check if DM chat is enabled
	if !support.Config.EnableDMChat {
		return false
	}

	// Check if this is a DM channel
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return false
	}
	if channel.Type != discordgo.ChannelTypeDM {
		return false
	}

	// Don't respond to bot's own messages
	if m.Author.ID == s.State.User.ID {
		return true
	}

	// Check if user is an admin
	isAdmin := commands.CheckAdmin(m.Author.ID)

	// Handle the message
	content := strings.TrimSpace(m.Content)

	// Check for command prefix
	if strings.HasPrefix(content, support.Config.Prefix) {
		input := strings.TrimPrefix(content, support.Config.Prefix)
		handleDMCommand(s, m, input, isAdmin)
		return true
	}

	// If not a command and user is admin, show help
	if isAdmin {
		sendDMHelp(s, m.ChannelID)
	} else {
		sendDMToChannel(s, m.ChannelID, "❌ You are not authorized to use DM commands. Only admins listed in the config can use this feature.")
	}

	return true
}

// handleDMCommand processes commands sent in DMs
func handleDMCommand(s *discordgo.Session, m *discordgo.MessageCreate, input string, isAdmin bool) {
	inputvars := strings.SplitN(input+" ", " ", 2)
	commandName := strings.ToLower(inputvars[0])
	args := strings.TrimSpace(inputvars[1])

	// DM-exclusive commands
	switch commandName {
	case "verify":
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use the verify command.")
			return
		}
		handleVerifyCommand(s, m, args)
		return

	case "confirm":
		handleConfirmCommand(s, m, args)
		return

	case "unlink":
		handleUnlinkCommand(s, m)
		return

	case "status":
		handleStatusCommand(s, m)
		return

	case "spy":
		// Spy mode - suppress specific logs
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use spy mode.")
			return
		}
		handleSpyCommand(s, m, args)
		return

	case "ghost":
		// Ghost mode - full invisibility
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use ghost mode.")
			return
		}
		handleGhostCommand(s, m, args)
		return

	case "c":
		// Open command (visible in chat and logs)
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use console commands.")
			return
		}
		// Check for auto-silent commands
		if IsAutoSilentCommand(m.Author.ID, args) {
			handleSilentCommand(s, m, args)
			return
		}
		handleOpenCommand(s, m, args)
		return

	case "sc":
		// Silent/hidden command (NOT visible in chat or logs)
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use silent commands.")
			return
		}
		handleSilentCommand(s, m, args)
		return

	case "help":
		sendDMHelp(s, m.ChannelID)
		return

	default:
		// Try to run as a regular command if admin
		if isAdmin {
			// Create a fake message for the command handler
			runDMRegularCommand(s, m, commandName, args)
		} else {
			sendDMToChannel(s, m.ChannelID, "❌ Unknown command. Use `"+support.Config.Prefix+"help` for available commands.")
		}
	}
}

// handleVerifyCommand initiates verification with a Factorio player
func handleVerifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	factorioUsername := strings.TrimSpace(args)
	if factorioUsername == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"verify <factorio_username>`\n\nThis will send a 6-digit verification code to the player in Factorio via a private whisper.")
		return
	}

	// Check if server is running
	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Generate verification code
	code, err := CreateVerificationRequest(m.Author.ID, factorioUsername)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ Failed to create verification request: "+err.Error())
		return
	}

	// Send the verification code to the Factorio player via silent whisper
	// Using RCON-style command that doesn't appear in logs
	// The /silent-command runs Lua code without logging to the console
	luaCommand := fmt.Sprintf(`/silent-command local p = game.get_player("%s"); if p and p.connected then p.print("[FactoCord Verification] Your code: %s - Enter this in Discord DM with: %sverify %s", {color={r=0.2,g=0.9,b=0.2}}) end`,
		escapeFactorioString(factorioUsername), code, support.Config.Prefix, code)

	support.Factorio.Send(luaCommand)

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Verification code sent to **%s** in Factorio!\n\nThe player should see a private message with a 6-digit code.\nOnce they tell you the code, use: `%sconfirm <code>`\n\n⏰ The code expires in 5 minutes.", factorioUsername, support.Config.Prefix))
}

// handleConfirmCommand confirms a verification code
func handleConfirmCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	code := strings.TrimSpace(args)
	if code == "" || len(code) != 6 {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"confirm <6-digit-code>`")
		return
	}

	success, factorioUsername, err := VerifyCode(m.Author.ID, code)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ Verification failed: "+err.Error())
		return
	}

	if !success {
		sendDMToChannel(s, m.ChannelID, "❌ Invalid verification code. Please try again.")
		return
	}

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Successfully verified! Your Discord account is now linked to Factorio player **%s**.\n\nYou can now use:\n• `%sc <command>` - Run commands (visible in game)\n• `%ssc <command>` - Run silent commands (hidden from logs)\n• `%ssc PlayerName.editor` - Run commands on specific players\n• `%sspy` - Manage spy mode (log suppression)", factorioUsername, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix))

	log.Printf("[DM] User %s (%s) verified as Factorio player %s", m.Author.Username, m.Author.ID, factorioUsername)
}

// handleUnlinkCommand removes the verification link
func handleUnlinkCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := UnlinkUser(m.Author.ID)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ "+err.Error())
		return
	}

	sendDMToChannel(s, m.ChannelID, "✅ Your Discord account has been unlinked from your Factorio account.")
}

// handleStatusCommand shows verification status
func handleStatusCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)

	isAdmin := commands.CheckAdmin(m.Author.ID)
	adminStatus := "❌ Not an admin"
	if isAdmin {
		adminStatus = "✅ Admin"
	}

	verifyStatus := "❌ Not verified"
	if isVerified {
		verifyStatus = fmt.Sprintf("✅ Verified as **%s**", factorioUsername)
	}

	serverStatus := "❌ Server offline"
	if support.Factorio.IsRunning() {
		serverStatus = "✅ Server online"
	}

	// Show spy mode status
	spyStatus := "❌ Inactive"
	suppressedLogs := GetSuppressedLogs(m.Author.ID)
	if len(suppressedLogs) > 0 {
		spyStatus = fmt.Sprintf("✅ Active - Suppressing: %s", strings.Join(suppressedLogs, ", "))
	}

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("**Your Status:**\n• Admin: %s\n• Verification: %s\n• Server: %s\n• Spy Mode: %s", adminStatus, verifyStatus, serverStatus, spyStatus))
}

// handleOpenCommand runs a Factorio command (visible in chat and logs)
func handleOpenCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	if args == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"c <factorio_command>`\n\nThis command is visible in the Factorio chat and logs.")
		return
	}

	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Check if user is verified (required for running as a specific player)
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)

	// Ensure command starts with /
	command := args
	if !strings.HasPrefix(command, "/") {
		command = "/" + command
	}

	// Run the command
	support.Factorio.Send(command)

	if isVerified {
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Command executed as **%s**: `%s`\n\n⚠️ This command was visible in the game chat and logs.", factorioUsername, command))
	} else {
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Command executed: `%s`\n\n⚠️ This command was visible in the game chat and logs.", command))
	}

	log.Printf("[DM] Open command from %s: %s", m.Author.Username, command)
}

// handleSpyCommand manages spy mode (log suppression)
func handleSpyCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)
	if !isVerified {
		sendDMToChannel(s, m.ChannelID, "❌ You must be verified to use spy mode. Use `"+support.Config.Prefix+"verify <your_factorio_username>` first.")
		return
	}

	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	args = strings.TrimSpace(args)
	parts := strings.Fields(args)

	if len(parts) == 0 {
		// Show current spy mode status and available options
		suppressedLogs := GetSuppressedLogs(m.Author.ID)
		currentStatus := "None"
		if len(suppressedLogs) > 0 {
			currentStatus = strings.Join(suppressedLogs, ", ")
		}

		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("**🕵️ Spy Mode - Log Suppression**\n\n"+
			"**Current suppressed logs:** %s\n\n"+
			"**Usage:**\n"+
			"• `%sspy add <LOG_TYPE>` - Suppress a log type\n"+
			"• `%sspy remove <LOG_TYPE>` - Stop suppressing a log type\n"+
			"• `%sspy clear` - Clear all suppressions\n"+
			"• `%sspy list` - Show available log types\n"+
			"• `%sspy on` - Enable spy mode (suppress ALL logs)\n"+
			"• `%sspy off` - Disable spy mode completely\n\n"+
			"**Example:** `%sspy add BUILT_ENTITY` - Hide building actions from logs",
			currentStatus,
			support.Config.Prefix, support.Config.Prefix, support.Config.Prefix,
			support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix))
		return
	}

	action := strings.ToLower(parts[0])

	switch action {
	case "list":
		sendDMToChannel(s, m.ChannelID, "**Available log types to suppress:**\n```\n"+strings.Join(availableLogTypes, "\n")+"```\n\nUse `ALL` to suppress all log types at once.")

	case "add":
		if len(parts) < 2 {
			sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"spy add <LOG_TYPE>`")
			return
		}
		logType := strings.ToUpper(parts[1])
		if !isValidLogType(logType) {
			sendDMToChannel(s, m.ChannelID, "❌ Invalid log type. Use `"+support.Config.Prefix+"spy list` to see available types.")
			return
		}
		AddSuppressedLog(m.Author.ID, logType)
		activateSpyMode(factorioUsername, m.Author.ID)
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Now suppressing **%s** logs for your player.", logType))

	case "remove":
		if len(parts) < 2 {
			sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"spy remove <LOG_TYPE>`")
			return
		}
		logType := strings.ToUpper(parts[1])
		RemoveSuppressedLog(m.Author.ID, logType)
		activateSpyMode(factorioUsername, m.Author.ID)
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Stopped suppressing **%s** logs.", logType))

	case "clear", "off":
		ClearSuppressedLogs(m.Author.ID)
		deactivateSpyMode(factorioUsername)
		sendDMToChannel(s, m.ChannelID, "✅ Spy mode deactivated. All logs will be visible again.")

	case "on":
		AddSuppressedLog(m.Author.ID, "ALL")
		activateSpyMode(factorioUsername, m.Author.ID)
		sendDMToChannel(s, m.ChannelID, "✅ Spy mode **FULLY ACTIVATED**. All your actions will be hidden from logs.\n\n⚠️ Use `"+support.Config.Prefix+"spy off` to deactivate.")

	default:
		sendDMToChannel(s, m.ChannelID, "❌ Unknown spy command. Use `"+support.Config.Prefix+"spy` for help.")
	}
}

// activateSpyMode sends Lua commands to suppress logs for a player
func activateSpyMode(factorioUsername string, discordUserID string) {
	suppressedLogs := GetSuppressedLogs(discordUserID)
	if len(suppressedLogs) == 0 {
		return
	}

	// Create a Lua script that sets up log suppression for this player
	// This uses global storage to track which players have spy mode enabled
	luaCode := fmt.Sprintf(`if not storage.factocord_spy_mode then storage.factocord_spy_mode = {} end; storage.factocord_spy_mode["%s"] = {%s}`,
		escapeFactorioString(factorioUsername), formatLuaStringArray(suppressedLogs))

	support.Factorio.Send("/silent-command " + luaCode)
}

// deactivateSpyMode removes spy mode for a player
func deactivateSpyMode(factorioUsername string) {
	luaCode := fmt.Sprintf(`if storage.factocord_spy_mode then storage.factocord_spy_mode["%s"] = nil end`,
		escapeFactorioString(factorioUsername))

	support.Factorio.Send("/silent-command " + luaCode)
}

// formatLuaStringArray formats a string array for Lua
func formatLuaStringArray(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	quoted := make([]string, len(arr))
	for i, s := range arr {
		quoted[i] = fmt.Sprintf(`"%s"`, s)
	}
	return strings.Join(quoted, ", ")
}

// isValidLogType checks if a log type is valid
func isValidLogType(logType string) bool {
	for _, lt := range availableLogTypes {
		if lt == logType {
			return true
		}
	}
	return false
}

// GetSuppressedLogs returns the suppressed log types for a user
func GetSuppressedLogs(discordUserID string) []string {
	spyModeMutex.RLock()
	defer spyModeMutex.RUnlock()
	return spyModeData.SuppressedLogs[discordUserID]
}

// AddSuppressedLog adds a log type to the suppression list
func AddSuppressedLog(discordUserID, logType string) {
	spyModeMutex.Lock()
	defer spyModeMutex.Unlock()

	logs := spyModeData.SuppressedLogs[discordUserID]
	for _, l := range logs {
		if l == logType {
			return // Already exists
		}
	}
	spyModeData.SuppressedLogs[discordUserID] = append(logs, logType)
}

// RemoveSuppressedLog removes a log type from the suppression list
func RemoveSuppressedLog(discordUserID, logType string) {
	spyModeMutex.Lock()
	defer spyModeMutex.Unlock()

	logs := spyModeData.SuppressedLogs[discordUserID]
	newLogs := make([]string, 0)
	for _, l := range logs {
		if l != logType {
			newLogs = append(newLogs, l)
		}
	}
	spyModeData.SuppressedLogs[discordUserID] = newLogs
}

// ClearSuppressedLogs clears all suppressed logs for a user
func ClearSuppressedLogs(discordUserID string) {
	spyModeMutex.Lock()
	defer spyModeMutex.Unlock()
	delete(spyModeData.SuppressedLogs, discordUserID)
}

// parsePlayerCommand parses commands in format "player.command" or just "command"
// Returns: targetPlayer, command, usesPlayerPrefix
func parsePlayerCommand(args string, defaultPlayer string) (string, string, bool) {
	// Check if format is "player.command" (e.g., "Shadow.editor")
	if strings.Contains(args, ".") {
		parts := strings.SplitN(args, ".", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			// Check if first part looks like a player name (not a Lua object like "game")
			firstPart := strings.ToLower(parts[0])
			luaObjects := []string{"game", "script", "remote", "commands", "settings", "rcon", "rendering", "global", "storage", "prototypes", "helpers"}
			for _, obj := range luaObjects {
				if firstPart == obj {
					return defaultPlayer, args, false
				}
			}
			return parts[0], parts[1], true
		}
	}
	return defaultPlayer, args, false
}

// handleSilentCommand runs a Factorio command silently (NOT visible in chat or logs)
func handleSilentCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	if args == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"sc <command>` or `"+support.Config.Prefix+"sc <player>.<command>`\n\n"+
			"**Quick Commands:**\n"+
			"• `editor` - Toggle map editor\n"+
			"• `cheat` - Toggle cheat mode\n"+
			"• `spectator` - Toggle spectator mode\n"+
			"• `god` - Toggle god mode (destroy character)\n"+
			"• `zoom <value>` - Set zoom level\n"+
			"• `speed <value>` - Set game speed\n"+
			"• `teleport <x> <y>` - Teleport to position\n"+
			"• `give <item> [count]` - Give items\n"+
			"• `research_all` - Research all technologies\n"+
			"• `chart_all` - Reveal entire map\n\n"+
			"**Target another player:**\n"+
			"• `"+support.Config.Prefix+"sc PlayerName.editor` - Toggle editor for PlayerName\n"+
			"• `"+support.Config.Prefix+"sc PlayerName.give iron-plate 1000`\n\n"+
			"**Raw Lua:**\n"+
			"• `"+support.Config.Prefix+"sc game.speed = 2`")
		return
	}

	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Get verified username as default target
	defaultPlayer, isVerified := GetFactorioUsername(m.Author.ID)

	// Parse player.command format
	targetPlayer, command, usesPlayerPrefix := parsePlayerCommand(args, defaultPlayer)

	// If using player prefix but not verified, still allow if they specified a player
	if usesPlayerPrefix && targetPlayer != "" {
		// Using explicit player target, allow even if not verified
	} else if !isVerified && needsPlayerContext(command) {
		sendDMToChannel(s, m.ChannelID, "❌ You must be verified to use player-specific commands. Use `"+support.Config.Prefix+"verify <your_factorio_username>` first.\n\nOr specify a player: `"+support.Config.Prefix+"sc PlayerName."+command+"`")
		return
	}

	// Convert command to Lua code
	luaCode := convertCommandToLua(command, targetPlayer, isVerified || usesPlayerPrefix)

	// Use /silent-command to run Lua code without logging
	fullCommand := "/silent-command " + luaCode

	support.Factorio.Send(fullCommand)

	// Build response message
	var responseMsg string
	if usesPlayerPrefix {
		responseMsg = fmt.Sprintf("✅ Silent command executed on **%s**: `%s`", targetPlayer, command)
	} else if isVerified {
		responseMsg = fmt.Sprintf("✅ Silent command executed as **%s**: `%s`", targetPlayer, command)
	} else {
		responseMsg = fmt.Sprintf("✅ Silent command executed: `%s`", args)
	}
	responseMsg += "\n\n🔒 This command was **hidden** from the game chat and logs."

	sendDMToChannel(s, m.ChannelID, responseMsg)

	// Note: We intentionally don't log the silent command content to maintain privacy
	log.Printf("[DM] Silent command from %s (content hidden for privacy)", m.Author.Username)
}

// needsPlayerContext checks if a command needs a player context
func needsPlayerContext(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	cmd := strings.ToLower(parts[0])
	playerCommands := []string{
		"editor", "cheat", "cheatmode", "spectator", "god", "godmode",
		"zoom", "teleport", "tp", "give", "insert", "clear_inventory", "clear",
		"speed_modifier", "reach", "long_reach", "mining_speed", "crafting_speed",
		"research_all", "chart_all", "reveal_all", "map",
	}
	for _, pc := range playerCommands {
		if cmd == pc {
			return true
		}
	}
	return false
}

// convertCommandToLua converts a shorthand command to Lua code
func convertCommandToLua(command string, targetPlayer string, hasPlayer bool) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return command
	}

	cmd := strings.ToLower(parts[0])
	cmdArgs := parts[1:]

	// Helper to get player Lua code
	playerLua := func() string {
		if hasPlayer && targetPlayer != "" {
			return fmt.Sprintf(`game.get_player("%s")`, escapeFactorioString(targetPlayer))
		}
		return "game.player"
	}

	switch cmd {
	// Player-specific commands
	case "editor":
		return fmt.Sprintf(`local p = %s; if p then p.toggle_map_editor() end`, playerLua())

	case "spectator":
		return fmt.Sprintf(`local p = %s; if p then p.spectator = not p.spectator; p.print("Spectator: " .. tostring(p.spectator)) end`, playerLua())

	case "cheat", "cheatmode", "cheat_mode":
		return fmt.Sprintf(`local p = %s; if p then p.cheat_mode = not p.cheat_mode; p.print("Cheat mode: " .. tostring(p.cheat_mode)) end`, playerLua())

	case "god", "godmode":
		return fmt.Sprintf(`local p = %s; if p then if p.character then p.character.destroy(); p.print("God mode enabled") else p.create_character(); p.print("God mode disabled") end end`, playerLua())

	case "zoom":
		if len(cmdArgs) > 0 {
			return fmt.Sprintf(`local p = %s; if p then p.zoom = %s end`, playerLua(), cmdArgs[0])
		}
		return fmt.Sprintf(`local p = %s; if p then p.print("Current zoom: " .. p.zoom) end`, playerLua())

	case "teleport", "tp":
		if len(cmdArgs) >= 2 {
			return fmt.Sprintf(`local p = %s; if p then p.teleport({%s, %s}) end`, playerLua(), cmdArgs[0], cmdArgs[1])
		} else if len(cmdArgs) == 1 {
			// Teleport to another player
			return fmt.Sprintf(`local p = %s; local t = game.get_player("%s"); if p and t then p.teleport(t.position, t.surface) end`, playerLua(), escapeFactorioString(cmdArgs[0]))
		}
		return command

	case "give", "insert":
		if len(cmdArgs) > 0 {
			itemName := cmdArgs[0]
			count := "1"
			if len(cmdArgs) > 1 {
				count = cmdArgs[1]
			}
			return fmt.Sprintf(`local p = %s; if p then p.insert{name="%s", count=%s} end`, playerLua(), escapeFactorioString(itemName), count)
		}
		return command

	case "clear_inventory", "clear":
		return fmt.Sprintf(`local p = %s; if p then p.clear_items_inside() end`, playerLua())

	case "reach", "long_reach":
		reach := "1000"
		if len(cmdArgs) > 0 {
			reach = cmdArgs[0]
		}
		return fmt.Sprintf(`local p = %s; if p then p.character_build_distance_bonus = %s; p.character_reach_distance_bonus = %s end`, playerLua(), reach, reach)

	case "mining_speed":
		speed := "1000"
		if len(cmdArgs) > 0 {
			speed = cmdArgs[0]
		}
		return fmt.Sprintf(`local p = %s; if p then p.force.manual_mining_speed_modifier = %s end`, playerLua(), speed)

	case "crafting_speed":
		speed := "1000"
		if len(cmdArgs) > 0 {
			speed = cmdArgs[0]
		}
		return fmt.Sprintf(`local p = %s; if p then p.force.manual_crafting_speed_modifier = %s end`, playerLua(), speed)

	// Game-wide commands (no player needed)
	case "speed", "game_speed":
		if len(cmdArgs) > 0 {
			return fmt.Sprintf(`game.speed = %s`, cmdArgs[0])
		}
		return `game.print("Current speed: " .. game.speed)`

	case "research_all":
		return fmt.Sprintf(`local p = %s; if p then p.force.research_all_technologies() end`, playerLua())

	case "chart_all", "reveal_all", "map":
		return fmt.Sprintf(`local p = %s; if p then p.force.chart_all() end`, playerLua())

	case "always_day", "day":
		return fmt.Sprintf(`local p = %s; if p then p.surface.always_day = true end`, playerLua())

	case "night":
		return fmt.Sprintf(`local p = %s; if p then p.surface.always_day = false end`, playerLua())

	case "freeze_time":
		return fmt.Sprintf(`local p = %s; if p then p.surface.freeze_daytime = not p.surface.freeze_daytime end`, playerLua())

	case "peaceful":
		return fmt.Sprintf(`local p = %s; if p then p.surface.peaceful_mode = not p.surface.peaceful_mode; p.print("Peaceful mode: " .. tostring(p.surface.peaceful_mode)) end`, playerLua())

	case "evolution":
		if len(cmdArgs) > 0 {
			return fmt.Sprintf(`local p = %s; if p then game.forces["enemy"].set_evolution_factor(%s, p.surface) end`, playerLua(), cmdArgs[0])
		}
		return fmt.Sprintf(`local p = %s; if p then p.print("Evolution: " .. game.forces["enemy"].get_evolution_factor(p.surface)) end`, playerLua())

	case "kill_enemies", "kill_biters":
		return fmt.Sprintf(`local p = %s; if p then for _, e in pairs(p.surface.find_entities_filtered{force="enemy"}) do e.destroy() end end`, playerLua())

	case "pollution", "clear_pollution":
		return fmt.Sprintf(`local p = %s; if p then p.surface.clear_pollution() end`, playerLua())

	case "time":
		return `game.print("Map age: " .. game.tick .. " ticks (" .. math.floor(game.tick/60/60/60) .. " hours)")`

	case "seed":
		return fmt.Sprintf(`local p = %s; if p then p.print("Seed: " .. p.surface.map_gen_settings.seed) end`, playerLua())

	case "version":
		return `game.print("Factorio version: " .. game.active_mods["base"])`

	default:
		// Return as raw Lua code
		return command
	}
}

// runDMRegularCommand runs a regular bot command from DM context
func runDMRegularCommand(s *discordgo.Session, m *discordgo.MessageCreate, commandName, args string) {
	// Check if it's a help command
	if commandName == strings.ToLower("help") {
		sendDMHelp(s, m.ChannelID)
		return
	}

	// Find and execute the command
	for _, command := range commands.Commands {
		if strings.ToLower(command.Name) == commandName {
			// Check admin permission
			if command.Admin != nil && command.Admin(args) {
				if !commands.CheckAdmin(m.Author.ID) {
					sendDMToChannel(s, m.ChannelID, "❌ You don't have permission for this command.")
					return
				}
			}

			// Create a custom session handler that redirects output to DM
			// For now, just notify that the command was executed
			sendDMToChannel(s, m.ChannelID, fmt.Sprintf("⚙️ Executing command: `%s %s`\n\n_Note: Command output will appear in the main Factorio channel._", commandName, args))
			command.Command(s, args)
			return
		}
	}

	sendDMToChannel(s, m.ChannelID, "❌ Unknown command: `"+commandName+"`. Use `"+support.Config.Prefix+"help` for available commands.")
}

// sendDMHelp sends the DM help message
func sendDMHelp(s *discordgo.Session, channelID string) {
	embed := &discordgo.MessageEmbed{
		Type:        "rich",
		Color:       0x7289DA,
		Title:       "🤖 FactoCord DM Commands",
		Description: "Control the Factorio server via Direct Messages.\n\n**Prefix:** `" + support.Config.Prefix + "`",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "🔐 Verification",
				Value: fmt.Sprintf("`%sverify <player>` - Link with Factorio player\n`%sconfirm <code>` - Confirm code\n`%sunlink` - Remove link\n`%sstatus` - Show status",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name: "🎮 Console Commands",
				Value: fmt.Sprintf("`%sc <cmd>` - Run command (**visible**)\n`%ssc <cmd>` - Silent command (**hidden**)\n`%ssc Player.cmd` - Run on specific player",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name:  "⚡ Quick Silent Commands",
				Value: "`editor` `cheat` `spectator` `god`\n`zoom <n>` `speed <n>` `teleport <x> <y>`\n`give <item> [count]` `research_all`\n`chart_all` `kill_enemies` `peaceful`",
			},
			{
				Name: "🕵️ Spy Mode",
				Value: fmt.Sprintf("`%sspy on` - Hide ALL your actions from logs\n`%sspy off` - Show actions again\n`%sspy add <TYPE>` - Hide specific log type\n`%sspy list` - Show available log types",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name: "👻 Ghost Mode",
				Value: fmt.Sprintf("`%sghost on` - Become completely invisible\n`%sghost off` - Become visible again\n`%sghost prelogin` - Activate on next join\n`%sghost commands` - Auto-silent commands",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name:  "📋 Bot Commands",
				Value: "All regular bot commands work here too:\n`server`, `save`, `kick`, `ban`, `unban`, `config`, `mod`, `mods`, `version`, `info`, `online`",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "👻 Ghost Mode: Full invisibility with fake login/logout",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Failed to send DM help: %v", err)
	}
}

// sendDMToChannel sends a message to a specific channel (usually DM)
func sendDMToChannel(s *discordgo.Session, channelID, message string) {
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Failed to send DM: %v", err)
	}
}

// escapeFactorioString escapes a string for use in Factorio Lua commands
func escapeFactorioString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// NotifyDiscordUserOfVerification notifies a Discord user when someone verifies
func NotifyDiscordUserOfVerification(s *discordgo.Session, discordUserID, factorioUsername string) {
	// This can be called from the Factorio side when verification is complete
	channel, err := s.UserChannelCreate(discordUserID)
	if err != nil {
		log.Printf("Failed to create DM channel for notification: %v", err)
		return
	}

	sendDMToChannel(s, channel.ID, fmt.Sprintf("✅ Verification notification: Player **%s** has been linked to your account.", factorioUsername))
}
